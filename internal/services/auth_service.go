package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/AI2HU/gego/internal/auth"
	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidSession     = errors.New("invalid session")
)

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

func (s *AuthService) Config() auth.Config {
	return s.config
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*models.AuthSessionResult, error) {
	user, err := s.db.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !auth.CheckPassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}

	return s.issueSession(ctx, user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*models.AuthSessionResult, error) {
	if refreshToken == "" {
		return nil, ErrInvalidSession
	}

	session, err := s.db.GetSessionByTokenHash(ctx, auth.HashRefreshToken(refreshToken))
	if err != nil {
		return nil, ErrInvalidSession
	}

	if session.RevokedAt != nil || time.Now().After(session.ExpiresAt) {
		return nil, ErrInvalidSession
	}

	user, err := s.db.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, ErrInvalidSession
	}

	if err := s.db.RevokeSession(ctx, session.ID); err != nil {
		return nil, fmt.Errorf("failed to rotate session: %w", err)
	}

	return s.issueSession(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	session, err := s.db.GetSessionByTokenHash(ctx, auth.HashRefreshToken(refreshToken))
	if err != nil {
		return nil
	}

	if err := s.db.RevokeSession(ctx, session.ID); err != nil {
		return nil
	}

	return nil
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

func (s *AuthService) issueSession(ctx context.Context, user *models.User) (*models.AuthSessionResult, error) {
	accessToken, expiresIn, err := auth.SignAccessToken(s.config, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	session := &models.UserSession{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		TokenHash: auth.HashRefreshToken(refreshToken),
		ExpiresAt: time.Now().Add(s.config.RefreshTokenTTL),
	}

	if err := s.db.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &models.AuthSessionResult{
		LoginResponse: models.LoginResponse{
			AccessToken: accessToken,
			TokenType:   "Bearer",
			ExpiresIn:   expiresIn,
			User:        models.ToAuthUserResponse(user),
		},
		RefreshToken: refreshToken,
	}, nil
}
