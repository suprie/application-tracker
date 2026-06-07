package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type LMStudioClient struct {
	BaseURL string
	Model   string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model           string          `json:"model"`
	Messages        []Message       `json:"messages"`
	Temperature     float64         `json:"temperature"`
	ResponseFormat  *ResponseFormat `json:"response_format,omitempty"`
	ReasoningEffort string          `json:"reasoning_effort"`
}

func NewLMStudioClient() LMStudioClient {
	return LMStudioClient{
		BaseURL: "http://localhost:1234",
		Model:   "gemma-4-12b-qat",
	}
}

func (c LMStudioClient) Generate(ctx context.Context, prompt string, responseFormat *ResponseFormat) (string, error) {
	chatRequest := ChatRequest{
		Model:       c.Model,
		Temperature: 0.2,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		ResponseFormat:  responseFormat,
		ReasoningEffort: "none",
	}

	payload, err := json.Marshal(chatRequest)
	if err != nil {
		return "", fmt.Errorf("marshal chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.BaseURL+"/v1/chat/completions",
		bytes.NewReader(payload),
	)

	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}

	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("reading body: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("lm studio error (%d) %s", res.StatusCode, bodyBytes)
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("unmarshal response: %w,\nbody: %s", err, string(bodyBytes))
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("empty LLM response")
	}
	return response.Choices[0].Message.Content, nil
}
