package dto

import (
	"time"

	"suprie/application_tracker/internal/domain"
)

// CompanyResponse is the API representation of a Company.
type CompanyResponse struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	NormalizedName  string    `json:"normalized_name"`
	WebsiteURL      *string   `json:"website_url"`
	Industry        *string   `json:"industry"`
	Size            *string   `json:"size"`
	Country         *string   `json:"country"`
	Notes           *string   `json:"notes"`
	Source          *string   `json:"source"`
	ResearchSummary *string   `json:"research_summary"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateCompanyRequest is the body of POST /api/companies.
type CreateCompanyRequest struct {
	Name            string  `json:"name" binding:"required"`
	WebsiteURL      *string `json:"website_url"`
	Industry        *string `json:"industry"`
	Size            *string `json:"size"`
	Country         *string `json:"country"`
	Notes           *string `json:"notes"`
	Source          *string `json:"source"`
	ResearchSummary *string `json:"research_summary"`
}

// UpdateCompanyRequest is the body of PUT /api/companies/:id. Pointer fields
// override the stored value only when non-nil.
type UpdateCompanyRequest struct {
	Name            *string `json:"name"`
	WebsiteURL      *string `json:"website_url"`
	Industry        *string `json:"industry"`
	Size            *string `json:"size"`
	Country         *string `json:"country"`
	Notes           *string `json:"notes"`
	Source          *string `json:"source"`
	ResearchSummary *string `json:"research_summary"`
}

// ToCompanyResponse maps a domain Company to an API response.
func ToCompanyResponse(c *domain.Company) CompanyResponse {
	return CompanyResponse{
		ID:              c.ID,
		Name:            c.Name,
		NormalizedName:  c.NormalizedName,
		WebsiteURL:      c.WebsiteURL,
		Industry:        c.Industry,
		Size:            c.Size,
		Country:         c.Country,
		Notes:           c.Notes,
		Source:          c.Source,
		ResearchSummary: c.ResearchSummary,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}
