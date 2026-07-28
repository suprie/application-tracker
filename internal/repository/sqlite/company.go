package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/repository"
)

// Compile-time check that CompanyRepository implements the interface.
var _ repository.CompanyRepository = (*CompanyRepository)(nil)

type CompanyRepository struct {
	db *sql.DB
}

func NewCompanyRepository(db *sql.DB) *CompanyRepository {
	return &CompanyRepository{db: db}
}

const companySelectColumns = `id, name, normalized_name, website_url, industry,
	size, country, notes, source, research_summary, created_at, updated_at`

const companyInsertColumns = `name, normalized_name, website_url, industry,
	size, country, notes, source, research_summary, created_at, updated_at`

// --- Read ---

func (r *CompanyRepository) GetByID(ctx context.Context, id int) (*domain.Company, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT "+companySelectColumns+" FROM companies WHERE id = ?", id)
	c, err := scanCompany(row)
	if err != nil {
		return nil, fmt.Errorf("get company id=%d: %w", id, err)
	}
	return c, nil
}

func (r *CompanyRepository) GetByNormalizedName(ctx context.Context, normalized string) (*domain.Company, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT "+companySelectColumns+" FROM companies WHERE normalized_name = ?", normalized)
	c, err := scanCompany(row)
	if err != nil {
		return nil, fmt.Errorf("get company normalized_name=%q: %w", normalized, err)
	}
	return c, nil
}

// --- Create ---

func (r *CompanyRepository) Create(ctx context.Context, c *domain.Company) error {
	now := time.Now()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = now
	}

	result, err := r.db.ExecContext(ctx,
		"INSERT INTO companies ("+companyInsertColumns+") VALUES ("+
			"?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		c.Name,
		c.NormalizedName,
		c.WebsiteURL,
		c.Industry,
		c.Size,
		c.Country,
		c.Notes,
		c.Source,
		c.ResearchSummary,
		c.CreatedAt,
		c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert company: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	c.ID = int(id)
	return nil
}

// --- List ---

func (r *CompanyRepository) List(ctx context.Context) ([]domain.Company, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT "+companySelectColumns+" FROM companies ORDER BY name ASC")
	if err != nil {
		return nil, fmt.Errorf("query companies: %w", err)
	}
	defer rows.Close()

	var results []domain.Company
	for rows.Next() {
		c, err := scanCompany(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	// Empty slice, not nil, for JSON friendliness.
	if results == nil {
		results = []domain.Company{}
	}
	return results, nil
}

// --- Update ---

func (r *CompanyRepository) Update(ctx context.Context, c *domain.Company) error {
	c.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx,
		"UPDATE companies SET name = ?, normalized_name = ?, website_url = ?, "+
			"industry = ?, size = ?, country = ?, notes = ?, source = ?, "+
			"research_summary = ?, updated_at = ? WHERE id = ?",
		c.Name,
		c.NormalizedName,
		c.WebsiteURL,
		c.Industry,
		c.Size,
		c.Country,
		c.Notes,
		c.Source,
		c.ResearchSummary,
		c.UpdatedAt,
		c.ID,
	)
	if err != nil {
		return fmt.Errorf("update company id=%d: %w", c.ID, err)
	}
	return nil
}

// --- scan ---

func scanCompany(s scanner) (*domain.Company, error) {
	var c domain.Company
	err := s.Scan(
		&c.ID,
		&c.Name,
		&c.NormalizedName,
		&c.WebsiteURL,
		&c.Industry,
		&c.Size,
		&c.Country,
		&c.Notes,
		&c.Source,
		&c.ResearchSummary,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
