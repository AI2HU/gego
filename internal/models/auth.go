package models

import (
	"context"
	"errors"
	"time"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

func (r Role) Valid() bool {
	return r == RoleAdmin || r == RoleMember
}

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string           `json:"access_token"`
	TokenType   string           `json:"token_type"`
	ExpiresIn   int64            `json:"expires_in"`
	User        AuthUserResponse `json:"user"`
}

type UserSession struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type PasswordInvite struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type AuthSessionResult struct {
	LoginResponse
	RefreshToken string `json:"-"`
}

type AuthUserResponse struct {
	ID              string    `json:"id"`
	Username        string    `json:"username"`
	Role            Role      `json:"role"`
	PasswordPending bool      `json:"password_pending"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateUserRequest struct {
	Email string `json:"email" binding:"required"`
	Role  Role   `json:"role" binding:"required"`
}

type CreateUserResponse struct {
	User      AuthUserResponse `json:"user"`
	InviteURL string           `json:"invite_url"`
	EmailSent bool             `json:"email_sent"`
}

type InviteUserResponse struct {
	User      AuthUserResponse `json:"user"`
	InviteURL string           `json:"invite_url"`
	EmailSent bool             `json:"email_sent"`
}

type SetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateUserRequest struct {
	Role     *Role   `json:"role"`
	Password *string `json:"password"`
}

func ToAuthUserResponse(user *User) AuthUserResponse {
	return AuthUserResponse{
		ID:              user.ID,
		Username:        user.Username,
		Role:            user.Role,
		PasswordPending: user.PasswordHash == "",
		CreatedAt:       user.CreatedAt,
	}
}

type JWTCustomClaims struct {
	Role Role `json:"role"`
}

func (c *JWTCustomClaims) Validate(ctx context.Context) error {
	if !c.Role.Valid() {
		return errors.New("invalid role claim")
	}
	return nil
}
