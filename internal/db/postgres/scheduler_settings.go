package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/AI2HU/gego/internal/models"
)

const schedulerSettingsID = "default"

func (p *Postgres) GetSchedulerSettings(ctx context.Context) (*models.SchedulerSettings, error) {
	query := `
		SELECT id, desired_running, created_at, updated_at
		FROM scheduler_settings
		WHERE id = $1`

	var s models.SchedulerSettings
	err := p.db.QueryRowContext(ctx, query, schedulerSettingsID).Scan(
		&s.ID, &s.DesiredRunning, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return &models.SchedulerSettings{ID: schedulerSettingsID}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduler settings: %w", err)
	}
	return &s, nil
}

func (p *Postgres) SetSchedulerDesiredRunning(ctx context.Context, desiredRunning bool) error {
	now := time.Now()
	query := `
		INSERT INTO scheduler_settings (id, desired_running, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			desired_running = EXCLUDED.desired_running,
			updated_at = EXCLUDED.updated_at`

	_, err := p.db.ExecContext(ctx, query, schedulerSettingsID, desiredRunning, now, now)
	if err != nil {
		return fmt.Errorf("failed to set scheduler desired running: %w", err)
	}
	return nil
}
