package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/AI2HU/gego/internal/db/sqlutil"
	"github.com/AI2HU/gego/internal/models"
)

func (p *Postgres) CreateSession(ctx context.Context, session *models.UserSession) error {
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()

	query := `
		INSERT INTO user_sessions (id, user_id, token_hash, expires_at, revoked_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := p.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.TokenHash, session.ExpiresAt,
		sqlutil.NullableTime(session.RevokedAt), session.CreatedAt, session.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (p *Postgres) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.UserSession, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at, updated_at
		FROM user_sessions WHERE token_hash = $1`

	var session models.UserSession
	var revokedAt sql.NullTime
	err := p.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&session.ID, &session.UserID, &session.TokenHash, &session.ExpiresAt,
		&revokedAt, &session.CreatedAt, &session.UpdatedAt,
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

func (p *Postgres) RevokeSession(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE user_sessions
		SET revoked_at = $1, updated_at = $2
		WHERE id = $3 AND revoked_at IS NULL`

	result, err := p.db.ExecContext(ctx, query, now, now, id)
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
