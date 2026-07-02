package upgrade

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"

	"github.com/AI2HU/gego/internal/config"
	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/db/sqlutil"
	"github.com/AI2HU/gego/internal/models"
)

type Options struct {
	SQLitePath  string
	PostgresURI string
	ConfigPath  string
	MigrationsDir string
	DryRun      bool
	Force       bool
}

type Result struct {
	Message         string
	RestartRequired bool
	Counts          map[string]int
}

func ResolveSQLitePath(cfg *config.Config) string {
	if cfg != nil && cfg.SQLDatabase.URI != "" && cfg.SQLDatabase.Provider == "sqlite" {
		return resolveDBPath(cfg.SQLDatabase.URI)
	}

	if dataPath := strings.TrimSpace(os.Getenv("GEGO_DATA_PATH")); dataPath != "" {
		return filepath.Join(dataPath, "gego.db")
	}

	home, err := os.UserHomeDir()
	if err == nil {
		return filepath.Join(home, ".gego", "gego.db")
	}

	return "gego.db"
}

func resolveDBPath(dbPath string) string {
	if strings.HasPrefix(dbPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return dbPath
		}
		dbPath = filepath.Join(home, dbPath[1:])
	}
	if !filepath.IsAbs(dbPath) {
		if absPath, err := filepath.Abs(dbPath); err == nil {
			return absPath
		}
	}
	return dbPath
}

func ResolvePostgresURI(cfg *config.Config) string {
	if uri := strings.TrimSpace(os.Getenv("GEGO_POSTGRES_URI")); uri != "" {
		return uri
	}
	if uri := strings.TrimSpace(os.Getenv("DATABASE_URL")); uri != "" {
		return uri
	}
	if cfg != nil && cfg.SQLDatabase.Provider == "postgres" && cfg.SQLDatabase.URI != "" {
		return cfg.SQLDatabase.URI
	}
	return ""
}

