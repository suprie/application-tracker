package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/matcher"
	"suprie/application_tracker/internal/rag"
	"suprie/application_tracker/internal/repository"
)

// RunMatch is the CLI wrapper for Match.
func RunMatch(masterProfilePath string, jdID int, repo repository.JobDescriptionRepository) {
	d := NewDeps(repo, nil)
	d.ProfilePath = masterProfilePath

	match, err := Match(context.Background(), d, jdID)
	if err != nil {
		log.Fatalf("%v", err)
	}

	jd, err := repo.GetByID(context.Background(), jdID)
	if err != nil {
		log.Fatalf("reloading job description id=%d: %v", jdID, err)
	}
	printMatchResult(jd, match)
}

// Match runs the AI fit-match between the master profile and the JD identified
// by jdID, persists the fit score, and returns the match response. If chunked
// profile data is available and fresh, only the BM25-retrieved chunks are used
// in the prompt; otherwise it falls back to the full profile.
func Match(ctx context.Context, d Deps, jdID int) (*matcher.MatchResponse, error) {
	jd, err := d.JDRepo.GetByID(ctx, jdID)
	if err != nil {
		return nil, fmt.Errorf("loading job description id=%d: %w", jdID, err)
	}

	profileBytes, err := os.ReadFile(d.ProfilePath)
	if err != nil {
		return nil, fmt.Errorf("reading master profile %s: %w", d.ProfilePath, err)
	}

	prompt := buildMatchPrompt(d.ProfilePath, string(profileBytes), jd)
	schema := matcher.BuildMatchJSONSchema()

	client := d.LLM("match")
	response, err := client.Generate(ctx, prompt, &schema)
	if err != nil {
		return nil, fmt.Errorf("generating match analysis: %w", err)
	}

	var match matcher.MatchResponse
	if err := json.Unmarshal([]byte(response), &match); err != nil {
		return nil, fmt.Errorf("unmarshal match response: %w", err)
	}

	if err := d.JDRepo.UpdateFitScore(ctx, jdID, match.FitScore, match.Summary); err != nil {
		return nil, fmt.Errorf("saving fit score: %w", err)
	}

	return &match, nil
}

// printMatchResult prints the match summary to stdout (CLI only).
func printMatchResult(jd *domain.JobDescription, match *matcher.MatchResponse) {
	fmt.Printf("\n=== Match Result for JD #%d ===\n", jd.ID)
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
