package domain

import (
	"time"
)

// Application status constants.
const (
	StatusDraft    = "Draft"
	StatusFitMatch = "Fit match"
	StatusApplied  = "Applied"
	StatusRejected = "Rejected"
	StatusOffer    = "Offer"
)

type JobDescription struct {
	ID                  int
	Company              *string
	RoleTitle            *string
	Seniority            *string
	EmploymentType       *string
	WorkArrangement      *string
	Location             *string
	RequirementsJSON     string
	ResponsibilitiesJSON string
	KeywordsJSON         string
	ParsingWarningJSON   string
	Status               string
	ApplyURL             *string
	FitScore             *int
	FitSummary           *string
	RankerResultJSON     *string
	AppliedAt            *time.Time
	CreatedAt            time.Time
}
