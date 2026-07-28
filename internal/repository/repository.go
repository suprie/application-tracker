package repository

import (
	"context"

	"suprie/application_tracker/internal/domain"
)

type JobDescriptionRepository interface {
	GetByID(ctx context.Context, id int) (*domain.JobDescription, error)
	Create(ctx context.Context, jobDescription *domain.JobDescription) error
	GetAll(ctx context.Context) ([]domain.JobDescription, error)
	// List returns all JDs, optionally filtered by status (empty string = all).
	List(ctx context.Context, status string) ([]domain.JobDescription, error)
	// UpdateStatus sets the status and optionally the applied_at timestamp.
	UpdateStatus(ctx context.Context, id int, status string) error
	// UpdateFitScore sets the fit score and summary.
	UpdateFitScore(ctx context.Context, id int, score int, summary string) error
	// UpdateRankerResult stores the experience ranker JSON output for this JD.
	UpdateRankerResult(ctx context.Context, id int, rankerJSON string) error
	// UpdateApplyURL sets the apply URL (nil clears it).
	UpdateApplyURL(ctx context.Context, id int, applyURL *string) error
}

// CompanyRepository persists researched employer records.
type CompanyRepository interface {
	GetByID(ctx context.Context, id int) (*domain.Company, error)
	GetByNormalizedName(ctx context.Context, normalized string) (*domain.Company, error)
	Create(ctx context.Context, company *domain.Company) error
	List(ctx context.Context) ([]domain.Company, error)
	Update(ctx context.Context, company *domain.Company) error
}
