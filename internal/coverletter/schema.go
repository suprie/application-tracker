package coverletter

import "suprie/application_tracker/internal/llm"

// BuildJSONSchema returns the JSON Schema for cover letter generation.
func BuildJSONSchema() llm.ResponseFormat {
	return llm.ResponseFormat{
		Type: "json_schema",
		JSONSchema: llm.JSONSchema{
			Name: "cover_letter",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"your_name":       map[string]any{"type": "string"},
					"your_address":    map[string]any{"type": "string"},
					"your_email":      map[string]any{"type": "string"},
					"your_phone":      map[string]any{"type": "string"},
					"recipient_name":  map[string]any{"type": "string"},
					"recipient_title": map[string]any{"type": "string"},
					"company_name":    map[string]any{"type": "string"},
					"company_address": map[string]any{"type": "string"},
					"subject":         map[string]any{"type": "string"},
					"opening":         map[string]any{"type": "string"},
					"opening_paragraphs":         map[string]any{"type": "string"},
					"body_paragraphs": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
					"closing_paragraphs": map[string]any{"type": "string"},
					"closing": map[string]any{"type": "string"},
				},
				"required": []string{
					"your_name", "recipient_name", "subject",
					"opening", "body_paragraphs", "closing",
				},
			},
		},
	}
}
