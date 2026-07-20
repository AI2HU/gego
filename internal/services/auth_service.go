package services

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/AI2HU/gego/internal/auth"
	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
)

const InviteTokenTTL = 7 * 24 * time.Hour

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidSession     = errors.New("invalid session")
	ErrInvalidInvite      = errors.New("invalid or expired invite")
	ErrUserNotFound       = errors.New("user not found")
	ErrCannotDeleteSelf   = errors.New("cannot delete your own account")
	ErrLastAdmin          = errors.New("cannot remove the last admin")
	ErrPasswordNotSet     = errors.New("password has not been set")
)

type AuthService struct {
	db     db.Database
	config auth.Config
}

type InviteTokenResult struct {
	User  *models.User
	Token string
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
	trimmed := strings.TrimSpace(username)
	user, err := s.db.GetUserByUsername(ctx, normalizeEmail(trimmed))
	if err != nil {
		user, err = s.db.GetUserByUsername(ctx, trimmed)
		if err != nil {
			return nil, ErrInvalidCredentials
		}
	}

	if user.PasswordHash == "" {
		return nil, ErrPasswordNotSet
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

	if user.PasswordHash == "" {
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
		Username:     strings.TrimSpace(username),
		PasswordHash: passwordHash,
		Role:         role,
	}

	if err := s.db.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// CreateInvitedUser creates a user without a password and issues a set-password invite token.
func (s *AuthService) CreateInvitedUser(ctx context.Context, email string, role models.Role) (*InviteTokenResult, error) {
	if !role.Valid() {
		return nil, fmt.Errorf("invalid role: %s", role)
	}

	normalized, err := normalizeAndValidateEmail(email)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New().String(),
		Username:     normalized,
		PasswordHash: "",
		Role:         role,
	}

	if err := s.db.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	token, err := s.issuePasswordInvite(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &InviteTokenResult{User: user, Token: token}, nil
}

// CreatePasswordInviteForUser revokes prior invites and issues a new set-password token.
func (s *AuthService) CreatePasswordInviteForUser(ctx context.Context, userID string) (*InviteTokenResult, error) {
	user, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	token, err := s.issuePasswordInvite(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &InviteTokenResult{User: user, Token: token}, nil
}

func (s *AuthService) SetPasswordWithInvite(ctx context.Context, token, password string) (*models.AuthSessionResult, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrInvalidInvite
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	invite, err := s.db.GetPasswordInviteByTokenHash(ctx, auth.HashRefreshToken(token))
	if err != nil {
		return nil, ErrInvalidInvite
	}
	if invite.RevokedAt != nil || time.Now().After(invite.ExpiresAt) {
		return nil, ErrInvalidInvite
	}

	user, err := s.db.GetUserByID(ctx, invite.UserID)
	if err != nil {
		return nil, ErrInvalidInvite
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = passwordHash
	if err := s.db.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	_ = s.db.RevokePasswordInvite(ctx, invite.ID)
	_ = s.db.RevokePasswordInvitesByUserID(ctx, user.ID)
	_ = s.db.RevokeSessionsByUserID(ctx, user.ID)

	return s.issueSession(ctx, user)
}

func (s *AuthService) ListUsers(ctx context.Context) ([]*models.User, error) {
	return s.db.ListUsers(ctx)
}

func (s *AuthService) ListUserProfiles(ctx context.Context) ([]models.AuthUserResponse, error) {
	users, err := s.db.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	profiles := make([]models.AuthUserResponse, 0, len(users))
	for _, user := range users {
		profiles = append(profiles, models.ToAuthUserResponse(user))
	}
	return profiles, nil
}

func (s *AuthService) UpdateUser(ctx context.Context, userID string, role *models.Role, password *string) (*models.User, error) {
	user, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	changed := false
	revokeSessions := false

	if role != nil {
		if !role.Valid() {
			return nil, fmt.Errorf("invalid role: %s", *role)
		}
		if user.Role != *role {
			if user.Role == models.RoleAdmin && *role != models.RoleAdmin {
				if err := s.ensureNotLastAdmin(ctx); err != nil {
					return nil, err
				}
			}
			user.Role = *role
			changed = true
			revokeSessions = true
		}
	}

	if password != nil {
		if len(*password) < 8 {
			return nil, fmt.Errorf("password must be at least 8 characters")
		}
		passwordHash, err := auth.HashPassword(*password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = passwordHash
		changed = true
		revokeSessions = true
	}

	if !changed {
		return user, nil
	}

	if err := s.db.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	if revokeSessions {
		_ = s.db.RevokeSessionsByUserID(ctx, user.ID)
		_ = s.db.RevokePasswordInvitesByUserID(ctx, user.ID)
	}

	return user, nil
}

func (s *AuthService) DeleteUser(ctx context.Context, userID, actorID string) error {
	if userID == actorID {
		return ErrCannotDeleteSelf
	}

	user, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.Role == models.RoleAdmin {
		if err := s.ensureNotLastAdmin(ctx); err != nil {
			return err
		}
	}

	if err := s.db.DeleteUser(ctx, userID); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) ensureNotLastAdmin(ctx context.Context) error {
	count, err := s.db.CountUsersByRole(ctx, models.RoleAdmin)
	if err != nil {
		return err
	}
	if count <= 1 {
		return ErrLastAdmin
	}
	return nil
}

func (s *AuthService) issuePasswordInvite(ctx context.Context, userID string) (string, error) {
	if err := s.db.RevokePasswordInvitesByUserID(ctx, userID); err != nil {
		return "", err
	}

	token, err := auth.GenerateRefreshToken()
	if err != nil {
		return "", err
	}

	invite := &models.PasswordInvite{
		ID:        uuid.New().String(),
		UserID:    userID,
		TokenHash: auth.HashRefreshToken(token),
		ExpiresAt: time.Now().Add(InviteTokenTTL),
	}

	if err := s.db.CreatePasswordInvite(ctx, invite); err != nil {
		return "", fmt.Errorf("failed to create password invite: %w", err)
	}

	return token, nil
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

func normalizeAndValidateEmail(email string) (string, error) {
	normalized := normalizeEmail(email)
	if normalized == "" {
		return "", fmt.Errorf("email is required")
	}
	addr, err := mail.ParseAddress(normalized)
	if err != nil || !strings.EqualFold(addr.Address, normalized) {
		return "", fmt.Errorf("invalid email address")
	}
	return normalized, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
