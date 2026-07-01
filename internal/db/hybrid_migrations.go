package db

import (
	"context"
	"fmt"
	"os"
)

func RunHybridMigrations(ctx context.Context, database Database) error {
	hybridDB, ok := database.(*HybridDB)
	if !ok {
		return fmt.Errorf("database is not a HybridDB instance")
	}

	sqlBackend := hybridDB.GetSQLBackend()
	if sqlBackend == nil {
		return fmt.Errorf("SQL database not available")
	}

	sqlDB := sqlBackend.GetDB()
	if sqlDB == nil {
		return fmt.Errorf("database connection not available")
	}

	migrationsDir := os.Getenv("GEGO_MIGRATIONS_DIR")
	return RunMigrations(ctx, sqlDB, sqlBackend.DriverName(), migrationsDir)
}