func SQLiteUpgradeRequired(cfg *config.Config) (bool, error) {
	if cfg == nil || cfg.SQLDatabase.Provider != "sqlite" {
		return false, nil
	}
	path := ResolveSQLitePath(cfg)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func RunSQLiteToPostgres(ctx context.Context, cfg *config.Config, opts Options) (*Result, error) {
	sqlitePath := opts.SQLitePath
	if sqlitePath == "" {
		sqlitePath = ResolveSQLitePath(cfg)
	}

	postgresURI := opts.PostgresURI
	if postgresURI == "" {
		postgresURI = ResolvePostgresURI(cfg)
	}
	if postgresURI == "" {
		return nil, fmt.Errorf("postgres URI is required: set GEGO_POSTGRES_URI or pass --postgres-uri")
	}

	if _, err := os.Stat(sqlitePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("sqlite database not found at %s", sqlitePath)
	} else if err != nil {
		return nil, err
	}

	migrationsDir := opts.MigrationsDir
	if migrationsDir == "" {
		migrationsDir = os.Getenv("GEGO_MIGRATIONS_DIR")
		if migrationsDir == "" {
			migrationsDir = "/migrations"
			if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
				workDir, _ := os.Getwd()
				migrationsDir = filepath.Join(workDir, "internal", "db", "migrations")
			}
		}
	}

	sqliteDB, err := openSQLite(ctx, sqlitePath)
	if err != nil {
		return nil, err
	}
	defer sqliteDB.Close()

	if err := db.RunMigrations(ctx, sqliteDB, "sqlite3", migrationsDir); err != nil {
		return nil, fmt.Errorf("sqlite migrations: %w", err)
	}

	pgDB, err := openPostgres(ctx, postgresURI)
	if err != nil {
		return nil, err
	}
	defer pgDB.Close()

	if err := db.RunMigrations(ctx, pgDB, "postgres", migrationsDir); err != nil {
		return nil, fmt.Errorf("postgres migrations: %w", err)
	}

	counts, err := readSQLiteCounts(ctx, sqliteDB)
	if err != nil {
		return nil, err
	}

	if opts.DryRun {
		return &Result{
			Message: fmt.Sprintf("dry run: would migrate %d users, %d llms, %d schedules, %d sessions",
				counts["users"], counts["llms"], counts["schedules"], counts["user_sessions"]),
			RestartRequired: false,
			Counts:          counts,
		}, nil
	}

	hasData, err := postgresHasData(ctx, pgDB)
	if err != nil {
		return nil, err
	}
	if hasData && !opts.Force {
		return nil, fmt.Errorf("postgres database is not empty; use --force to overwrite")
	}

	tx, err := pgDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if opts.Force {
		if _, err := tx.ExecContext(ctx, "TRUNCATE user_sessions, schedules, llms, users RESTART IDENTITY CASCADE"); err != nil {
			return nil, fmt.Errorf("truncate postgres tables: %w", err)
		}
	}

	if err := copyUsers(ctx, sqliteDB, tx); err != nil {
		return nil, err
	}
	if err := copyLLMs(ctx, sqliteDB, tx); err != nil {
		return nil, err
	}
	if err := copySchedules(ctx, sqliteDB, tx); err != nil {
		return nil, err
	}
	if err := copySessions(ctx, sqliteDB, tx); err != nil {
		return nil, err
	}

	if err := verifyCounts(ctx, sqliteDB, tx); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	backupPath := fmt.Sprintf("%s.bak.%d", sqlitePath, time.Now().Unix())
	if err := os.Rename(sqlitePath, backupPath); err != nil {
		return nil, fmt.Errorf("backup sqlite file: %w", err)
	}

	configPath := opts.ConfigPath
	if configPath == "" {
		configPath = config.GetConfigPath()
	}
	if err := updateConfig(configPath, postgresURI); err != nil {
		return nil, fmt.Errorf("update config: %w", err)
	}

	return &Result{
		Message: fmt.Sprintf("migrated %d users, %d llms, %d schedules, %d sessions; sqlite backed up to %s",
			counts["users"], counts["llms"], counts["schedules"], counts["user_sessions"], backupPath),
		RestartRequired: true,
		Counts:          counts,
	}, nil
}

func openSQLite(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	return db, nil
}

func openPostgres(ctx context.Context, uri string) (*sql.DB, error) {
	db, err := sql.Open("pgx", uri)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return db, nil
}

func readSQLiteCounts(ctx context.Context, sqliteDB *sql.DB) (map[string]int, error) {
	counts := map[string]int{}
	for _, table := range []string{"users", "llms", "schedules", "user_sessions"} {
		var count int
		if err := sqliteDB.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count); err != nil {
			return nil, fmt.Errorf("count %s: %w", table, err)
		}
		counts[table] = count
	}
	return counts, nil
}

func postgresHasData(ctx context.Context, pgDB *sql.DB) (bool, error) {
	var count int
	if err := pgDB.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM users) +
			(SELECT COUNT(*) FROM llms) +
			(SELECT COUNT(*) FROM schedules) +
			(SELECT COUNT(*) FROM user_sessions)
	`).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func copyUsers(ctx context.Context, sqliteDB *sql.DB, tx *sql.Tx) error {
	rows, err := sqliteDB.QueryContext(ctx, `
		SELECT id, username, password_hash, role, created_at, updated_at FROM users`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO users (id, username, password_hash, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			user.ID, user.Username, user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt,
		); err != nil {
			return err
		}
	}
	return rows.Err()
}

func copyLLMs(ctx context.Context, sqliteDB *sql.DB, tx *sql.Tx) error {
	rows, err := sqliteDB.QueryContext(ctx, `
		SELECT id, name, provider, model, api_key, base_url, config, enabled, created_at, updated_at FROM llms`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var llm models.LLMConfig
		var configJSON string
		if err := rows.Scan(&llm.ID, &llm.Name, &llm.Provider, &llm.Model, &llm.APIKey, &llm.BaseURL,
			&configJSON, &llm.Enabled, &llm.CreatedAt, &llm.UpdatedAt); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO llms (id, name, provider, model, api_key, base_url, config, enabled, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10)`,
			llm.ID, llm.Name, llm.Provider, llm.Model, llm.APIKey, llm.BaseURL,
			sqlutil.MapToJSON(sqlutil.JSONToMap(configJSON)), llm.Enabled, llm.CreatedAt, llm.UpdatedAt,
		); err != nil {
			return err
		}
	}
	return rows.Err()
}

