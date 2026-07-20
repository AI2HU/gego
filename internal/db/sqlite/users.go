package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/AI2HU/gego/internal/models"
)

func (s *SQLite) CreateUser(ctx context.Context, user *models.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (id, username, password_hash, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *SQLite) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM users WHERE username = ?`

	var user models.User
	err := s.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", username)
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *SQLite) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM users WHERE id = ?`

	var user models.User
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *SQLite) ListUsers(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM users
		ORDER BY created_at ASC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

func (s *SQLite) UpdateUser(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET username = ?, password_hash = ?, role = ?, updated_at = ?
		WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.UpdatedAt,
		user.ID,
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

func (s *SQLite) DeleteUser(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
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

func (s *SQLite) CountUsersByRole(ctx context.Context, role models.Role) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role = ?`, role).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users by role: %w", err)
	}
	return count, nil
}
