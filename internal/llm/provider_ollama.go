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
// support structured output (neither json_schema nor json_object). The JSON
// schema is inlined into the prompt text so the model still knows the expected
// output shape. No response_format field is sent — many Ollama models reject it.
type OllamaProvider struct{}

func (OllamaProvider) NewRequest(ctx context.Context, baseURL, model, prompt string, rf *ResponseFormat, apiKey *string) (*http.Request, error) {
	msg := prompt

	// Inline the schema into the prompt so the model produces valid JSON.
	if rf != nil {
		if schemaJSON, err := json.Marshal(rf.JSONSchema.Schema); err == nil {
			msg += "\n\nYou MUST respond with ONLY a JSON object. Do not wrap it in markdown fences.\n" +
				"CRITICAL: The schema below DESCRIBES the shape. You must produce an INSTANCE of it.\n" +
				"Do NOT output the schema itself. Do NOT include the word \"properties\" as a key.\n" +
				"For example, if the schema says {\"properties\":{\"name\":{\"type\":\"string\"}}},\n" +
				"you should output {\"name\":\"Alice\"}, NOT {\"properties\":{\"name\":\"Alice\"}}.\n\n" +
				"Schema:\n" + string(schemaJSON)
		}
	}

	body := ollamaRequest{
		Model:       model,
		Temperature: 0.2,
		Stream:      false,
		Messages: []message{
			{Role: "user", Content: msg},
		},
		Reasoning: "none",
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
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
	Stream      bool      `json:"stream"`
	Reasoning   string    `json:"reasoning_effort"`
}
