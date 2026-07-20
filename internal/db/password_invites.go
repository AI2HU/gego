package db

import (
	"context"
	"errors"

	"github.com/AI2HU/gego/internal/models"
)

// ErrPasswordInvitesUnsupported is returned when password invites are used with legacy SQLite.
var ErrPasswordInvitesUnsupported = errors.New("password invites require PostgreSQL; run `gego db upgrade-from-sqlite`")

// PasswordInviteDatabase defines invite-token operations for PostgreSQL-backed deployments.
type PasswordInviteDatabase interface {
	CreatePasswordInvite(ctx context.Context, invite *models.PasswordInvite) error
	GetPasswordInviteByTokenHash(ctx context.Context, tokenHash string) (*models.PasswordInvite, error)
	RevokePasswordInvite(ctx context.Context, id string) error
	RevokePasswordInvitesByUserID(ctx context.Context, userID string) error
}

func passwordInviteBackend(sqlDB SQLBackend) (PasswordInviteDatabase, error) {
	p, ok := sqlDB.(PasswordInviteDatabase)
	if !ok {
		return nil, ErrPasswordInvitesUnsupported
	}
	return p, nil
}
