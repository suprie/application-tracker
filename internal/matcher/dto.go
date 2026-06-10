package matcher

// MatchResponse is the structured output from the LLM fit-match analysis.
type MatchResponse struct {
	FitScore  int      `json:"fit_score"`
	GoNoGo    string   `json:"go_no_go"`
	Strengths []string `json:"strengths"`
	Gaps      []string `json:"gaps"`
	Summary   string   `json:"summary"`
}
