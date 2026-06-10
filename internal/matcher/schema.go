package matcher

import "suprie/application_tracker/internal/llm"

// BuildMatchJSONSchema returns the JSON Schema for the LLM structured response.
func BuildMatchJSONSchema() llm.ResponseFormat {
	return llm.ResponseFormat{
		Type: "json_schema",
		JSONSchema: llm.JSONSchema{
			Name: "fit_match",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"fit_score": map[string]any{
						"type":        "integer",
						"description": "Fit percentage, 0-100. Higher = better match.",
						"minimum":     0,
						"maximum":     100,
					},
					"go_no_go": map[string]any{
						"type":        "string",
						"description": "Overall recommendation: 'go' or 'no_go'.",
						"enum":        []string{"go", "no_go"},
					},
					"strengths": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
						"description": "Key areas where the candidate matches or exceeds requirements.",
					},
					"gaps": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
						"description": "Key areas where the candidate is missing required skills or experience.",
					},
					"summary": map[string]any{
						"type":        "string",
						"description": "Concise 2-3 sentence summary of the overall fit.",
					},
				},
				"required": []string{"fit_score", "go_no_go", "strengths", "gaps", "summary"},
			},
		},
	}
}
