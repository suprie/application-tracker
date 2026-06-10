package llm

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type LMStudioClient struct {
	BaseURL  string
	Model    string
	APIKey   *string
	Provider Provider
}

// NewLMStudioClient creates a client configured from environment variables.
// Pass a stage name (e.g. "parse_jd", "match") to use stage-specific overrides:
//
//	AI_PROVIDER_{STAGE} → AI_PROVIDER → "openai"
//	AI_MODEL_{STAGE}    → AI_MODEL    → built-in default
//	AI_URL_{STAGE}      → AI_URL      → built-in default
//	AI_API_KEY_{STAGE}  → AI_API_KEY  → empty (no auth header sent)
func NewLMStudioClient(stage ...string) LMStudioClient {
	stageName := ""
	if len(stage) > 0 {
		stageName = stage[0]
	}

	providerName := resolveStage("AI_PROVIDER", stageName, "openai")
	provider, err := GetProvider(providerName)
	if err != nil {
		// Fall back to OpenAI if the configured provider is unknown.
		provider, _ = GetProvider("openai")
	}

	return LMStudioClient{
		BaseURL:  resolveStage("AI_URL", stageName, "http://localhost:1234/v1"),
		Model:    resolveStage("AI_MODEL", stageName, "gemma-4-12b-qat"),
		APIKey:   resolveStageAPIKey("AI_API_KEY", stageName),
		Provider: provider,
	}
}

// Generate sends a prompt to the LLM and returns the generated text.
func (c LMStudioClient) Generate(ctx context.Context, prompt string, responseFormat *ResponseFormat) (string, error) {
	req, err := c.Provider.NewRequest(ctx, c.BaseURL, c.Model, prompt, responseFormat, c.APIKey)
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}

	res, err := httpClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("reading body: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("llm error (%d): %s", res.StatusCode, string(bodyBytes))
	}

	content, err := c.Provider.ParseResponse(bodyBytes)
	if err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}
	return content, nil
}

// --- env helpers ---

func resolveStage(baseKey, stage, fallback string) string {
	if stage != "" {
		if v := os.Getenv(baseKey + "_" + strings.ToUpper(stage)); v != "" {
			return v
		}
	}
	return envOrDefault(baseKey, fallback)
}

func resolveStageAPIKey(baseKey, stage string) *string {
	if stage != "" {
		if v := os.Getenv(baseKey + "_" + strings.ToUpper(stage)); v != "" {
			return &v
		}
	}
	if v := os.Getenv(baseKey); v != "" {
		return &v
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// httpClient is a test-replaceable HTTP client.
var httpClient = func() *http.Client { return http.DefaultClient } //nolint:gochecknoglobals
