package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/AI2HU/gego/internal/models"
)

type accessTokenClaims struct {
	Role models.Role `json:"role"`
	jwt.RegisteredClaims
}

func SignAccessToken(cfg Config, user *models.User) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(cfg.AccessTokenTTL)

	claims := accessTokenClaims{
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			Subject:   user.ID,
			Audience:  jwt.ClaimStrings{cfg.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(cfg.Secret)
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign access token: %w", err)
	}

	return signed, int64(cfg.AccessTokenTTL.Seconds()), nil
}
