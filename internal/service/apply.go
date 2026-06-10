package service

import (
	"context"
	"fmt"
	"log"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/repository"
)

// RunApply marks the job description as applied.
func RunApply(jdID int, repo repository.JobDescriptionRepository) {
	jd, err := repo.GetByID(context.Background(), jdID)
	if err != nil {
		log.Fatalf("loading job description id=%d: %v", jdID, err)
	}

	if jd.Status == domain.StatusApplied {
		fmt.Printf("Job #%d is already marked as Applied.\n", jdID)
		return
	}

	if err := repo.UpdateStatus(context.Background(), jdID, domain.StatusApplied); err != nil {
		log.Fatalf("updating status: %v", err)
	}

	fmt.Printf("✅ Job #%d — ", jdID)
	if jd.Company != nil {
		fmt.Printf("%s — ", *jd.Company)
	}
	if jd.RoleTitle != nil {
		fmt.Printf("%s", *jd.RoleTitle)
	}
	fmt.Printf(" — status changed to Applied.\n")

	if jd.FitScore != nil {
		fmt.Printf("   Fit score at time of application: %d/100\n", *jd.FitScore)
	}
}
