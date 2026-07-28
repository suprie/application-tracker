package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"suprie/application_tracker/internal/coverletter"
	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/ranker"
	"suprie/application_tracker/internal/repository"
)

// CoverLetterOutput is the result of CoverLetter. PDFPath is empty when the
// LaTeX compiler is unavailable or failed — the .tex source is still written.
type CoverLetterOutput struct {
	TeXPath string
	PDFPath string
}

// RunCoverLetter is the CLI wrapper for CoverLetter.
func RunCoverLetter(jdID int, masterProfilePath string, repo repository.JobDescriptionRepository) {
	d := NewDeps(repo, nil)
	d.ProfilePath = masterProfilePath

	out, err := CoverLetter(context.Background(), d, jdID)
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Printf("LaTeX written to %s\n", out.TeXPath)
	if out.PDFPath != "" {
		fmt.Printf("PDF compiled to %s\n", out.PDFPath)
		fmt.Println("\nCover letter generated successfully.")
	} else {
		fmt.Fprintf(os.Stderr, "   The .tex file is at %s — compile manually.\n", out.TeXPath)
	}
}

// CoverLetter generates a tailored LaTeX cover letter for the given JD and
// compiles it to PDF. If a cached ranker result exists for this JD (from
// Rank), it is used to build a focused prompt; otherwise the full profile is
// used. The .tex is always written; the PDF is best-effort.
func CoverLetter(ctx context.Context, d Deps, jdID int) (CoverLetterOutput, error) {
	jd, err := d.JDRepo.GetByID(ctx, jdID)
	if err != nil {
		return CoverLetterOutput{}, fmt.Errorf("loading job description id=%d: %w", jdID, err)
	}

	profileBytes, err := os.ReadFile(d.ProfilePath)
	if err != nil {
		return CoverLetterOutput{}, fmt.Errorf("reading master profile %s: %w", d.ProfilePath, err)
	}

	prompt := buildCoverLetterPrompt(string(profileBytes), jd)
	schema := coverletter.BuildJSONSchema()

	client := d.LLM("cover_letter")
	response, err := client.Generate(ctx, prompt, &schema)
	if err != nil {
		return CoverLetterOutput{}, fmt.Errorf("generating cover letter: %w", err)
	}

	var cl coverletter.Response
	if err = json.Unmarshal([]byte(response), &cl); err != nil {
		return CoverLetterOutput{}, fmt.Errorf("unmarshal cover letter response: %w", err)
	}

	// Render LaTeX.
	tex, err := coverletter.RenderLaTeX(coverletter.TemplateData{
		YourName:          cl.YourName,
		YourAddress:       cl.YourAddress,
		YourEmail:         cl.YourEmail,
		YourPhone:         cl.YourPhone,
		RecipientName:     cl.RecipientName,
		RecipientTitle:    cl.RecipientTitle,
		CompanyName:       cl.CompanyName,
		CompanyAddress:    cl.CompanyAddress,
		Subject:           cl.Subject,
		Opening:           cl.Opening,
		OpeningParagraphs: cl.OpeningParagraphs,
		BodyParagraphs:    cl.BodyParagraphs,
		ClosingParagraphs: cl.ClosingParagraphs,
		Closing:           cl.Closing,
	})
	if err != nil {
		return CoverLetterOutput{}, fmt.Errorf("rendering LaTeX: %w", err)
	}

	// Ensure output directory exists.
	if err := os.MkdirAll(d.GeneratedDir, 0755); err != nil {
		return CoverLetterOutput{}, fmt.Errorf("creating output dir: %w", err)
	}

	base := filepath.Join(d.GeneratedDir, fmt.Sprintf("cover_letter_%d", jdID))
	texPath := base + ".tex"

	if err := os.WriteFile(texPath, []byte(tex), 0644); err != nil {
		return CoverLetterOutput{}, fmt.Errorf("writing tex file: %w", err)
	}

	out := CoverLetterOutput{TeXPath: texPath, PDFPath: base + ".pdf"}

	// Compile to PDF (best-effort; pdflatex may be absent).
	latexCmd := os.Getenv("LATEX_CMD")
	if latexCmd == "" {
		latexCmd = "pdflatex"
	}

	texAbsPath, _ := filepath.Abs(texPath)
	cmd := exec.Command(latexCmd,
		"-interaction=nonstopmode",
		"-output-directory="+filepath.Dir(texAbsPath),
		texAbsPath,
	)
	var latexOut bytes.Buffer
	cmd.Stdout = &latexOut
	cmd.Stderr = &latexOut
	if err := cmd.Run(); err != nil {
		d.Logger.Printf("⚠️  LaTeX compilation failed: %v\n%s", err, latexOut.String())
		out.PDFPath = ""
	}

	return out, nil
}

// buildCoverLetterPrompt uses a cached ranker result if available, otherwise
// falls back to the full-profile prompt.
func buildCoverLetterPrompt(fullProfile string, jd *domain.JobDescription) string {
	// Use cached ranker result if available.
	if jd.RankerResultJSON != nil && *jd.RankerResultJSON != "" {
		var ranked ranker.Response
		if err := json.Unmarshal([]byte(*jd.RankerResultJSON), &ranked); err == nil &&
			len(ranked.SelectedExperiences) > 0 {
			fmt.Printf("✅ Using cached ranker result (%d experiences, %d skills)\n",
				len(ranked.SelectedExperiences), len(ranked.SelectedSkills))
			return coverletter.BuildCoverLetterPromptWithRanked(&ranked, jd)
		}
	}

	// Fall back to full profile.
	fmt.Println("⚠️  No cached ranker result — using full profile (run 'ats rank' first for better results)")
	return coverletter.BuildCoverLetterPrompt(fullProfile, jd)
}
