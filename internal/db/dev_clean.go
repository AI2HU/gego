package db

import (
	"context"
	"fmt"
)

func (h *HybridDB) CleanDev(ctx context.Context) error {
	if err := h.nosqlDB.CleanAll(ctx); err != nil {
		return fmt.Errorf("clean nosql database: %w", err)
	}
	if err := h.cleanSQL(ctx); err != nil {
		return fmt.Errorf("clean sql database: %w", err)
	}
	return nil
}

func (h *HybridDB) cleanSQL(ctx context.Context) error {
	type sqlCleaner interface {
		CleanAll(ctx context.Context) error
	}
	cleaner, ok := h.sqlDB.(sqlCleaner)
	if !ok {
		return fmt.Errorf("sql backend does not support dev clean")
	}
	return cleaner.CleanAll(ctx)
}
