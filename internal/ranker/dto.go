package ranker

// Response is the structured output from the experience ranker agent.
type Response struct {
	SelectedExperiences []RankedExperience `json:"selected_experiences"`
	SelectedSkills      []RankedSkill      `json:"selected_skills"`
	RecommendedNarrative string            `json:"recommended_narrative"`
	Warnings            []string           `json:"warnings"`
}

// RankedExperience is one ranked experience with scores and evidence.
type RankedExperience struct {
	ExperienceID string       `json:"experience_id"`
	Title        string       `json:"title"`
	Summary      string       `json:"summary"`
	WhySelected  string       `json:"why_selected"`
	Scores       ExperienceScores `json:"scores"`
	Evidence     []string     `json:"evidence"`
}

// ExperienceScores holds the 0–5 rubric scores for one experience.
type ExperienceScores struct {
	RelevanceToJob       int `json:"relevance_to_job"`
	BusinessOrProductImpact int `json:"business_or_product_impact"`
	TechnicalDepth       int `json:"technical_depth"`
	SenioritySignal      int `json:"seniority_signal"`
	Rarity               int `json:"rarity"`
	NarrativeFit         int `json:"narrative_fit"`
	FinalScore           float64 `json:"final_score"`
}

// RankedSkill is one selected skill with evidence.
type RankedSkill struct {
	Skill    string `json:"skill"`
	Evidence string `json:"evidence"`
}
