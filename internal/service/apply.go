package service

import (
	"context"
	"fmt"
	"log"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/repository"
)

// RunApply is the CLI wrapper for Apply.
func RunApply(jdID int, repo repository.JobDescriptionRepository) {
	d := NewDeps(repo, nil)
	jd, changed, err := Apply(context.Background(), d, jdID)
	if err != nil {
		log.Fatalf("%v", err)
	}

	if !changed {
		fmt.Printf("Job #%d is already marked as Applied.\n", jdID)
		return
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

// Apply marks the job description as Applied. It is idempotent: applying an
// already-applied JD is not an error. changed reports whether the status was
// actually modified.
func Apply(ctx context.Context, d Deps, jdID int) (*domain.JobDescription, bool, error) {
	jd, err := d.JDRepo.GetByID(ctx, jdID)
	if err != nil {
		return nil, false, fmt.Errorf("loading job description id=%d: %w", jdID, err)
	}

	if jd.Status == domain.StatusApplied {
		return jd, false, nil
	}

	if err := d.JDRepo.UpdateStatus(ctx, jdID, domain.StatusApplied); err != nil {
		return nil, false, fmt.Errorf("updating status: %w", err)
	}
	jd.Status = domain.StatusApplied
	return jd, true, nil
}
