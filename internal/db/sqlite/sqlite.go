package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/AI2HU/gego/internal/models"
)

// SQLite implements the Database interface for SQLite
type SQLite struct {
	db     *sql.DB
	config *models.Config
}

// New creates a new SQLite database instance
func New(config *models.Config) (*SQLite, error) {
	return &SQLite{
		config: config,
	}, nil
}

// Connect establishes connection to SQLite
func (s *SQLite) Connect(ctx context.Context) error {
	dbPath := s.config.URI
	if strings.HasPrefix(dbPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		dbPath = filepath.Join(home, dbPath[1:])
	} else if !filepath.IsAbs(dbPath) {
		absPath, err := filepath.Abs(dbPath)
		if err != nil {
			return fmt.Errorf("failed to resolve absolute path: %w", err)
		}
		dbPath = absPath
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open SQLite database at path '%s': %w", dbPath, err)
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping SQLite database at path '%s': %w", dbPath, err)
	}

	s.db = db

	return nil
}

// Disconnect closes the SQLite connection
func (s *SQLite) Disconnect(ctx context.Context) error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Ping checks the database connection
func (s *SQLite) Ping(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("not connected to database")
	}
	return s.db.PingContext(ctx)
}

func mapToJSON(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}
	result := "{"
	first := true
	for k, v := range m {
		if !first {
			result += ","
		}
		result += fmt.Sprintf(`"%s":"%s"`, k, v)
		first = false
	}
	result += "}"
	return result
}

func jsonToMap(jsonStr string) map[string]string {
	if jsonStr == "" || jsonStr == "{}" {
		return make(map[string]string)
	}
	return make(map[string]string)
}

func sliceToJSON(slice []string) string {
	if len(slice) == 0 {
		return "[]"
	}
	result := "["
	for i, s := range slice {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf(`"%s"`, s)
	}
	result += "]"
	return result
}

func jsonToSlice(jsonStr string) []string {
	if jsonStr == "" || jsonStr == "[]" {
		return []string{}
	}

	jsonStr = strings.TrimSpace(jsonStr)
	if !strings.HasPrefix(jsonStr, "[") || !strings.HasSuffix(jsonStr, "]") {
		return []string{}
	}

	jsonStr = jsonStr[1 : len(jsonStr)-1]
	if jsonStr == "" {
		return []string{}
	}

	parts := strings.Split(jsonStr, ",")
	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
			part = part[1 : len(part)-1]
		}
		result = append(result, part)
	}

	return result
}

