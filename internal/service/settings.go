package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/llm"
	"suprie/application_tracker/internal/repository"
)

var ErrInvalidLLMSettings = errors.New("invalid LLM settings")

// GetLLMSettings returns nil when the server is still using environment settings.
func GetLLMSettings(ctx context.Context, d Deps) (*domain.LLMSettings, error) {
	s, err := d.SettingsRepo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading LLM settings: %w", err)
	}
	return s, nil
}

func UpdateLLMSettings(ctx context.Context, d Deps, patch domain.LLMSettingsUpdate) (*domain.LLMSettings, error) {
	if patch.Provider != nil && !llm.SupportsProvider(*patch.Provider) {
		return nil, fmt.Errorf("%w: choose a supported provider", ErrInvalidLLMSettings)
	}
	if patch.BaseURL != nil && *patch.BaseURL != "" {
		u, err := url.Parse(*patch.BaseURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") ||
			u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			return nil, fmt.Errorf("%w: base URL must be an HTTP(S) endpoint without credentials, a query, or a fragment", ErrInvalidLLMSettings)
		}
	}
	if patch.APIKey != nil && strings.ContainsAny(*patch.APIKey, "\r\n") {
		return nil, fmt.Errorf("%w: API key contains a line break", ErrInvalidLLMSettings)
	}

	s, err := d.SettingsRepo.Update(ctx, patch)
	if errors.Is(err, repository.ErrSettingsProviderRequired) {
		return nil, fmt.Errorf("%w: %v", ErrInvalidLLMSettings, err)
	}
	return s, err
}

func SettingsAwareLLMFactory(d Deps) func(stage string) llm.LLMClient {
	return func(stage string) llm.LLMClient {
		return settingsAwareLLM{deps: d, stage: stage}
	}
}

type settingsAwareLLM struct {
	deps  Deps
	stage string
}

func (c settingsAwareLLM) Generate(ctx context.Context, prompt string, format *llm.ResponseFormat) (string, error) {
	s, err := GetLLMSettings(ctx, c.deps)
	if err != nil {
		return "", err
	}
	if s == nil {
		return llm.NewClient(c.stage).Generate(ctx, prompt, format)
	}
	client := llm.NewClientFromSettings(c.stage, llm.Settings{
		Provider: s.Provider,
		Model:    s.Model,
		BaseURL:  s.BaseURL,
		APIKey:   s.APIKey,
	})
	return client.Generate(ctx, prompt, format)
}
