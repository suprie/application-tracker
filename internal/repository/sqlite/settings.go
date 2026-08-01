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

func (r *SettingsRepository) Save(ctx context.Context, s *domain.LLMSettings) error {
	s.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO llm_settings (id, provider, model, base_url, api_key, updated_at)
		 VALUES (1, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   provider = excluded.provider,
		   model = excluded.model,
		   base_url = excluded.base_url,
		   api_key = excluded.api_key,
		   updated_at = excluded.updated_at`,
		s.Provider, s.Model, s.BaseURL, s.APIKey, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save llm settings: %w", err)
	}
	return nil
}
