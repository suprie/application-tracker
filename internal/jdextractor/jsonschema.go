package jdextractor

import (
	"suprie/application_tracker/internal/llm"
)

func BuildJSONSchema() llm.ResponseFormat {
	schema := llm.ResponseFormat{
		Type: "json_schema",
		JSONSchema: llm.JSONSchema{
			Name: "parsed_job_description",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"company":          map[string]any{"type": []string{"string", "null"}},
					"role_title":       map[string]any{"type": []string{"string", "null"}},
					"seniority":        map[string]any{"type": []string{"string", "null"}},
					"location":         map[string]any{"type": []string{"string", "null"}},
					"work_arrangement": map[string]any{"type": []string{"string", "null"}},
					"employment_type":  map[string]any{"type": []string{"string", "null"}},
					"requirements": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"must_have": map[string]any{
								"type":  "array",
								"items": map[string]any{"type": "string"},
							},
							"nice_have": map[string]any{
								"type":  "array",
								"items": map[string]any{"type": "string"},
							},
						},
						"required": []string{
							"must_have",
							"nice_have",
						},
					},
					"responsibilities": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
					"parsing_warnings": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
					"keywords": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
				},
				"required": []string{
					"company",
					"role_title",
					"seniority",
					"location",
					"work_arrangement",
					"employment_type",
					"requirements",
					"responsibilities",
					"parsing_warnings",
				},
			},
		},
	}

	return schema
}
