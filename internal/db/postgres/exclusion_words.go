package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/AI2HU/gego/internal/models"
)

func (p *Postgres) CreateExclusionWord(ctx context.Context, word *models.ExclusionWord) error {
	word.Word = strings.TrimSpace(word.Word)
	word.CreatedAt = time.Now()
	word.UpdatedAt = time.Now()

	query := `
		INSERT INTO exclusion_words (id, word, created_at, updated_at)
		VALUES ($1, $2, $3, $4)`

	_, err := p.db.ExecContext(ctx, query,
		word.ID, word.Word, word.CreatedAt, word.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create exclusion word: %w", err)
	}
	return nil
}

func (p *Postgres) GetExclusionWord(ctx context.Context, id string) (*models.ExclusionWord, error) {
	query := `
		SELECT id, word, created_at, updated_at
		FROM exclusion_words WHERE id = $1`

	var word models.ExclusionWord
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&word.ID, &word.Word, &word.CreatedAt, &word.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("exclusion word not found: %s", id)
	}
	if err != nil {
		return nil, err
	}
	return &word, nil
}

func (p *Postgres) GetExclusionWordByWord(ctx context.Context, word string) (*models.ExclusionWord, error) {
	query := `
		SELECT id, word, created_at, updated_at
		FROM exclusion_words WHERE LOWER(word) = LOWER($1)`

	var result models.ExclusionWord
	err := p.db.QueryRowContext(ctx, query, strings.TrimSpace(word)).Scan(
		&result.ID, &result.Word, &result.CreatedAt, &result.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (p *Postgres) ListExclusionWords(ctx context.Context) ([]*models.ExclusionWord, error) {
	query := `
		SELECT id, word, created_at, updated_at
		FROM exclusion_words
		ORDER BY LOWER(word) ASC`

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var words []*models.ExclusionWord
	for rows.Next() {
		var word models.ExclusionWord
		if err := rows.Scan(
			&word.ID, &word.Word, &word.CreatedAt, &word.UpdatedAt,
		); err != nil {
			return nil, err
		}
		words = append(words, &word)
	}
	return words, rows.Err()
}

func (p *Postgres) DeleteExclusionWord(ctx context.Context, id string) error {
	result, err := p.db.ExecContext(ctx, `DELETE FROM exclusion_words WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete exclusion word: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("exclusion word not found: %s", id)
	}
	return nil
}

func (p *Postgres) CountExclusionWords(ctx context.Context) (int, error) {
	var count int
	err := p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM exclusion_words`).Scan(&count)
	return count, err
}
