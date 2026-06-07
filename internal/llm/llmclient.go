package llm

import (
	"context"
)

type ResponseFormat struct {
	Type       string     `json:"type"`
	JSONSchema JSONSchema `json:"json_schema"`
}

type JSONSchema struct {
	Name   string         `json:"name"`
	Schema map[string]any `json:"schema"`
}

type LLMClient interface {
	Generate(ctx context.Context, prompt string, responseFormat *ResponseFormat) (string, error)
}
