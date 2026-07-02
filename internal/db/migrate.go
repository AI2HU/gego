package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func ResolveMigrationsDir(baseDir, driver string) (string, error) {
	if baseDir == "" {
		baseDir = "/migrations"
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			workDir, _ := os.Getwd()
			baseDir = filepath.Join(workDir, "internal", "db", "migrations")
		}
	}

	subdir := driver
	switch driver {
	case "postgres":
		subdir = "postgres"
	case "sqlite3":
		subdir = "sqlite"
	default:
		return "", fmt.Errorf("unsupported migration driver: %s", driver)
	}

	absPath, err := filepath.Abs(filepath.Join(baseDir, subdir))
	if err != nil {
		return "", fmt.Errorf("failed to resolve migrations path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("migrations directory not found: %s", absPath)
	}

	return absPath, nil
}

func RunMigrations(ctx context.Context, db *sql.DB, driver string, migrationsDir string) error {
	absPath, err := ResolveMigrationsDir(migrationsDir, driver)
	if err != nil {
		return err
	}

	sourceURL := fmt.Sprintf("file://%s", absPath)
	if !strings.HasPrefix(absPath, "/") {
		sourceURL = fmt.Sprintf("file:///%s", absPath)
	}

	var m *migrate.Migrate

	switch driver {
	case "postgres":
		driverInstance, err := postgres.WithInstance(db, &postgres.Config{})
		if err != nil {
			return fmt.Errorf("failed to create postgres driver: %w", err)
		}
		m, err = migrate.NewWithDatabaseInstance(sourceURL, "postgres", driverInstance)
		if err != nil {
			return fmt.Errorf("failed to create migrate instance: %w", err)
		}
	case "sqlite3":
		driverInstance, err := sqlite3.WithInstance(db, &sqlite3.Config{})
		if err != nil {
			return fmt.Errorf("failed to create sqlite driver: %w", err)
		}
		m, err = migrate.NewWithDatabaseInstance(sourceURL, "sqlite3", driverInstance)
		if err != nil {
			return fmt.Errorf("failed to create migrate instance: %w", err)
		}
	default:
		return fmt.Errorf("unsupported migration driver: %s", driver)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
