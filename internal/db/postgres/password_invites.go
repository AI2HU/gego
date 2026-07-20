package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/AI2HU/gego/internal/db/sqlutil"
	"github.com/AI2HU/gego/internal/models"
)

func (p *Postgres) CreatePasswordInvite(ctx context.Context, invite *models.PasswordInvite) error {
	invite.CreatedAt = time.Now()

	query := `
		INSERT INTO password_invites (id, user_id, token_hash, expires_at, revoked_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := p.db.ExecContext(ctx, query,
		invite.ID, invite.UserID, invite.TokenHash, invite.ExpiresAt,
		sqlutil.NullableTime(invite.RevokedAt), invite.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create password invite: %w", err)
	}
	return nil
}

func (p *Postgres) GetPasswordInviteByTokenHash(ctx context.Context, tokenHash string) (*models.PasswordInvite, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM password_invites
		WHERE token_hash = $1`

	var invite models.PasswordInvite
	var revokedAt sql.NullTime
	err := p.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&invite.ID, &invite.UserID, &invite.TokenHash, &invite.ExpiresAt,
		&revokedAt, &invite.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invite not found")
	}
	if err != nil {
		return nil, err
	}

	invite.RevokedAt = sqlutil.ScanNullableTime(revokedAt)
	return &invite, nil
}

func (p *Postgres) RevokePasswordInvite(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE password_invites
		SET revoked_at = $1
		WHERE id = $2 AND revoked_at IS NULL`

	result, err := p.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to revoke password invite: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("invite not found or already revoked: %s", id)
	}
	return nil
}

func (p *Postgres) RevokePasswordInvitesByUserID(ctx context.Context, userID string) error {
	now := time.Now()
	query := `
		UPDATE password_invites
		SET revoked_at = $1
		WHERE user_id = $2 AND revoked_at IS NULL`

	_, err := p.db.ExecContext(ctx, query, now, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke password invites: %w", err)
	}
	return nil
}
