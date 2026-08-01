package llm

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type LMStudioClient struct {
	BaseURL  string
	Model    string
	APIKey   *string
	Provider Provider
	Stage    string       // human-readable label for request logs (e.g. "match", "parse_cv")
	client   *http.Client // per-stage client — avoids Ollama cancelling the first request when a second starts
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
		Stage:    stageName,
		client:   newStageHTTPClient(),
	}
}

// Settings is the subset of persisted configuration (e.g. from a web
// settings page) needed to build an LLMClient, overriding the AI_* env vars.
type Settings struct {
	Provider string
	Model    string
	BaseURL  string
	APIKey   string
}

// NewClientFromSettings builds an LLMClient from explicit settings rather
// than environment variables — the seam used when a user configures the
// provider through the web UI instead of AI_* env vars.
func NewClientFromSettings(stage string, s Settings) LLMClient {
	if cmd, ok := harnessCommands[s.Provider]; ok {
		return NewHarnessClient(cmd, stage)
	}

	provider, err := GetProvider(s.Provider)
	if err != nil {
		provider, _ = GetProvider("openai")
	}

	baseURL := s.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:1234/v1"
	}
	model := s.Model
	if model == "" {
		model = "gemma-4-12b-qat"
	}
	var apiKey *string
	if s.APIKey != "" {
		apiKey = &s.APIKey
	}

	return LMStudioClient{
		BaseURL:  baseURL,
		Model:    model,
		APIKey:   apiKey,
		Provider: provider,
		Stage:    stage,
		client:   newStageHTTPClient(),
	}
}

// newStageHTTPClient returns a fresh per-stage HTTP client — avoids Ollama
// cancelling the first request when a second stage starts.
func newStageHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
}

// NewClient builds an LLMClient for the given stage, choosing between an
// HTTP API client and a harness CLI client based on AI_PROVIDER (or its
// stage-specific override) — "claude-harness" and "codex-harness" select
// the matching CLI; anything else resolves to an HTTP Provider as before.
func NewClient(stage ...string) LLMClient {
	stageName := ""
	if len(stage) > 0 {
		stageName = stage[0]
	}
	providerName := resolveStage("AI_PROVIDER", stageName, "openai")
	if cmd, ok := harnessCommands[providerName]; ok {
		return NewHarnessClient(cmd, stageName)
	}
	return NewLMStudioClient(stage...)
}

// Generate sends a prompt to the LLM and returns the generated text.
func (c LMStudioClient) Generate(ctx context.Context, prompt string, responseFormat *ResponseFormat) (string, error) {
	req, err := c.Provider.NewRequest(ctx, c.BaseURL, c.Model, prompt, responseFormat, c.APIKey)
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}

	log.Printf("→ LLM request [%s] %s %s model=%s prompt=%dchars",
		c.Stage, req.Method, req.URL.String(), c.Model, len(prompt))

	start := time.Now()
	hc := c.client
	if hc == nil {
		hc = httpClient()
	}
	res, err := hc.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		log.Printf("✗ LLM error [%s] after %s: %v", c.Stage, elapsed.Round(time.Millisecond), err)
		return "", fmt.Errorf("send request: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("✗ LLM read error [%s]: %v", c.Stage, err)
		return "", fmt.Errorf("reading body: %w", err)
	}

	log.Printf("← LLM response [%s] %d %s %dchars",
		c.Stage, res.StatusCode, elapsed.Round(time.Millisecond), len(bodyBytes))

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
