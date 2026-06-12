package auth

import (
	"fmt"
	"os"
	"time"

	appconfig "github.com/AI2HU/gego/internal/config"
)

const minSecretLength = 32

type Config struct {
	Issuer         string
	Audience       string
	AccessTokenTTL time.Duration
	Secret         []byte
}

func NewConfig(cfg *appconfig.Config) (Config, error) {
	authCfg := cfg.Auth
	if authCfg.Issuer == "" {
		authCfg.Issuer = "gego-api"
	}
	if authCfg.Audience == "" {
		authCfg.Audience = "gego-api"
	}
	accessTokenTTL := 15 * time.Minute
	if authCfg.AccessTokenTTL != "" {
		parsed, err := time.ParseDuration(authCfg.AccessTokenTTL)
		if err != nil {
			return Config{}, fmt.Errorf("invalid auth.access_token_ttl: %w", err)
		}
		if parsed <= 0 {
			return Config{}, fmt.Errorf("auth.access_token_ttl must be positive")
		}
		accessTokenTTL = parsed
	}

	secret := os.Getenv("GEGO_JWT_SECRET")
	if len(secret) < minSecretLength {
		return Config{}, fmt.Errorf("GEGO_JWT_SECRET must be set and at least %d characters", minSecretLength)
	}

	return Config{
		Issuer:         authCfg.Issuer,
		Audience:       authCfg.Audience,
		AccessTokenTTL: accessTokenTTL,
		Secret:         []byte(secret),
	}, nil
}
