package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/AI2HU/gego/internal/db/sqlutil"
	"github.com/AI2HU/gego/internal/models"
)

func (s *SQLite) CreateSession(ctx context.Context, session *models.UserSession) error {
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()

	query := `
		INSERT INTO user_sessions (id, user_id, token_hash, expires_at, revoked_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.TokenHash,
		session.ExpiresAt,
		sqlutil.NullableTime(session.RevokedAt),
		session.CreatedAt,
		session.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func (s *SQLite) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.UserSession, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at, updated_at
		FROM user_sessions WHERE token_hash = ?`

	var session models.UserSession
	var revokedAt sql.NullTime
	err := s.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.ExpiresAt,
		&revokedAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, err
	}

	session.RevokedAt = sqlutil.ScanNullableTime(revokedAt)
	return &session, nil
}

func (s *SQLite) RevokeSession(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE user_sessions
		SET revoked_at = ?, updated_at = ?
		WHERE id = ? AND revoked_at IS NULL`

	result, err := s.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("session not found or already revoked: %s", id)
	}

	return nil
}

func (s *SQLite) RevokeSessionsByUserID(ctx context.Context, userID string) error {
	now := time.Now()
	query := `
		UPDATE user_sessions
		SET revoked_at = ?, updated_at = ?
		WHERE user_id = ? AND revoked_at IS NULL`

	_, err := s.db.ExecContext(ctx, query, now, now, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user sessions: %w", err)
	}
	return nil
}
