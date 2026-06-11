package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func init() {
	RegisterProvider("ollama", OllamaProvider{})
}

// OllamaProvider implements the Ollama OpenAI-compatible API. Ollama does not
// support json_schema structured output — it only understands json_object mode.
// The JSON schema is inlined into the prompt so the model still knows the
// expected output shape.
type OllamaProvider struct{}

func (OllamaProvider) NewRequest(ctx context.Context, baseURL, model, prompt string, rf *ResponseFormat, apiKey *string) (*http.Request, error) {
	body := ollamaRequest{
		Model:       model,
		Temperature: 0.2,
		Stream:      false,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}

	// Ollama supports json_object mode but not json_schema. Inline the
	// schema into the prompt so the model produces valid structured output.
	if rf != nil {
		body.ResponseFormat = &jsonObjectFormat{Type: "json_object"}
		if schemaJSON, err := json.Marshal(rf.JSONSchema.Schema); err == nil {
			body.Messages[0].Content += "\n\nYou MUST respond with ONLY a JSON object. Do not wrap it in markdown fences.\nThe JSON object must match this schema:\n" + string(schemaJSON)
		}
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != nil {
		req.Header.Set("Authorization", "Bearer "+*apiKey)
	}
	return req, nil
}

func (OllamaProvider) ParseResponse(body []byte) (string, error) {
	return parseOpenAIResponse(body)
}

type ollamaRequest struct {
	Model          string            `json:"model"`
	Messages       []message         `json:"messages"`
	Temperature    float64           `json:"temperature"`
	Stream         bool              `json:"stream"`
	ResponseFormat *jsonObjectFormat `json:"response_format,omitempty"`
}
