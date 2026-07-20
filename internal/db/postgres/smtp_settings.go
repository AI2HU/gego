package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/AI2HU/gego/internal/models"
)

const smtpSettingsID = "default"

func (p *Postgres) GetSMTPSettings(ctx context.Context) (*models.SMTPSettings, error) {
	query := `
		SELECT id, host, port, username, password, from_email, from_name, use_tls, enabled, created_at, updated_at
		FROM smtp_settings
		WHERE id = $1`

	var s models.SMTPSettings
	err := p.db.QueryRowContext(ctx, query, smtpSettingsID).Scan(
		&s.ID, &s.Host, &s.Port, &s.Username, &s.Password,
		&s.FromEmail, &s.FromName, &s.UseTLS, &s.Enabled,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return &models.SMTPSettings{
			ID:     smtpSettingsID,
			Port:   587,
			UseTLS: true,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SMTP settings: %w", err)
	}
	return &s, nil
}

func (p *Postgres) UpsertSMTPSettings(ctx context.Context, settings *models.SMTPSettings) error {
	now := time.Now()
	settings.ID = smtpSettingsID
	settings.UpdatedAt = now
	if settings.CreatedAt.IsZero() {
		settings.CreatedAt = now
	}

	query := `
		INSERT INTO smtp_settings (
			id, host, port, username, password, from_email, from_name, use_tls, enabled, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			host = EXCLUDED.host,
			port = EXCLUDED.port,
			username = EXCLUDED.username,
			password = EXCLUDED.password,
			from_email = EXCLUDED.from_email,
			from_name = EXCLUDED.from_name,
			use_tls = EXCLUDED.use_tls,
			enabled = EXCLUDED.enabled,
			updated_at = EXCLUDED.updated_at`

	_, err := p.db.ExecContext(ctx, query,
		settings.ID, settings.Host, settings.Port, settings.Username, settings.Password,
		settings.FromEmail, settings.FromName, settings.UseTLS, settings.Enabled,
		settings.CreatedAt, settings.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert SMTP settings: %w", err)
	}
	return nil
}
