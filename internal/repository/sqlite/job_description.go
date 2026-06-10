package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/repository"
)

// Compile-time check that JobDescriptionRepository implements the interface.
var _ repository.JobDescriptionRepository = (*JobDescriptionRepository)(nil)

type JobDescriptionRepository struct {
	db *sql.DB
}

func NewJobDescriptionRepository(db *sql.DB) *JobDescriptionRepository {
	return &JobDescriptionRepository{db: db}
}

const selectColumns = `id, company, role_title, seniority, employment_type,
	work_arrangement, location, requirements_json,
	responsibilities_json, keywords_json, parsing_warning_json,
	apply_url, status, fit_score, fit_summary, applied_at, created_at`

const insertColumns = `company, role_title, seniority, employment_type,
	work_arrangement, location, requirements_json,
	responsibilities_json, keywords_json, parsing_warning_json,
	apply_url, status, fit_score, fit_summary, applied_at, created_at`

// --- Read ---

func (r *JobDescriptionRepository) GetByID(ctx context.Context, id int) (*domain.JobDescription, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT "+selectColumns+" FROM job_descriptions WHERE id = ?", id)

	jd, err := scanJobDescription(row)
	if err != nil {
		return nil, fmt.Errorf("get job_description id=%d: %w", id, err)
	}
	return jd, nil
}

// --- Create ---

func (r *JobDescriptionRepository) Create(ctx context.Context, jd *domain.JobDescription) error {
	if jd.Status == "" {
		jd.Status = domain.StatusDraft
	}

	result, err := r.db.ExecContext(ctx,
		"INSERT INTO job_descriptions ("+insertColumns+") VALUES ("+
			"?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		jd.Company,
		jd.RoleTitle,
		jd.Seniority,
		jd.EmploymentType,
		jd.WorkArrangement,
		jd.Location,
		jd.RequirementsJSON,
		jd.ResponsibilitiesJSON,
		jd.KeywordsJSON,
		jd.ParsingWarningJSON,
		jd.ApplyURL,
		jd.Status,
		jd.FitScore,
		jd.FitSummary,
		jd.AppliedAt,
		jd.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert job_description: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	jd.ID = int(id)
	return nil
}

// --- List ---

func (r *JobDescriptionRepository) GetAll(ctx context.Context) ([]domain.JobDescription, error) {
	return r.List(ctx, "")
}

func (r *JobDescriptionRepository) List(ctx context.Context, status string) ([]domain.JobDescription, error) {
	query := "SELECT " + selectColumns + " FROM job_descriptions"
	args := []any{}

	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query job_descriptions: %w", err)
	}
	defer rows.Close()

	var results []domain.JobDescription
	for rows.Next() {
		jd, err := scanJobDescription(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *jd)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	// Return empty slice, not nil, for JSON serialization friendliness.
	if results == nil {
		results = []domain.JobDescription{}
	}
	return results, nil
}

// --- Update ---

func (r *JobDescriptionRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	var appliedAt *time.Time
	if status == domain.StatusApplied {
		now := time.Now()
		appliedAt = &now
	}

	_, err := r.db.ExecContext(ctx,
		"UPDATE job_descriptions SET status = ?, applied_at = ? WHERE id = ?",
		status, appliedAt, id,
	)
	if err != nil {
		return fmt.Errorf("update status id=%d to %q: %w", id, status, err)
	}
	return nil
}

func (r *JobDescriptionRepository) UpdateFitScore(ctx context.Context, id int, score int, summary string) error {
	status := domain.StatusFitMatch
	_, err := r.db.ExecContext(ctx,
		"UPDATE job_descriptions SET status = ?, fit_score = ?, fit_summary = ? WHERE id = ?",
		status, score, summary, id,
	)
	if err != nil {
		return fmt.Errorf("update fit_score id=%d: %w", id, err)
	}
	return nil
}

// --- scan ---

// scanner abstracts *sql.Row and *sql.Rows for a shared scan function.
type scanner interface {
	Scan(dest ...any) error
}

func scanJobDescription(s scanner) (*domain.JobDescription, error) {
	var jd domain.JobDescription
	err := s.Scan(
		&jd.ID,
		&jd.Company,
		&jd.RoleTitle,
		&jd.Seniority,
		&jd.EmploymentType,
		&jd.WorkArrangement,
		&jd.Location,
		&jd.RequirementsJSON,
		&jd.ResponsibilitiesJSON,
		&jd.KeywordsJSON,
		&jd.ParsingWarningJSON,
		&jd.ApplyURL,
		&jd.Status,
		&jd.FitScore,
		&jd.FitSummary,
		&jd.AppliedAt,
		&jd.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &jd, nil
}
