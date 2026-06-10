package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/repository"
)

// RunList displays all job descriptions, optionally filtered by status.
func RunList(statusFilter string, repo repository.JobDescriptionRepository) {
	// Normalize the filter: "draft" → "Draft", "" → "".
	var normalized string
	switch strings.ToLower(statusFilter) {
	case "draft":
		normalized = domain.StatusDraft
	case "fit match", "fitmatch", "fit_match":
		normalized = domain.StatusFitMatch
	case "applied":
		normalized = domain.StatusApplied
	case "rejected":
		normalized = domain.StatusRejected
	case "offer":
		normalized = domain.StatusOffer
	case "":
		normalized = ""
	default:
		log.Fatalf("unknown status filter: %q (use: draft, fitmatch, applied, rejected, offer)", statusFilter)
	}

	jds, err := repo.List(context.Background(), normalized)
	if err != nil {
		log.Fatalf("listing job descriptions: %v", err)
	}

	if len(jds) == 0 {
		if normalized != "" {
			fmt.Printf("No job descriptions with status %q.\n", normalized)
		} else {
			fmt.Println("No job descriptions yet. Use parse-jd to add one.")
		}
		return
	}

	fmt.Printf("\n%-4s %-25s %-25s %-10s %-5s\n", "ID", "Company", "Role", "Status", "Fit")
	fmt.Println(strings.Repeat("-", 75))
	for _, jd := range jds {
		company := "-"
		if jd.Company != nil {
			company = *jd.Company
		}
		role := "-"
		if jd.RoleTitle != nil {
			role = *jd.RoleTitle
		}
		fit := "-"
		if jd.FitScore != nil {
			fit = fmt.Sprintf("%d%%", *jd.FitScore)
		}
		fmt.Printf("%-4d %-25s %-25s %-10s %-5s\n",
			jd.ID,
			truncate(company, 25),
			truncate(role, 25),
			jd.Status,
			fit,
		)
	}
	fmt.Println()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-2] + ".."
}
