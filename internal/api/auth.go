package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/auth"
	"github.com/AI2HU/gego/internal/services"
)

const refreshTokenCookiePath = "/api/v1/auth"

func (s *Server) login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	result, err := s.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			s.errorResponse(c, http.StatusUnauthorized, "invalid credentials")
			return
		}
		s.errorResponse(c, http.StatusInternalServerError, "Failed to login: "+err.Error())
		return
	}

	s.setRefreshTokenCookie(c, result.RefreshToken)
	s.successResponse(c, result.LoginResponse)
}

func (s *Server) refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(auth.RefreshTokenCookieName)
	if err != nil || refreshToken == "" {
		s.errorResponse(c, http.StatusUnauthorized, "refresh token required")
		return
	}

	result, err := s.authService.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		s.clearRefreshTokenCookie(c)
		if errors.Is(err, services.ErrInvalidSession) {
			s.errorResponse(c, http.StatusUnauthorized, "invalid session")
			return
		}
		s.errorResponse(c, http.StatusInternalServerError, "Failed to refresh session: "+err.Error())
		return
	}

	s.setRefreshTokenCookie(c, result.RefreshToken)
	s.successResponse(c, result.LoginResponse)
}

func (s *Server) logout(c *gin.Context) {
	refreshToken, _ := c.Cookie(auth.RefreshTokenCookieName)
	_ = s.authService.Logout(c.Request.Context(), refreshToken)
	s.clearRefreshTokenCookie(c)
	s.successResponse(c, gin.H{"message": "logged out"})
}

func (s *Server) me(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		s.errorResponse(c, http.StatusUnauthorized, "authentication required")
		return
	}

	profile, err := s.authService.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		s.errorResponse(c, http.StatusNotFound, "User not found: "+err.Error())
		return
	}

	s.successResponse(c, profile)
}

func (s *Server) requirePerm(perm auth.Permission) gin.HandlerFunc {
	return auth.RequirePermission(perm)
}

func (s *Server) setRefreshTokenCookie(c *gin.Context, token string) {
	cfg := s.authService.Config()
	maxAge := int(cfg.RefreshTokenTTL.Seconds())
	if token == "" {
		maxAge = -1
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		auth.RefreshTokenCookieName,
		token,
		maxAge,
		refreshTokenCookiePath,
		"",
		cfg.CookieSecure,
		true,
	)
}

func (s *Server) clearRefreshTokenCookie(c *gin.Context) {
	s.setRefreshTokenCookie(c, "")
}
