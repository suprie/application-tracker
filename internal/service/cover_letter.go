package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"suprie/application_tracker/internal/coverletter"
	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/llm"
	"suprie/application_tracker/internal/rag"
	"suprie/application_tracker/internal/ranker"
	"suprie/application_tracker/internal/repository"
)

// RunCoverLetter generates a tailored LaTeX cover letter for the given JD,
// compiles it to PDF, and keeps both files.
//
// When chunked profile data is available and fresh, a two-stage pipeline runs:
//  1. Experience Ranker — re-ranks BM25 chunks by strategic fit
//  2. Cover Letter Writer — uses the ranker's curated selections
//
// Otherwise, falls back to the full-profile prompt in a single LLM call.
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

	// Build prompt — with ranker if RAG chunks are available.
	prompt, rr := buildCoverLetterPromptWithRanker(masterProfilePath, profileStr, jd)

	// Persist ranker result for reuse.
	if rr != nil && rr.JSON != "" {
		if err := repo.UpdateRankerResult(context.Background(), jdID, rr.JSON); err != nil {
			fmt.Printf("⚠️  Failed to persist ranker result: %v\n", err)
		}
		// Give Ollama time to release GPU memory before the next request.
		fmt.Println("⏳ Waiting for Ollama to release resources (5s)...")
		time.Sleep(5 * time.Second)
	}

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

// rankerResult holds the output of the experience ranker for persistence.
type rankerResult struct {
	Prompt string          // cover letter prompt built from ranked output
	JSON   string          // raw ranker JSON for storage
	Ranked *ranker.Response // parsed ranker response
}

// buildCoverLetterPromptWithRanker attempts the two-stage RAG pipeline:
// BM25 retrieval → Experience Ranker → cover letter prompt from ranked output.
// Falls back to the full-profile prompt when RAG is unavailable or fails.
func buildCoverLetterPromptWithRanker(profilePath, fullProfile string, jd *domain.JobDescription) (string, *rankerResult) {
	store := rag.NewChunkStore()
	if err := store.Load(rag.ChunksFileName); err != nil {
		return coverletter.BuildCoverLetterPrompt(fullProfile, jd), nil
	}
	if store.IsStale(profilePath) || store.Len() == 0 {
		return coverletter.BuildCoverLetterPrompt(fullProfile, jd), nil
	}

	query := buildCoverLetterQuery(jd)
	chunks, err := rag.Retrieve(query, store, 8) // retrieve more for the ranker to choose from
	if err != nil || len(chunks) == 0 {
		return coverletter.BuildCoverLetterPrompt(fullProfile, jd), nil
	}

	fmt.Printf("🔍 BM25 retrieved %d chunks — running experience ranker\n", len(chunks))

	// Stage 1: Experience Ranker.
	summary := ranker.ProfileSummary(store.Chunks())
	rankerPrompt := ranker.BuildPrompt(summary, chunks, jd)
	rankerSchema := ranker.BuildJSONSchema()

	rankerClient := llm.NewLMStudioClient("ranker")
	rankerResp, err := rankerClient.Generate(context.Background(), rankerPrompt, &rankerSchema)
	if err != nil {
		fmt.Printf("⚠️  Ranker failed: %v — falling back to full profile\n", err)
		return coverletter.BuildCoverLetterPrompt(fullProfile, jd), nil
	}

	var ranked ranker.Response
	if err := json.Unmarshal([]byte(rankerResp), &ranked); err != nil {
		fmt.Printf("⚠️  Ranker response parse failed: %v — falling back to full profile\n", err)
		return coverletter.BuildCoverLetterPrompt(fullProfile, jd), nil
	}

	fmt.Printf("✅ Ranker selected %d experiences, %d skills\n",
		len(ranked.SelectedExperiences), len(ranked.SelectedSkills))

	// Stage 2: Build cover letter prompt from ranked output.
	return coverletter.BuildCoverLetterPromptWithRanked(&ranked, jd), &rankerResult{
		Prompt: "", // caller uses the returned prompt string
		JSON:   rankerResp,
		Ranked: &ranked,
	}
}

// buildCoverLetterQuery constructs a retrieval query from JD fields.
func buildCoverLetterQuery(jd *domain.JobDescription) string {
	var parts []string
	if jd.RoleTitle != nil {
		parts = append(parts, *jd.RoleTitle)
	}
	parts = append(parts, jd.RequirementsJSON)
	parts = append(parts, jd.ResponsibilitiesJSON)
	return strings.Join(parts, " ")
}
