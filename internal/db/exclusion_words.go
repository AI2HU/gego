package db

import (
	"context"
	"errors"

	"github.com/AI2HU/gego/internal/models"
)

// ErrExclusionWordsUnsupported is returned when exclusion word operations are used with legacy SQLite.
var ErrExclusionWordsUnsupported = errors.New("exclusion words require PostgreSQL; run `gego db upgrade-from-sqlite`")

// ExclusionWordDatabase defines exclusion word operations for PostgreSQL-backed deployments.
// Implement in internal/db/postgres/ only; legacy SQLite does not support this interface.
type ExclusionWordDatabase interface {
	CreateExclusionWord(ctx context.Context, word *models.ExclusionWord) error
	GetExclusionWord(ctx context.Context, id string) (*models.ExclusionWord, error)
	GetExclusionWordByWord(ctx context.Context, word string) (*models.ExclusionWord, error)
	ListExclusionWords(ctx context.Context) ([]*models.ExclusionWord, error)
	DeleteExclusionWord(ctx context.Context, id string) error
	CountExclusionWords(ctx context.Context) (int, error)
}

func exclusionWordBackend(sqlDB SQLBackend) (ExclusionWordDatabase, error) {
	ew, ok := sqlDB.(ExclusionWordDatabase)
	if !ok {
		return nil, ErrExclusionWordsUnsupported
	}
	return ew, nil
}
