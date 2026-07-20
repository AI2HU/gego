package db

import (
	"context"
	"errors"

	"github.com/AI2HU/gego/internal/models"
)

// ErrSMTPSettingsUnsupported is returned when SMTP settings are used with legacy SQLite.
var ErrSMTPSettingsUnsupported = errors.New("SMTP settings require PostgreSQL; run `gego db upgrade-from-sqlite`")

// SMTPSettingsDatabase defines SMTP configuration operations for PostgreSQL-backed deployments.
type SMTPSettingsDatabase interface {
	GetSMTPSettings(ctx context.Context) (*models.SMTPSettings, error)
	UpsertSMTPSettings(ctx context.Context, settings *models.SMTPSettings) error
}

func smtpSettingsBackend(sqlDB SQLBackend) (SMTPSettingsDatabase, error) {
	s, ok := sqlDB.(SMTPSettingsDatabase)
	if !ok {
		return nil, ErrSMTPSettingsUnsupported
	}
	return s, nil
}
