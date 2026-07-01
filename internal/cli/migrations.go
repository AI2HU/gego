package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AI2HU/gego/internal/db"
)

func runDatabaseMigrations(ctx context.Context, database db.Database) error {
	hybridDB, ok := database.(*db.HybridDB)
	if !ok {
		return fmt.Errorf("database is not a HybridDB instance")
	}

	sqlBackend := hybridDB.GetSQLBackend()
	if sqlBackend == nil {
		return fmt.Errorf("SQL database not available")
	}

	migrationsDir := os.Getenv("GEGO_MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "/migrations"
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			workDir, _ := os.Getwd()
			migrationsDir = filepath.Join(workDir, "internal", "db", "migrations")
		}
	}

	sqlDB := sqlBackend.GetDB()
	if sqlDB == nil {
		return fmt.Errorf("database connection not available")
	}

	return db.RunMigrations(ctx, sqlDB, sqlBackend.DriverName(), migrationsDir)
}
