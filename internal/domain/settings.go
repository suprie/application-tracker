package domain

import "time"

// LLMSettings is the single, global LLM configuration used by every pipeline
// stage. It overrides the AI_* environment variables when present.
type LLMSettings struct {
	Provider  string // "openai", "deepseek", "ollama", "claude-harness", "codex-harness"
	Model     string
	BaseURL   string
	APIKey    string
	UpdatedAt time.Time
}
