package dto

import (
	"encoding/json"
	"time"

	"suprie/application_tracker/internal/domain"
)

// JDListItem is the lightweight representation used in the list view.
type JDListItem struct {
	ID        int       `json:"id"`
	Company   *string   `json:"company"`
	RoleTitle *string   `json:"role_title"`
	Status    string    `json:"status"`
	FitScore  *int      `json:"fit_score"`
	ApplyURL  *string   `json:"apply_url"`
	CreatedAt time.Time `json:"created_at"`
}

// JDResponse is the full representation used in the detail view. JSON columns
// are inlined as raw JSON rather than escaped strings.
type JDResponse struct {
	ID               int             `json:"id"`
	Company          *string         `json:"company"`
	RoleTitle        *string         `json:"role_title"`
	Seniority        *string         `json:"seniority"`
	EmploymentType   *string         `json:"employment_type"`
	WorkArrangement  *string         `json:"work_arrangement"`
	Location         *string         `json:"location"`
	ApplyURL         *string         `json:"apply_url"`
	Status           string          `json:"status"`
	FitScore         *int            `json:"fit_score"`
	FitSummary       *string         `json:"fit_summary"`
	Requirements     json.RawMessage `json:"requirements"`
	Responsibilities json.RawMessage `json:"responsibilities"`
	Keywords         json.RawMessage `json:"keywords"`
	ParsingWarnings  json.RawMessage `json:"parsing_warnings"`
	RankerResult     json.RawMessage `json:"ranker_result,omitempty"`
	AppliedAt        *time.Time      `json:"applied_at"`
	CreatedAt        time.Time       `json:"created_at"`
}

// PatchJDRequest is the body of PATCH /api/jds/:id. Fields are optional.
type PatchJDRequest struct {
	Status   *string `json:"status"`
	ApplyURL *string `json:"apply_url"`
}

// rawJSON wraps a stored JSON column for inline marshalling. Empty → null.
func rawJSON(s string) json.RawMessage {
	if s == "" {
		return json.RawMessage("null")
	}
	return json.RawMessage(s)
}

// ToJDListItem maps a domain JobDescription to a list item.
func ToJDListItem(jd *domain.JobDescription) JDListItem {
	return JDListItem{
		ID:        jd.ID,
		Company:   jd.Company,
		RoleTitle: jd.RoleTitle,
		Status:    jd.Status,
		FitScore:  jd.FitScore,
		ApplyURL:  jd.ApplyURL,
		CreatedAt: jd.CreatedAt,
	}
}

// ToJDResponse maps a domain JobDescription to a full API response.
func ToJDResponse(jd *domain.JobDescription) JDResponse {
	var ranker json.RawMessage
	if jd.RankerResultJSON != nil && *jd.RankerResultJSON != "" {
		ranker = rawJSON(*jd.RankerResultJSON)
	}
	return JDResponse{
		ID:               jd.ID,
		Company:          jd.Company,
		RoleTitle:        jd.RoleTitle,
		Seniority:        jd.Seniority,
		EmploymentType:   jd.EmploymentType,
		WorkArrangement:  jd.WorkArrangement,
		Location:         jd.Location,
		ApplyURL:         jd.ApplyURL,
		Status:           jd.Status,
		FitScore:         jd.FitScore,
		FitSummary:       jd.FitSummary,
		Requirements:     rawJSON(jd.RequirementsJSON),
		Responsibilities: rawJSON(jd.ResponsibilitiesJSON),
		Keywords:         rawJSON(jd.KeywordsJSON),
		ParsingWarnings:  rawJSON(jd.ParsingWarningJSON),
		RankerResult:     ranker,
		AppliedAt:        jd.AppliedAt,
		CreatedAt:        jd.CreatedAt,
	}
}
