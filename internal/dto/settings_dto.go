package dto

import (
	"time"

	"suprie/application_tracker/internal/domain"
)

// LLMSettingsResponse is the API representation of the global LLM config.
// The API key is never echoed back — only whether one is set.
type LLMSettingsResponse struct {
	Provider  string     `json:"provider"`
	Model     string     `json:"model"`
	BaseURL   string     `json:"base_url"`
	APIKeySet bool       `json:"api_key_set"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// UpdateLLMSettingsRequest is the body of PUT /api/settings/llm. Pointer
// fields override the stored value only when present; APIKey omitted leaves
// the stored key unchanged, an explicit "" clears it.
type UpdateLLMSettingsRequest struct {
	Provider *string `json:"provider"`
	Model    *string `json:"model"`
	BaseURL  *string `json:"base_url"`
	APIKey   *string `json:"api_key"`
}

func ToLLMSettingsResponse(s *domain.LLMSettings) LLMSettingsResponse {
	resp := LLMSettingsResponse{
		Provider:  s.Provider,
		Model:     s.Model,
		BaseURL:   s.BaseURL,
		APIKeySet: s.APIKey != "",
	}
	if !s.UpdatedAt.IsZero() {
		resp.UpdatedAt = &s.UpdatedAt
	}
	return resp
}
