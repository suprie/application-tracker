package domain

import "time"

// Company is a manually researched employer record. normalized_name drives
// deduplication and lookup from parsed job descriptions.
type Company struct {
	ID              int
	Name            string
	NormalizedName  string
	WebsiteURL      *string
	Industry        *string
	Size            *string
	Country         *string
	Notes           *string
	Source          *string
	ResearchSummary *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
