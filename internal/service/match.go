package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/llm"
	"suprie/application_tracker/internal/matcher"
	"suprie/application_tracker/internal/rag"
	"suprie/application_tracker/internal/repository"
)

// RunMatch performs an AI-powered semantic fit match between the master profile
// and the job description identified by jdID. If chunked profile data is available
// and fresh, only the chunks most relevant to the JD (via BM25) are used in the
// prompt. Otherwise, falls back to the full master profile.
func RunMatch(masterProfilePath string, jdID int, repo repository.JobDescriptionRepository) {
	jd, err := repo.GetByID(context.Background(), jdID)
	if err != nil {
		log.Fatalf("loading job description id=%d: %v", jdID, err)
	}

	profileBytes, err := os.ReadFile(masterProfilePath)
	if err != nil {
		log.Fatalf("reading master profile %s: %v", masterProfilePath, err)
	}

	profileStr := string(profileBytes)
	prompt := buildMatchPrompt(masterProfilePath, profileStr, jd)
	schema := matcher.BuildMatchJSONSchema()

	lmClient := llm.NewLMStudioClient("match")
	response, err := lmClient.Generate(context.Background(), prompt, &schema)
	if err != nil {
		log.Fatalf("generating match analysis: %v", err)
	}

	var match matcher.MatchResponse
	if err := json.Unmarshal([]byte(response), &match); err != nil {
		log.Fatalf("unmarshal match response: %v", err)
	}

	if err := repo.UpdateFitScore(context.Background(), jdID, match.FitScore, match.Summary); err != nil {
		log.Fatalf("saving fit score: %v", err)
	}

	fmt.Printf("\n=== Match Result for JD #%d ===\n", jdID)
	if jd.Company != nil {
		fmt.Printf("Company: %s\n", *jd.Company)
	}
	if jd.RoleTitle != nil {
		fmt.Printf("Role:    %s\n", *jd.RoleTitle)
	}
	fmt.Printf("Fit Score: %d/100\n", match.FitScore)
	fmt.Printf("Verdict:   %s\n", match.GoNoGo)

	if len(match.Strengths) > 0 {
		fmt.Println("\nStrengths:")
		for _, s := range match.Strengths {
			fmt.Printf("  ✅ %s\n", s)
		}
	}

	if len(match.Gaps) > 0 {
		fmt.Println("\nGaps:")
		for _, g := range match.Gaps {
			fmt.Printf("  ⚠️  %s\n", g)
		}
	}

	fmt.Printf("\nSummary: %s\n", match.Summary)
}

// buildMatchPrompt attempts BM25-based prompt construction. Falls back to the
// full-profile prompt when chunk data is unavailable, stale, or retrieval fails.
func buildMatchPrompt(profilePath, fullProfile string, jd *domain.JobDescription) string {
	store := rag.NewChunkStore()
	if err := store.Load(rag.ChunksFileName); err != nil {
		return matcher.BuildMatchPrompt(fullProfile, jd)
	}
	if store.IsStale(profilePath) || store.Len() == 0 {
		return matcher.BuildMatchPrompt(fullProfile, jd)
	}

	query := buildMatchQuery(jd)
	chunks, err := rag.Retrieve(query, store, 5)
	if err != nil || len(chunks) == 0 {
		return matcher.BuildMatchPrompt(fullProfile, jd)
	}

	fmt.Printf("🔍 Using %d retrieved profile chunks for matching\n", len(chunks))
	return matcher.BuildMatchPromptWithChunks(chunks, jd)
}

// buildMatchQuery constructs a retrieval query from the JD fields most
// useful for finding relevant profile experience.
func buildMatchQuery(jd *domain.JobDescription) string {
	var parts []string
	if jd.RoleTitle != nil {
		parts = append(parts, *jd.RoleTitle)
	}
	parts = append(parts, jd.RequirementsJSON)
	parts = append(parts, jd.ResponsibilitiesJSON)
	parts = append(parts, jd.KeywordsJSON)
	return strings.Join(parts, " ")
}
