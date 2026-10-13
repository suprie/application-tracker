package service

import (
	"context"
	"errors"
	"fmt"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/dto"
	"suprie/application_tracker/internal/llm"
)

// GetLLMSettings returns the stored LLM settings, or nil if none are saved
// yet (the server then falls back to AI_* env vars).
func GetLLMSettings(ctx context.Context, d Deps) (*domain.LLMSettings, error) {
	s, err := d.SettingsRepo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading llm settings: %w", err)
	}
	return s, nil
}

// UpdateLLMSettings merges the non-nil fields of req onto the stored
// settings (or a zero-value baseline if none exist yet) and persists it.
func UpdateLLMSettings(ctx context.Context, d Deps, req dto.UpdateLLMSettingsRequest) (*domain.LLMSettings, error) {
	existing, err := d.SettingsRepo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading llm settings: %w", err)
	}
	if existing == nil {
		existing = &domain.LLMSettings{}
	}

	if req.Provider != nil {
		existing.Provider = *req.Provider
	}
	if req.Model != nil {
		existing.Model = *req.Model
	}
	if req.BaseURL != nil {
		existing.BaseURL = *req.BaseURL
	}
	if req.APIKey != nil {
		existing.APIKey = *req.APIKey
	}
	if existing.Provider == "" {
		return nil, errors.New("provider is required")
	}

	if err := d.SettingsRepo.Save(ctx, existing); err != nil {
		return nil, fmt.Errorf("saving llm settings: %w", err)
	}
	return existing, nil
}

// SettingsAwareLLMFactory builds an LLM client factory that prefers the
// repo-backed settings (set via the web UI) and falls back to AI_* env vars
// when none have been saved yet.
func SettingsAwareLLMFactory(d Deps) func(stage string) llm.LLMClient {
	return func(stage string) llm.LLMClient {
		s, err := d.SettingsRepo.Get(context.Background())
		if err != nil || s == nil {
			return llm.NewClient(stage)
		}
		return llm.NewClientFromSettings(stage, llm.Settings{
			Provider: s.Provider,
			Model:    s.Model,
			BaseURL:  s.BaseURL,
			APIKey:   s.APIKey,
		})
	}
}
