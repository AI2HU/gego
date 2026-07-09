package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/AI2HU/gego/internal/models"
)

func (p *Postgres) CreateBrand(ctx context.Context, brand *models.Brand) error {
	brand.Name = strings.TrimSpace(brand.Name)
	brand.CreatedAt = time.Now()
	brand.UpdatedAt = time.Now()

	query := `
		INSERT INTO brands (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)`

	_, err := p.db.ExecContext(ctx, query,
		brand.ID, brand.Name, brand.CreatedAt, brand.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create brand: %w", err)
	}
	return nil
}

func (p *Postgres) GetBrand(ctx context.Context, id string) (*models.Brand, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM brands WHERE id = $1`

	var brand models.Brand
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&brand.ID, &brand.Name, &brand.CreatedAt, &brand.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("brand not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	aliases, err := p.ListBrandAliasesByBrandID(ctx, id)
	if err != nil {
		return nil, err
	}
	brand.Aliases = aliases
	return &brand, nil
}

func (p *Postgres) GetBrandByName(ctx context.Context, name string) (*models.Brand, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM brands WHERE LOWER(name) = LOWER($1)`

	var brand models.Brand
	err := p.db.QueryRowContext(ctx, query, strings.TrimSpace(name)).Scan(
		&brand.ID, &brand.Name, &brand.CreatedAt, &brand.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	aliases, err := p.ListBrandAliasesByBrandID(ctx, brand.ID)
	if err != nil {
		return nil, err
	}
	brand.Aliases = aliases
	return &brand, nil
}

func (p *Postgres) ListBrands(ctx context.Context) ([]*models.Brand, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM brands
		ORDER BY LOWER(name) ASC`

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []*models.Brand
	for rows.Next() {
		var brand models.Brand
		if err := rows.Scan(
			&brand.ID, &brand.Name, &brand.CreatedAt, &brand.UpdatedAt,
		); err != nil {
			return nil, err
		}
		aliases, err := p.ListBrandAliasesByBrandID(ctx, brand.ID)
		if err != nil {
			return nil, err
		}
		brand.Aliases = aliases
		brands = append(brands, &brand)
	}
	return brands, rows.Err()
}

func (p *Postgres) UpdateBrand(ctx context.Context, brand *models.Brand) error {
	brand.Name = strings.TrimSpace(brand.Name)
	brand.UpdatedAt = time.Now()

	result, err := p.db.ExecContext(ctx, `
		UPDATE brands SET name = $1, updated_at = $2 WHERE id = $3`,
		brand.Name, brand.UpdatedAt, brand.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update brand: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("brand not found: %s", brand.ID)
	}
	return nil
}

func (p *Postgres) DeleteBrand(ctx context.Context, id string) error {
	result, err := p.db.ExecContext(ctx, `DELETE FROM brands WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete brand: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("brand not found: %s", id)
	}
	return nil
}

func (p *Postgres) CreateBrandAlias(ctx context.Context, alias *models.BrandAlias) error {
	alias.Alias = strings.TrimSpace(alias.Alias)
	alias.CreatedAt = time.Now()
	alias.UpdatedAt = time.Now()

	query := `
		INSERT INTO brand_aliases (id, brand_id, alias, case_sensitive, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := p.db.ExecContext(ctx, query,
		alias.ID, alias.BrandID, alias.Alias, alias.CaseSensitive,
		alias.CreatedAt, alias.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create brand alias: %w", err)
	}
	return nil
}

func (p *Postgres) GetBrandAlias(ctx context.Context, id string) (*models.BrandAlias, error) {
	query := `
		SELECT id, brand_id, alias, case_sensitive, created_at, updated_at
		FROM brand_aliases WHERE id = $1`

	var alias models.BrandAlias
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&alias.ID, &alias.BrandID, &alias.Alias, &alias.CaseSensitive,
		&alias.CreatedAt, &alias.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("brand alias not found: %s", id)
	}
	if err != nil {
		return nil, err
	}
	return &alias, nil
}

func (p *Postgres) GetBrandAliasByAlias(ctx context.Context, alias string) (*models.BrandAlias, error) {
	query := `
		SELECT id, brand_id, alias, case_sensitive, created_at, updated_at
		FROM brand_aliases WHERE LOWER(alias) = LOWER($1)`

	var result models.BrandAlias
	err := p.db.QueryRowContext(ctx, query, strings.TrimSpace(alias)).Scan(
		&result.ID, &result.BrandID, &result.Alias, &result.CaseSensitive,
		&result.CreatedAt, &result.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (p *Postgres) ListBrandAliasesByBrandID(ctx context.Context, brandID string) ([]*models.BrandAlias, error) {
	query := `
		SELECT id, brand_id, alias, case_sensitive, created_at, updated_at
		FROM brand_aliases
		WHERE brand_id = $1
		ORDER BY LOWER(alias) ASC`

	rows, err := p.db.QueryContext(ctx, query, brandID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aliases []*models.BrandAlias
	for rows.Next() {
		var alias models.BrandAlias
		if err := rows.Scan(
			&alias.ID, &alias.BrandID, &alias.Alias, &alias.CaseSensitive,
			&alias.CreatedAt, &alias.UpdatedAt,
		); err != nil {
			return nil, err
		}
		aliases = append(aliases, &alias)
	}
	return aliases, rows.Err()
}

func (p *Postgres) UpdateBrandAlias(ctx context.Context, alias *models.BrandAlias) error {
	alias.Alias = strings.TrimSpace(alias.Alias)
	alias.UpdatedAt = time.Now()

	result, err := p.db.ExecContext(ctx, `
		UPDATE brand_aliases
		SET alias = $1, case_sensitive = $2, updated_at = $3
		WHERE id = $4`,
		alias.Alias, alias.CaseSensitive, alias.UpdatedAt, alias.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update brand alias: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("brand alias not found: %s", alias.ID)
	}
	return nil
}

func (p *Postgres) DeleteBrandAlias(ctx context.Context, id string) error {
	result, err := p.db.ExecContext(ctx, `DELETE FROM brand_aliases WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete brand alias: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("brand alias not found: %s", id)
	}
	return nil
}

func (p *Postgres) DeleteAllBrands(ctx context.Context) (int, error) {
	result, err := p.db.ExecContext(ctx, `DELETE FROM brands`)
	if err != nil {
		return 0, fmt.Errorf("failed to delete all brands: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rows), nil
}
