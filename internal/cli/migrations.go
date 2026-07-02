package cli

import (
	"context"

	"github.com/AI2HU/gego/internal/db"
)

func runDatabaseMigrations(ctx context.Context, database db.Database) error {
	return db.RunHybridMigrations(ctx, database)
}
