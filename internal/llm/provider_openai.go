package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func init() {
	RegisterProvider("openai", OpenAIProvider{})
}

// OpenAIProvider implements the OpenAI /chat/completions API with full
// json_schema structured output support.
type OpenAIProvider struct{}

func (OpenAIProvider) NewRequest(ctx context.Context, baseURL, model, prompt string, rf *ResponseFormat, apiKey *string) (*http.Request, error) {
	body := openAIRequest{
		Model:       model,
		Temperature: 0.2,
		Stream:      false,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}
	if rf != nil {
		body.ResponseFormat = rf
		body.ReasoningEffort = "none"
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

func (OpenAIProvider) ParseResponse(body []byte) (string, error) {
	return parseOpenAIResponse(body)
}

type openAIRequest struct {
	Model           string          `json:"model"`
	Messages        []message       `json:"messages"`
	Temperature     float64         `json:"temperature"`
	Stream          bool            `json:"stream"`
	ResponseFormat  *ResponseFormat `json:"response_format,omitempty"`
	ReasoningEffort string          `json:"reasoning_effort,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func parseOpenAIResponse(body []byte) (string, error) {
	var resp openAIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty LLM response")
	}
	return resp.Choices[0].Message.Content, nil
}
