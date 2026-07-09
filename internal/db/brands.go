package db

import (
	"context"
	"errors"

	"github.com/AI2HU/gego/internal/models"
)

// ErrBrandsUnsupported is returned when brand operations are used with legacy SQLite.
var ErrBrandsUnsupported = errors.New("brands require PostgreSQL; run `gego db upgrade-from-sqlite`")

// BrandDatabase defines brand mapping operations for PostgreSQL-backed deployments.
type BrandDatabase interface {
	CreateBrand(ctx context.Context, brand *models.Brand) error
	GetBrand(ctx context.Context, id string) (*models.Brand, error)
	GetBrandByName(ctx context.Context, name string) (*models.Brand, error)
	ListBrands(ctx context.Context) ([]*models.Brand, error)
	UpdateBrand(ctx context.Context, brand *models.Brand) error
	DeleteBrand(ctx context.Context, id string) error

	CreateBrandAlias(ctx context.Context, alias *models.BrandAlias) error
	GetBrandAlias(ctx context.Context, id string) (*models.BrandAlias, error)
	GetBrandAliasByAlias(ctx context.Context, alias string) (*models.BrandAlias, error)
	ListBrandAliasesByBrandID(ctx context.Context, brandID string) ([]*models.BrandAlias, error)
	UpdateBrandAlias(ctx context.Context, alias *models.BrandAlias) error
	DeleteBrandAlias(ctx context.Context, id string) error
	DeleteAllBrands(ctx context.Context) (int, error)
}

func brandBackend(sqlDB SQLBackend) (BrandDatabase, error) {
	b, ok := sqlDB.(BrandDatabase)
	if !ok {
		return nil, ErrBrandsUnsupported
	}
	return b, nil
}
