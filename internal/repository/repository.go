package repository

import(
	"context"

	"suprie/application_tracker/internal/domain"
)

type JobDescriptionRepository interface {
	GetByID(ctx context.Context, id int) (*domain.JobDescription, error)
	Create(ctx context.Context, jobDescription *domain.JobDescription) error
	GetAll(ctx context.Context) ([]domain.JobDescription, error)
}
