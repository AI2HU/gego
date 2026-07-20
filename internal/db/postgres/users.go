package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/AI2HU/gego/internal/models"
)

func (p *Postgres) CreateUser(ctx context.Context, user *models.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (id, username, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := p.db.ExecContext(ctx, query,
		user.ID, user.Username, user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (p *Postgres) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM users WHERE username = $1`

	var user models.User
	err := p.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", username)
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (p *Postgres) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM users WHERE id = $1`

	var user models.User
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (p *Postgres) ListUsers(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM users
		ORDER BY created_at ASC`

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, rows.Err()
}

func (p *Postgres) UpdateUser(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET username = $1, password_hash = $2, role = $3, updated_at = $4
		WHERE id = $5`

	result, err := p.db.ExecContext(ctx, query,
		user.Username, user.PasswordHash, user.Role, user.UpdatedAt, user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found: %s", user.ID)
	}
	return nil
}

func (p *Postgres) DeleteUser(ctx context.Context, id string) error {
	result, err := p.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found: %s", id)
	}
	return nil
}

func (p *Postgres) CountUsersByRole(ctx context.Context, role models.Role) (int, error) {
	var count int
	err := p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role = $1`, role).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users by role: %w", err)
	}
	return count, nil
}
