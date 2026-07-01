package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/AI2HU/gego/internal/db/sqlutil"
	"github.com/AI2HU/gego/internal/models"
)

type Postgres struct {
	db     *sql.DB
	config *models.Config
}

func New(config *models.Config) (*Postgres, error) {
	return &Postgres{config: config}, nil
}

func (p *Postgres) Connect(ctx context.Context) error {
	db, err := sql.Open("pgx", p.config.URI)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	p.db = db
	return nil
}

func (p *Postgres) Disconnect(ctx context.Context) error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p *Postgres) Ping(ctx context.Context) error {
	if p.db == nil {
		return fmt.Errorf("not connected to database")
	}
	return p.db.PingContext(ctx)
}

func (p *Postgres) GetDB() *sql.DB {
	return p.db
}

func (p *Postgres) DriverName() string {
	return "postgres"
}

func (p *Postgres) CreateLLM(ctx context.Context, llm *models.LLMConfig) error {
	llm.CreatedAt = time.Now()
	llm.UpdatedAt = time.Now()

	query := `
		INSERT INTO llms (id, name, provider, model, api_key, base_url, config, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10)`

	_, err := p.db.ExecContext(ctx, query,
		llm.ID, llm.Name, llm.Provider, llm.Model, llm.APIKey, llm.BaseURL,
		sqlutil.MapToJSON(llm.Config), llm.Enabled, llm.CreatedAt, llm.UpdatedAt,
	)
	return err
}

func (p *Postgres) GetLLM(ctx context.Context, id string) (*models.LLMConfig, error) {
	query := `
		SELECT id, name, provider, model, api_key, base_url, config::text, enabled, created_at, updated_at
		FROM llms WHERE id = $1`

	var llm models.LLMConfig
	var configJSON string

	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&llm.ID, &llm.Name, &llm.Provider, &llm.Model, &llm.APIKey, &llm.BaseURL,
		&configJSON, &llm.Enabled, &llm.CreatedAt, &llm.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("LLM not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	llm.Config = sqlutil.JSONToMap(configJSON)
	return &llm, nil
}

func (p *Postgres) ListLLMs(ctx context.Context, enabled *bool) ([]*models.LLMConfig, error) {
	query := `
		SELECT id, name, provider, model, api_key, base_url, config::text, enabled, created_at, updated_at
		FROM llms`
	args := []any{}

	if enabled != nil {
		query += " WHERE enabled = $1"
		args = append(args, *enabled)
	}

	query += " ORDER BY created_at DESC"

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var llms []*models.LLMConfig
	for rows.Next() {
		var llm models.LLMConfig
		var configJSON string

		if err := rows.Scan(
			&llm.ID, &llm.Name, &llm.Provider, &llm.Model, &llm.APIKey, &llm.BaseURL,
			&configJSON, &llm.Enabled, &llm.CreatedAt, &llm.UpdatedAt,
		); err != nil {
			return nil, err
		}

		llm.Config = sqlutil.JSONToMap(configJSON)
		llms = append(llms, &llm)
	}

	return llms, rows.Err()
}

func (p *Postgres) UpdateLLM(ctx context.Context, llm *models.LLMConfig) error {
	llm.UpdatedAt = time.Now()

	query := `
		UPDATE llms
		SET name = $1, provider = $2, model = $3, api_key = $4, base_url = $5, config = $6::jsonb, enabled = $7, updated_at = $8
		WHERE id = $9`

	result, err := p.db.ExecContext(ctx, query,
		llm.Name, llm.Provider, llm.Model, llm.APIKey, llm.BaseURL,
		sqlutil.MapToJSON(llm.Config), llm.Enabled, llm.UpdatedAt, llm.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("LLM not found: %s", llm.ID)
	}
	return nil
}

func (p *Postgres) DeleteLLM(ctx context.Context, id string) error {
	result, err := p.db.ExecContext(ctx, "DELETE FROM llms WHERE id = $1", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("LLM not found: %s", id)
	}
	return nil
}

func (p *Postgres) DeleteAllLLMs(ctx context.Context) (int, error) {
	result, err := p.db.ExecContext(ctx, "DELETE FROM llms")
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rowsAffected), nil
}

