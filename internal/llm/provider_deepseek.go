package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func init() {
	RegisterProvider("deepseek", DeepSeekProvider{})
}

// DeepSeekProvider implements the DeepSeek API, which is OpenAI-compatible
// but uses json_object response format and does not support reasoning_effort.
type DeepSeekProvider struct{}

func (DeepSeekProvider) NewRequest(ctx context.Context, baseURL, model, prompt string, rf *ResponseFormat, apiKey *string) (*http.Request, error) {
	body := deepSeekRequest{
		Model:       model,
		Temperature: 0.2,
		Stream:      false,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}

	// DeepSeek supports json_object mode. The JSON schema is inlined into
	// the prompt so the model knows the expected output shape.
	if rf != nil {
		body.ResponseFormat = &jsonObjectFormat{Type: "json_object"}
		if schemaJSON, err := json.Marshal(rf.JSONSchema.Schema); err == nil {
			body.Messages[0].Content += "\n\nRespond with a JSON object matching this schema:\n" + string(schemaJSON)
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

func (DeepSeekProvider) ParseResponse(body []byte) (string, error) {
	return parseOpenAIResponse(body)
}

type deepSeekRequest struct {
	Model          string            `json:"model"`
	Messages       []message         `json:"messages"`
	Temperature    float64           `json:"temperature"`
	Stream         bool              `json:"stream"`
	ResponseFormat *jsonObjectFormat `json:"response_format,omitempty"`
}

type jsonObjectFormat struct {
	Type string `json:"type"`
}
