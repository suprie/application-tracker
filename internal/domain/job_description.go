package domain

import (
	"time"
)

type JobDescription struct {
	Id                   int
	Company              *string
	RoleTitle            *string
	Seniority            *string
	EmploymentType       *string
	WorkArrangement      *string
	Location             *string
	RequirementsJson     string
	ResponsibilitiesJson string
	KeywordsJson         string
	ParsingWarningJson   string
	CreatedAt            time.Time
}