// CreateLLM creates a new LLM configuration
func (s *SQLite) CreateLLM(ctx context.Context, llm *models.LLMConfig) error {
	llm.CreatedAt = time.Now()
	llm.UpdatedAt = time.Now()

	query := `
		INSERT INTO llms (id, name, provider, model, api_key, base_url, config, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query,
		llm.ID,
		llm.Name,
		llm.Provider,
		llm.Model,
		llm.APIKey,
		llm.BaseURL,
		mapToJSON(llm.Config),
		llm.Enabled,
		llm.CreatedAt,
		llm.UpdatedAt,
	)

	return err
}

// GetLLM retrieves an LLM configuration by ID
func (s *SQLite) GetLLM(ctx context.Context, id string) (*models.LLMConfig, error) {
	query := `
		SELECT id, name, provider, model, api_key, base_url, config, enabled, created_at, updated_at
		FROM llms WHERE id = ?`

	var llm models.LLMConfig
	var configJSON string

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&llm.ID,
		&llm.Name,
		&llm.Provider,
		&llm.Model,
		&llm.APIKey,
		&llm.BaseURL,
		&configJSON,
		&llm.Enabled,
		&llm.CreatedAt,
		&llm.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("LLM not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	llm.Config = jsonToMap(configJSON)
	return &llm, nil
}

// ListLLMs lists all LLM configurations, optionally filtered by enabled status
func (s *SQLite) ListLLMs(ctx context.Context, enabled *bool) ([]*models.LLMConfig, error) {
	query := `
		SELECT id, name, provider, model, api_key, base_url, config, enabled, created_at, updated_at
		FROM llms`
	args := []interface{}{}

	if enabled != nil {
		query += " WHERE enabled = ?"
		args = append(args, *enabled)
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var llms []*models.LLMConfig
	for rows.Next() {
		var llm models.LLMConfig
		var configJSON string

		err := rows.Scan(
			&llm.ID,
			&llm.Name,
			&llm.Provider,
			&llm.Model,
			&llm.APIKey,
			&llm.BaseURL,
			&configJSON,
			&llm.Enabled,
			&llm.CreatedAt,
			&llm.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		llm.Config = jsonToMap(configJSON)
		llms = append(llms, &llm)
	}

	return llms, nil
}

// UpdateLLM updates an existing LLM configuration
func (s *SQLite) UpdateLLM(ctx context.Context, llm *models.LLMConfig) error {
	llm.UpdatedAt = time.Now()

	query := `
		UPDATE llms 
		SET name = ?, provider = ?, model = ?, api_key = ?, base_url = ?, config = ?, enabled = ?, updated_at = ?
		WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query,
		llm.Name,
		llm.Provider,
		llm.Model,
		llm.APIKey,
		llm.BaseURL,
		mapToJSON(llm.Config),
		llm.Enabled,
		llm.UpdatedAt,
		llm.ID,
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

// DeleteLLM deletes an LLM configuration
func (s *SQLite) DeleteLLM(ctx context.Context, id string) error {
	query := "DELETE FROM llms WHERE id = ?"
	result, err := s.db.ExecContext(ctx, query, id)
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

// DeleteAllLLMs deletes all LLM configurations
func (s *SQLite) DeleteAllLLMs(ctx context.Context) (int, error) {
	query := "DELETE FROM llms"
	result, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}

// CreateSchedule creates a new schedule
func (s *SQLite) CreateSchedule(ctx context.Context, schedule *models.Schedule) error {
	schedule.CreatedAt = time.Now()
	schedule.UpdatedAt = time.Now()

	query := `
		INSERT INTO schedules (id, name, prompt_ids, llm_ids, cron_expr, temperature, enabled, last_run, next_run, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query,
		schedule.ID,
		schedule.Name,
		sliceToJSON(schedule.PromptIDs),
		sliceToJSON(schedule.LLMIDs),
		schedule.CronExpr,
		schedule.Temperature,
		schedule.Enabled,
		schedule.LastRun,
		schedule.NextRun,
		schedule.CreatedAt,
		schedule.UpdatedAt,
	)

	return err
}

// GetSchedule retrieves a schedule by ID
func (s *SQLite) GetSchedule(ctx context.Context, id string) (*models.Schedule, error) {
	query := `
		SELECT id, name, prompt_ids, llm_ids, cron_expr, temperature, enabled, last_run, next_run, created_at, updated_at
		FROM schedules WHERE id = ?`

	var schedule models.Schedule
	var promptIDsJSON, llmIDsJSON string

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&schedule.ID,
		&schedule.Name,
		&promptIDsJSON,
		&llmIDsJSON,
		&schedule.CronExpr,
		&schedule.Temperature,
		&schedule.Enabled,
		&schedule.LastRun,
		&schedule.NextRun,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("schedule not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	schedule.PromptIDs = jsonToSlice(promptIDsJSON)
	schedule.LLMIDs = jsonToSlice(llmIDsJSON)
	return &schedule, nil
}

// ListSchedules lists all schedules, optionally filtered by enabled status
func (s *SQLite) ListSchedules(ctx context.Context, enabled *bool) ([]*models.Schedule, error) {
	query := `
		SELECT id, name, prompt_ids, llm_ids, cron_expr, temperature, enabled, last_run, next_run, created_at, updated_at
		FROM schedules`
	args := []interface{}{}

	if enabled != nil {
		query += " WHERE enabled = ?"
		args = append(args, *enabled)
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.Schedule
	for rows.Next() {
		var schedule models.Schedule
		var promptIDsJSON, llmIDsJSON string

		err := rows.Scan(
			&schedule.ID,
			&schedule.Name,
			&promptIDsJSON,
			&llmIDsJSON,
			&schedule.CronExpr,
			&schedule.Temperature,
			&schedule.Enabled,
			&schedule.LastRun,
			&schedule.NextRun,
			&schedule.CreatedAt,
			&schedule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		schedule.PromptIDs = jsonToSlice(promptIDsJSON)
		schedule.LLMIDs = jsonToSlice(llmIDsJSON)
		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

// UpdateSchedule updates an existing schedule
func (s *SQLite) UpdateSchedule(ctx context.Context, schedule *models.Schedule) error {
	schedule.UpdatedAt = time.Now()

	query := `
		UPDATE schedules 
		SET name = ?, prompt_ids = ?, llm_ids = ?, cron_expr = ?, temperature = ?, enabled = ?, last_run = ?, next_run = ?, updated_at = ?
		WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query,
		schedule.Name,
		sliceToJSON(schedule.PromptIDs),
		sliceToJSON(schedule.LLMIDs),
		schedule.CronExpr,
		schedule.Temperature,
		schedule.Enabled,
		schedule.LastRun,
		schedule.NextRun,
		schedule.UpdatedAt,
		schedule.ID,
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

// DeleteSchedule deletes a schedule
func (s *SQLite) DeleteSchedule(ctx context.Context, id string) error {
	query := "DELETE FROM schedules WHERE id = ?"
	result, err := s.db.ExecContext(ctx, query, id)
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

// DeleteAllSchedules deletes all schedules
func (s *SQLite) DeleteAllSchedules(ctx context.Context) (int, error) {
	query := "DELETE FROM schedules"
	result, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}
