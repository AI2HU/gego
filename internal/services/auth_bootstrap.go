package services

import (
	"context"
	"fmt"
	"os"

	"github.com/AI2HU/gego/internal/auth"
	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
)

const (
	EnvBootstrapAdminUsername = "GEGO_BOOTSTRAP_ADMIN_USERNAME"
	EnvBootstrapAdminPassword = "GEGO_BOOTSTRAP_ADMIN_PASSWORD"
)

func BootstrapAdminFromEnv(ctx context.Context, database db.Database) (*models.User, error) {
	users, err := database.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing users: %w", err)
	}
	if len(users) > 0 {
		return nil, nil
	}

	username := os.Getenv(EnvBootstrapAdminUsername)
	password := os.Getenv(EnvBootstrapAdminPassword)

	if username == "" {
		username = "admin"
	}

	if password == "" {
		return nil, fmt.Errorf(
			"no API users exist: set %s (required) and optionally %s (defaults to admin), or run: gego user create --username <name> --role admin",
			EnvBootstrapAdminPassword,
			EnvBootstrapAdminUsername,
		)
	}

	authService := NewAuthService(database, auth.Config{})
	user, err := authService.CreateUser(ctx, username, password, models.RoleAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap admin user: %w", err)
	}

	return user, nil
}
