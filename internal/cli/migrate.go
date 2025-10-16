package cli

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/AI2HU/gego/internal/config"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Manage database migrations",
	Long:  `Run database migrations using gomigrate.`,
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Run all pending migrations",
	Long:  `Apply all pending database migrations.`,
	RunE:  runMigrateUp,
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  `Show the current migration status and version.`,
	RunE:  runMigrateStatus,
}

var migrateVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show current migration version",
	Long:  `Show the current database migration version.`,
	RunE:  runMigrateVersion,
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateVersionCmd)
}

func runMigrateUp(cmd *cobra.Command, args []string) error {
	fmt.Println("🔄 Running database migrations...")

	// Get SQLite path from config or use default
	sqlitePath := "gego.db" // Default path

	// Try to load from config if it exists
	if configPath := config.GetConfigPath(); config.Exists(configPath) {
		if cfg, err := config.Load(configPath); err == nil {
			sqlitePath = cfg.SQLDatabase.URI
		}
	}

	if err := runMigrations(sqlitePath); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("✅ Migrations completed successfully!")
	return nil
}

func runMigrateStatus(cmd *cobra.Command, args []string) error {
	fmt.Println("📊 Migration Status")
	fmt.Println("===================")

	// Get SQLite path from config or use default
	sqlitePath := "gego.db"
	if configPath := config.GetConfigPath(); config.Exists(configPath) {
		if cfg, err := config.Load(configPath); err == nil {
			sqlitePath = cfg.SQLDatabase.URI
		}
	}

	// Check if migrate command is available
	if _, err := exec.LookPath("migrate"); err != nil {
		return fmt.Errorf("migrate command not found. Please install golang-migrate: https://github.com/golang-migrate/migrate")
	}

	// Get migrations directory path
	migrationsDir := filepath.Join("internal", "db", "migrations")

	// Convert relative SQLite path to absolute path
	absSQLitePath, err := filepath.Abs(sqlitePath)
	if err != nil {
		return fmt.Errorf("failed to resolve SQLite path: %w", err)
	}

	// Construct database URL for SQLite
	dbURL := fmt.Sprintf("sqlite3://%s", absSQLitePath)

	// Run migrate version command
	cmdExec := exec.Command("migrate",
		"-path", migrationsDir,
		"-database", dbURL,
		"version")

	output, err := cmdExec.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("Current migration version: %s", string(output))
	return nil
}

func runMigrateVersion(cmd *cobra.Command, args []string) error {
	return runMigrateStatus(cmd, args)
}
