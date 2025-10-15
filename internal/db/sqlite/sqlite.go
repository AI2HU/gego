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
	// Expand the URI path (handle ~ and relative paths)
	dbPath := s.config.URI
	if strings.HasPrefix(dbPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		dbPath = filepath.Join(home, dbPath[1:])
	} else if !filepath.IsAbs(dbPath) {
		// Convert relative path to absolute
		absPath, err := filepath.Abs(dbPath)
		if err != nil {
			return fmt.Errorf("failed to resolve absolute path: %w", err)
		}
		dbPath = absPath
	}

	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open SQLite database at path '%s': %w", dbPath, err)
	}

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping SQLite database at path '%s': %w", dbPath, err)
	}

	s.db = db

	// Create tables
	if err := s.createTables(ctx); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// Run migrations
	if err := s.runMigrations(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

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

// createTables creates necessary tables
func (s *SQLite) createTables(ctx context.Context) error {
	// Create LLMs table
	createLLMsTable := `
	CREATE TABLE IF NOT EXISTS llms (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		provider TEXT NOT NULL,
		model TEXT NOT NULL,
		api_key TEXT,
		base_url TEXT,
		config TEXT, -- JSON string for additional config
		enabled BOOLEAN NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`

	// Create Schedules table
	createSchedulesTable := `
	CREATE TABLE IF NOT EXISTS schedules (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		prompt_ids TEXT NOT NULL, -- JSON array of prompt IDs
		llm_ids TEXT NOT NULL,    -- JSON array of LLM IDs
		cron_expr TEXT NOT NULL,
		temperature REAL DEFAULT 0.7,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		last_run DATETIME,
		next_run DATETIME,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`

	// Create indexes
	createIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_llms_provider ON llms(provider);",
		"CREATE INDEX IF NOT EXISTS idx_llms_enabled ON llms(enabled);",
		"CREATE INDEX IF NOT EXISTS idx_schedules_enabled ON schedules(enabled);",
		"CREATE INDEX IF NOT EXISTS idx_schedules_next_run ON schedules(next_run);",
	}

	queries := []string{createLLMsTable, createSchedulesTable}
	queries = append(queries, createIndexes...)

	for _, query := range queries {
		if _, err := s.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// runMigrations runs database migrations
func (s *SQLite) runMigrations(ctx context.Context) error {
	// Migration 1: Add temperature column to schedules table if it doesn't exist
	query := `ALTER TABLE schedules ADD COLUMN temperature REAL DEFAULT 0.7;`
	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		// Column might already exist, which is fine
		// In SQLite, ALTER TABLE ADD COLUMN fails if column already exists
		// We can ignore this error
	}

	return nil
}

// Helper function to convert map to JSON string
func mapToJSON(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}
	// Simple JSON conversion for map[string]string
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

// Helper function to parse JSON string to map
func jsonToMap(jsonStr string) map[string]string {
	// Simple JSON parsing for map[string]string
	// This is a basic implementation - in production, use proper JSON library
	if jsonStr == "" || jsonStr == "{}" {
		return make(map[string]string)
	}
	// For now, return empty map - proper JSON parsing would be needed
	return make(map[string]string)
}

// Helper function to convert string slice to JSON array
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

// Helper function to parse JSON array to string slice
func jsonToSlice(jsonStr string) []string {
	// Simple JSON parsing for []string
	// This is a basic implementation - in production, use proper JSON library
	if jsonStr == "" || jsonStr == "[]" {
		return []string{}
	}

	// Remove brackets and split by comma
	jsonStr = strings.TrimSpace(jsonStr)
	if !strings.HasPrefix(jsonStr, "[") || !strings.HasSuffix(jsonStr, "]") {
		return []string{}
	}

	// Remove brackets
	jsonStr = jsonStr[1 : len(jsonStr)-1]
	if jsonStr == "" {
		return []string{}
	}

	// Split by comma and clean up quotes
	parts := strings.Split(jsonStr, ",")
	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// Remove quotes if present
		if strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
			part = part[1 : len(part)-1]
		}
		result = append(result, part)
	}

	return result
}

// LLM Operations

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

// Schedule Operations

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
