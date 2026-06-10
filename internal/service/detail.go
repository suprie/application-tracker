package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"suprie/application_tracker/internal/repository"
)

// RunDetail displays the full job description with match and application info.
func RunDetail(jdID int, repo repository.JobDescriptionRepository) {
	jd, err := repo.GetByID(context.Background(), jdID)
	if err != nil {
		log.Fatalf("loading job description id=%d: %v", jdID, err)
	}

	fmt.Printf("\n=== Job Description #%d ===\n\n", jd.ID)
	printField("Company", jd.Company)
	printField("Role", jd.RoleTitle)
	printField("Seniority", jd.Seniority)
	printField("Employment", jd.EmploymentType)
	printField("Work Mode", jd.WorkArrangement)
	printField("Location", jd.Location)
	printField("Apply URL", jd.ApplyURL)

	fmt.Println()
	fmt.Printf("Status: %s\n", jd.Status)

	if jd.FitScore != nil {
		fmt.Printf("Fit Score: %d/100\n", *jd.FitScore)
	}
	if jd.FitSummary != nil {
		fmt.Printf("Fit Summary: %s\n", *jd.FitSummary)
	}
	if jd.AppliedAt != nil {
		fmt.Printf("Applied At: %s\n", jd.AppliedAt.Format("2006-01-02 15:04"))
	}

	fmt.Println("\n--- Requirements ---")
	prettyPrintJSON(jd.RequirementsJSON)

	fmt.Println("\n--- Responsibilities ---")
	prettyPrintJSON(jd.ResponsibilitiesJSON)

	fmt.Println("\n--- Keywords ---")
	prettyPrintJSON(jd.KeywordsJSON)

	if jd.ParsingWarningJSON != "" && jd.ParsingWarningJSON != "[]" && jd.ParsingWarningJSON != "null" {
		fmt.Println("\n⚠️  Parsing Warnings:")
		prettyPrintJSON(jd.ParsingWarningJSON)
	}

	fmt.Println()
}

func printField(label string, value *string) {
	if value != nil {
		fmt.Printf("%-15s %s\n", label+":", *value)
	}
}

func prettyPrintJSON(raw string) {
	if raw == "" || raw == "[]" || raw == "{}" {
		fmt.Println("  (none)")
		return
	}

	// Try to parse and re-marshal with indentation.
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		fmt.Printf("  %s\n", raw)
		return
	}

	pretty, err := json.MarshalIndent(v, "  ", "  ")
	if err != nil {
		fmt.Printf("  %s\n", raw)
		return
	}

	for _, line := range strings.Split(string(pretty), "\n") {
		fmt.Printf("  %s\n", line)
	}
}
