package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/dto"
	"suprie/application_tracker/internal/textutils"
)

// ErrCompanyExists signals a create/update would collide on normalized_name.
var ErrCompanyExists = errors.New("company with this normalized name already exists")

// CreateCompany validates the request, normalizes the name, and persists a
// new company.
func CreateCompany(ctx context.Context, d Deps, req dto.CreateCompanyRequest) (*domain.Company, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	normalized := textutils.NormalizeCompanyName(req.Name)
	if normalized == "" {
		return nil, errors.New("name resolves to an empty normalized form")
	}

	if existing, err := d.CompanyRepo.GetByNormalizedName(ctx, normalized); err == nil && existing != nil {
		return nil, fmt.Errorf("%w: %q", ErrCompanyExists, normalized)
	}

	c := &domain.Company{
		Name:            req.Name,
		NormalizedName:  normalized,
		WebsiteURL:      req.WebsiteURL,
		Industry:        req.Industry,
		Size:            req.Size,
		Country:         req.Country,
		Notes:           req.Notes,
		Source:          req.Source,
		ResearchSummary: req.ResearchSummary,
	}
	if err := d.CompanyRepo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("creating company: %w", err)
	}
	return c, nil
}

// ListCompanies returns all companies, optionally filtered by a normalized
// substring of q.
func ListCompanies(ctx context.Context, d Deps, q string) ([]domain.Company, error) {
	all, err := d.CompanyRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing companies: %w", err)
	}
	if q == "" {
		return all, nil
	}

	nq := textutils.NormalizeCompanyName(q)
	out := make([]domain.Company, 0, len(all))
	for _, c := range all {
		if strings.Contains(c.NormalizedName, nq) {
			out = append(out, c)
		}
	}
	return out, nil
}

// GetCompany fetches a single company by ID.
func GetCompany(ctx context.Context, d Deps, id int) (*domain.Company, error) {
	c, err := d.CompanyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("loading company id=%d: %w", id, err)
	}
	return c, nil
}

// UpdateCompany applies the non-nil fields of req to the stored company.
// Renaming re-derives the normalized name.
func UpdateCompany(ctx context.Context, d Deps, id int, req dto.UpdateCompanyRequest) (*domain.Company, error) {
	c, err := d.CompanyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("loading company id=%d: %w", id, err)
	}

	if req.Name != nil {
		newNorm := textutils.NormalizeCompanyName(*req.Name)
		if existing, err := d.CompanyRepo.GetByNormalizedName(ctx, newNorm); err == nil && existing != nil && existing.ID != id {
			return nil, fmt.Errorf("%w: %q", ErrCompanyExists, newNorm)
		}
		c.Name = *req.Name
		c.NormalizedName = newNorm
	}
	if req.WebsiteURL != nil {
		c.WebsiteURL = req.WebsiteURL
	}
	if req.Industry != nil {
		c.Industry = req.Industry
	}
	if req.Size != nil {
		c.Size = req.Size
	}
	if req.Country != nil {
		c.Country = req.Country
	}
	if req.Notes != nil {
		c.Notes = req.Notes
	}
	if req.Source != nil {
		c.Source = req.Source
	}
	if req.ResearchSummary != nil {
		c.ResearchSummary = req.ResearchSummary
	}

	if err := d.CompanyRepo.Update(ctx, c); err != nil {
		return nil, fmt.Errorf("updating company: %w", err)
	}
	return c, nil
}
