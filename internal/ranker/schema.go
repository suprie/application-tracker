package ranker

import "suprie/application_tracker/internal/llm"

// BuildJSONSchema returns the JSON Schema for the experience ranker response.
func BuildJSONSchema() llm.ResponseFormat {
	return llm.ResponseFormat{
		Type: "json_schema",
		JSONSchema: llm.JSONSchema{
			Name: "experience_ranker",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"selected_experiences": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"experience_id": map[string]any{"type": "string"},
								"title":         map[string]any{"type": "string"},
								"summary":       map[string]any{"type": "string"},
								"why_selected":  map[string]any{"type": "string"},
								"scores": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"relevance_to_job":          map[string]any{"type": "integer", "minimum": 0, "maximum": 5},
										"business_or_product_impact": map[string]any{"type": "integer", "minimum": 0, "maximum": 5},
										"technical_depth":           map[string]any{"type": "integer", "minimum": 0, "maximum": 5},
										"seniority_signal":          map[string]any{"type": "integer", "minimum": 0, "maximum": 5},
										"rarity":                    map[string]any{"type": "integer", "minimum": 0, "maximum": 5},
										"narrative_fit":             map[string]any{"type": "integer", "minimum": 0, "maximum": 5},
										"final_score":               map[string]any{"type": "number"},
									},
									"required": []string{
										"relevance_to_job", "business_or_product_impact",
										"technical_depth", "seniority_signal",
										"rarity", "narrative_fit", "final_score",
									},
								},
								"evidence": map[string]any{
									"type":  "array",
									"items": map[string]any{"type": "string"},
								},
							},
							"required": []string{"experience_id", "title", "summary", "why_selected", "scores", "evidence"},
						},
					},
					"selected_skills": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"skill":    map[string]any{"type": "string"},
								"evidence": map[string]any{"type": "string"},
							},
							"required": []string{"skill", "evidence"},
						},
					},
					"recommended_narrative": map[string]any{"type": "string"},
					"warnings": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
				},
				"required": []string{"selected_experiences", "selected_skills", "recommended_narrative"},
			},
		},
	}
}
