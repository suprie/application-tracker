package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/repository"
)

var _ repository.SettingsRepository = (*SettingsRepository)(nil)

// SettingsRepository persists the single global LLM configuration as a
// fixed id=1 row.
type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) Get(ctx context.Context) (*domain.LLMSettings, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT provider, model, base_url, api_key, updated_at FROM llm_settings WHERE id = 1")

	var s domain.LLMSettings
	err := row.Scan(&s.Provider, &s.Model, &s.BaseURL, &s.APIKey, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get llm settings: %w", err)
	}
	return &s, nil
}

func (r *SettingsRepository) Update(ctx context.Context, patch domain.LLMSettingsUpdate) (*domain.LLMSettings, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO llm_settings (id, provider, model, base_url, api_key, updated_at)
		SELECT 1, COALESCE(?1, ''), COALESCE(?2, ''), COALESCE(?3, ''), COALESCE(?4, ''), ?5
		WHERE ?1 IS NOT NULL OR EXISTS (SELECT 1 FROM llm_settings WHERE id = 1)
		ON CONFLICT(id) DO UPDATE SET
			provider = COALESCE(?1, provider),
			model = COALESCE(?2, model),
			base_url = COALESCE(?3, base_url),
			api_key = CASE
				WHEN ?4 IS NOT NULL THEN ?4
				WHEN COALESCE(?1, provider) <> provider OR COALESCE(?3, base_url) <> base_url THEN ''
				ELSE api_key
			END,
			updated_at = excluded.updated_at
		RETURNING provider, model, base_url, api_key, updated_at`,
		patch.Provider, patch.Model, patch.BaseURL, patch.APIKey, time.Now().UTC(),
	)

	var s domain.LLMSettings
	err := row.Scan(&s.Provider, &s.Model, &s.BaseURL, &s.APIKey, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrSettingsProviderRequired
	}
	if err != nil {
		return nil, fmt.Errorf("update llm settings: %w", err)
	}
	return &s, nil
}
