package db

import (
	"context"
	"errors"

	"github.com/AI2HU/gego/internal/models"
)

var ErrSchedulerSettingsUnsupported = errors.New("scheduler settings require PostgreSQL; run `gego db upgrade-from-sqlite`")

type SchedulerSettingsDatabase interface {
	GetSchedulerSettings(ctx context.Context) (*models.SchedulerSettings, error)
	SetSchedulerDesiredRunning(ctx context.Context, desiredRunning bool) error
}

func schedulerSettingsBackend(sqlDB SQLBackend) (SchedulerSettingsDatabase, error) {
	s, ok := sqlDB.(SchedulerSettingsDatabase)
	if !ok {
		return nil, ErrSchedulerSettingsUnsupported
	}
	return s, nil
}
