package dto

type JobDescriptionResponse struct {
	Company          *string       `json:"company"`
	RoleTitle        *string       `json:"role_title"`
	Seniority        *string       `json:"seniority"`
	Location         *string       `json:"location"`
	WorkArrangement  *string       `json:"work_arrangement"`
	EmploymentType   *string       `json:"employment_type"`
	Requirements     *Requirements `json:"requirements"`
	Responsibilities []string      `json:"responsibilities"`
	ParsingWarnings  []string      `json:"parsing_warnings"`
	Keywords         []string      `json:"keywords"`
}

type Requirements struct {
	MustHave []string `json:"must_have"`
	NiceHave []string `json:"nice_have"`
}
