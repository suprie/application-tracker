package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"suprie/application_tracker/internal/llm"
	"suprie/application_tracker/internal/matcher"
	"suprie/application_tracker/internal/repository"
)

// RunMatch performs an AI-powered semantic fit match between the master profile
// and the job description identified by jdID. Results are persisted via repo.
func RunMatch(masterProfilePath string, jdID int, repo repository.JobDescriptionRepository) {
	jd, err := repo.GetByID(context.Background(), jdID)
	if err != nil {
		log.Fatalf("loading job description id=%d: %v", jdID, err)
	}

	profileBytes, err := os.ReadFile(masterProfilePath)
	if err != nil {
		log.Fatalf("reading master profile %s: %v", masterProfilePath, err)
	}

	prompt := matcher.BuildMatchPrompt(string(profileBytes), jd)
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
