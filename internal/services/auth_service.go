package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/AI2HU/gego/internal/auth"
	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	db     db.Database
	config auth.Config
}

func NewAuthService(database db.Database, config auth.Config) *AuthService {
	return &AuthService{
		db:     database,
		config: config,
	}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*models.LoginResponse, error) {
	user, err := s.db.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !auth.CheckPassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}

	token, expiresIn, err := auth.SignAccessToken(s.config, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	return &models.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User:        models.ToAuthUserResponse(user),
	}, nil
}

func (s *AuthService) GetUserProfile(ctx context.Context, userID string) (*models.AuthUserResponse, error) {
	user, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	profile := models.ToAuthUserResponse(user)
	return &profile, nil
}

func (s *AuthService) CreateUser(ctx context.Context, username, password string, role models.Role) (*models.User, error) {
	if !role.Valid() {
		return nil, fmt.Errorf("invalid role: %s", role)
	}
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
	}

	if err := s.db.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) ListUsers(ctx context.Context) ([]*models.User, error) {
	return s.db.ListUsers(ctx)
}