func (p *Postgres) CreateSchedule(ctx context.Context, schedule *models.Schedule) error {
	schedule.CreatedAt = time.Now()
	schedule.UpdatedAt = time.Now()

	query := `
		INSERT INTO schedules (id, name, prompt_ids, llm_ids, cron_expr, temperature, enabled, last_run, next_run, created_at, updated_at)
		VALUES ($1, $2, $3::jsonb, $4::jsonb, $5, $6, $7, $8, $9, $10, $11)`

	_, err := p.db.ExecContext(ctx, query,
		schedule.ID, schedule.Name,
		sqlutil.SliceToJSON(schedule.PromptIDs), sqlutil.SliceToJSON(schedule.LLMIDs),
		schedule.CronExpr, schedule.Temperature, schedule.Enabled,
		schedule.LastRun, schedule.NextRun, schedule.CreatedAt, schedule.UpdatedAt,
	)
	return err
}

func (p *Postgres) GetSchedule(ctx context.Context, id string) (*models.Schedule, error) {
	query := `
		SELECT id, name, prompt_ids::text, llm_ids::text, cron_expr, temperature, enabled, last_run, next_run, created_at, updated_at
		FROM schedules WHERE id = $1`

	var schedule models.Schedule
	var promptIDsJSON, llmIDsJSON string

	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&schedule.ID, &schedule.Name, &promptIDsJSON, &llmIDsJSON,
		&schedule.CronExpr, &schedule.Temperature, &schedule.Enabled,
		&schedule.LastRun, &schedule.NextRun, &schedule.CreatedAt, &schedule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("schedule not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	schedule.PromptIDs = sqlutil.JSONToSlice(promptIDsJSON)
	schedule.LLMIDs = sqlutil.JSONToSlice(llmIDsJSON)
	return &schedule, nil
}

func (p *Postgres) ListSchedules(ctx context.Context, enabled *bool) ([]*models.Schedule, error) {
	query := `
		SELECT id, name, prompt_ids::text, llm_ids::text, cron_expr, temperature, enabled, last_run, next_run, created_at, updated_at
		FROM schedules`
	args := []any{}

	if enabled != nil {
		query += " WHERE enabled = $1"
		args = append(args, *enabled)
	}

	query += " ORDER BY created_at DESC"

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.Schedule
	for rows.Next() {
		var schedule models.Schedule
		var promptIDsJSON, llmIDsJSON string

		if err := rows.Scan(
			&schedule.ID, &schedule.Name, &promptIDsJSON, &llmIDsJSON,
			&schedule.CronExpr, &schedule.Temperature, &schedule.Enabled,
			&schedule.LastRun, &schedule.NextRun, &schedule.CreatedAt, &schedule.UpdatedAt,
		); err != nil {
			return nil, err
		}

		schedule.PromptIDs = sqlutil.JSONToSlice(promptIDsJSON)
		schedule.LLMIDs = sqlutil.JSONToSlice(llmIDsJSON)
		schedules = append(schedules, &schedule)
	}

	return schedules, rows.Err()
}

func (p *Postgres) UpdateSchedule(ctx context.Context, schedule *models.Schedule) error {
	schedule.UpdatedAt = time.Now()

	query := `
		UPDATE schedules
		SET name = $1, prompt_ids = $2::jsonb, llm_ids = $3::jsonb, cron_expr = $4, temperature = $5,
		    enabled = $6, last_run = $7, next_run = $8, updated_at = $9
		WHERE id = $10`

	result, err := p.db.ExecContext(ctx, query,
		schedule.Name,
		sqlutil.SliceToJSON(schedule.PromptIDs), sqlutil.SliceToJSON(schedule.LLMIDs),
		schedule.CronExpr, schedule.Temperature, schedule.Enabled,
		schedule.LastRun, schedule.NextRun, schedule.UpdatedAt, schedule.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("schedule not found: %s", schedule.ID)
	}
	return nil
}

func (p *Postgres) DeleteSchedule(ctx context.Context, id string) error {
	result, err := p.db.ExecContext(ctx, "DELETE FROM schedules WHERE id = $1", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("schedule not found: %s", id)
	}
	return nil
}

func (p *Postgres) DeleteAllSchedules(ctx context.Context) (int, error) {
	result, err := p.db.ExecContext(ctx, "DELETE FROM schedules")
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rowsAffected), nil
}