func copySchedules(ctx context.Context, sqliteDB *sql.DB, tx *sql.Tx) error {
	rows, err := sqliteDB.QueryContext(ctx, `
		SELECT id, name, prompt_ids, llm_ids, cron_expr, temperature, enabled, last_run, next_run, created_at, updated_at
		FROM schedules`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var schedule models.Schedule
		var promptIDsJSON, llmIDsJSON string
		if err := rows.Scan(&schedule.ID, &schedule.Name, &promptIDsJSON, &llmIDsJSON,
			&schedule.CronExpr, &schedule.Temperature, &schedule.Enabled,
			&schedule.LastRun, &schedule.NextRun, &schedule.CreatedAt, &schedule.UpdatedAt); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO schedules (id, name, prompt_ids, llm_ids, cron_expr, temperature, enabled, last_run, next_run, created_at, updated_at)
			VALUES ($1, $2, $3::jsonb, $4::jsonb, $5, $6, $7, $8, $9, $10, $11)`,
			schedule.ID, schedule.Name,
			sqlutil.SliceToJSON(sqlutil.JSONToSlice(promptIDsJSON)),
			sqlutil.SliceToJSON(sqlutil.JSONToSlice(llmIDsJSON)),
			schedule.CronExpr, schedule.Temperature, schedule.Enabled,
			schedule.LastRun, schedule.NextRun, schedule.CreatedAt, schedule.UpdatedAt,
		); err != nil {
			return err
		}
	}
	return rows.Err()
}

func copySessions(ctx context.Context, sqliteDB *sql.DB, tx *sql.Tx) error {
	rows, err := sqliteDB.QueryContext(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at, updated_at FROM user_sessions`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var session models.UserSession
		var revokedAt sql.NullTime
		if err := rows.Scan(&session.ID, &session.UserID, &session.TokenHash, &session.ExpiresAt,
			&revokedAt, &session.CreatedAt, &session.UpdatedAt); err != nil {
			return err
		}
		session.RevokedAt = sqlutil.ScanNullableTime(revokedAt)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO user_sessions (id, user_id, token_hash, expires_at, revoked_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			session.ID, session.UserID, session.TokenHash, session.ExpiresAt,
			sqlutil.NullableTime(session.RevokedAt), session.CreatedAt, session.UpdatedAt,
		); err != nil {
			return err
		}
	}
	return rows.Err()
}

func verifyCounts(ctx context.Context, sqliteDB *sql.DB, tx *sql.Tx) error {
	for _, table := range []string{"users", "llms", "schedules", "user_sessions"} {
		var sqliteCount, pgCount int
		if err := sqliteDB.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&sqliteCount); err != nil {
			return err
		}
		if err := tx.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&pgCount); err != nil {
			return err
		}
		if sqliteCount != pgCount {
			return fmt.Errorf("count mismatch for %s: sqlite=%d postgres=%d", table, sqliteCount, pgCount)
		}
	}
	return nil
}

func updateConfig(configPath, postgresURI string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		cfg = config.DefaultConfig()
	}
	cfg.SQLDatabase.Provider = "postgres"
	cfg.SQLDatabase.URI = postgresURI
	cfg.SQLDatabase.Database = "gego"
	return cfg.Save(configPath)
}
