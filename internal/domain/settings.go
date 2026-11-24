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

// LLMSettingsUpdate contains only the fields a caller wants to change.
type LLMSettingsUpdate struct {
	Provider *string `json:"provider"`
	Model    *string `json:"model"`
	BaseURL  *string `json:"base_url"`
	APIKey   *string `json:"api_key"`
}
