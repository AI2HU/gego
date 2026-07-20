package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v3"
	"github.com/auth0/go-jwt-middleware/v3/validator"
	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/models"
)

type Middleware struct {
	cfg        Config
	jwtHandler gin.HandlerFunc
}

func NewMiddleware(cfg Config) (*Middleware, error) {
	keyFunc := func(ctx context.Context) (any, error) {
		return cfg.Secret, nil
	}

	jwtValidator, err := validator.New(
		validator.WithKeyFunc(keyFunc),
		validator.WithAlgorithm(validator.HS256),
		validator.WithIssuer(cfg.Issuer),
		validator.WithAudience(cfg.Audience),
		validator.WithCustomClaims(func() *models.JWTCustomClaims {
			return &models.JWTCustomClaims{}
		}),
		validator.WithAllowedClockSkew(30*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set up JWT validator: %w", err)
	}

	middleware, err := jwtmiddleware.New(
		jwtmiddleware.WithValidator(jwtValidator),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set up JWT middleware: %w", err)
	}

	jwtHandler := func(c *gin.Context) {
		encounteredError := true
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			encounteredError = false
			c.Request = r

			claims, err := jwtmiddleware.GetClaims[*validator.ValidatedClaims](r.Context())
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
					Success: false,
					Error:   "failed to read token claims",
				})
				return
			}

			customClaims, ok := claims.CustomClaims.(*models.JWTCustomClaims)
			if !ok || customClaims == nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
					Success: false,
					Error:   "invalid token claims",
				})
				return
			}

			SetAuthContext(c, claims.RegisteredClaims.Subject, customClaims.Role)
			c.Next()
		})

		middleware.CheckJWT(handler).ServeHTTP(c.Writer, c.Request)

		if encounteredError {
			c.Abort()
		}
	}

	return &Middleware{
		cfg:        cfg,
		jwtHandler: jwtHandler,
	}, nil
}

func (m *Middleware) Authenticate() gin.HandlerFunc {
	return m.jwtHandler
}

func RequirePermission(perm Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, err := GetRole(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "authentication required",
			})
			return
		}

		if !HasPermission(role, perm) {
			c.AbortWithStatusJSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error:   "insufficient permissions",
			})
			return
		}

		c.Next()
	}
}

func RequireRole(required models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, err := GetRole(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "authentication required",
			})
			return
		}

		if role != required {
			c.AbortWithStatusJSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error:   "admin role required",
			})
			return
		}

		c.Next()
	}
}
