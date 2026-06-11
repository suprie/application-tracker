package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"suprie/application_tracker/internal/coverletter"
	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/llm"
	"suprie/application_tracker/internal/ranker"
	"suprie/application_tracker/internal/repository"
)

// RunCoverLetter generates a tailored LaTeX cover letter for the given JD,
// compiles it to PDF, and keeps both files.
//
// If a cached ranker result exists for this JD (from `ats rank <id>`), it is
// used to build a focused prompt. Otherwise, falls back to the full profile.
func RunCoverLetter(jdID int, masterProfilePath string, repo repository.JobDescriptionRepository) {
	jd, err := repo.GetByID(context.Background(), jdID)
	if err != nil {
		log.Fatalf("loading job description id=%d: %v", jdID, err)
	}

	profileBytes, err := os.ReadFile(masterProfilePath)
	if err != nil {
		log.Fatalf("reading master profile %s: %v", masterProfilePath, err)
	}

	profileStr := string(profileBytes)
	prompt := buildCoverLetterPrompt(profileStr, jd)

	schema := coverletter.BuildJSONSchema()

	lmClient := llm.NewLMStudioClient("cover_letter")
	response, err := lmClient.Generate(context.Background(), prompt, &schema)
	if err != nil {
		log.Fatalf("generating cover letter: %v", err)
	}

	var cl coverletter.Response
	if err = json.Unmarshal([]byte(response), &cl); err != nil {
		log.Fatalf("unmarshal cover letter response: %v", err)
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
		log.Fatalf("rendering LaTeX: %v", err)
	}

	// Ensure output directory exists.
	_ = os.MkdirAll("generated", 0755)

	base := fmt.Sprintf("generated/cover_letter_%d", jdID)
	texPath := base + ".tex"

	if err := os.WriteFile(texPath, []byte(tex), 0644); err != nil {
		log.Fatalf("writing tex file: %v", err)
	}

	fmt.Printf("LaTeX written to %s\n", texPath)

	// Compile to PDF.
	latexCmd := os.Getenv("LATEX_CMD")
	if latexCmd == "" {
		latexCmd = "pdflatex"
	}

	texAbsPath, _ := filepath.Abs(texPath)
	outDir := filepath.Dir(texAbsPath)

	cmd := exec.Command(latexCmd,
		"-interaction=nonstopmode",
		"-output-directory="+outDir,
		texAbsPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  LaTeX compilation failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "   The .tex file is at %s — compile manually.\n", texPath)
		return
	}

	fmt.Printf("PDF compiled to %s\n", base+".pdf")
	fmt.Println("\nCover letter generated successfully.")
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
