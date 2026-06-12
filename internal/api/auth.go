package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/auth"
	"github.com/AI2HU/gego/internal/services"
)

func (s *Server) login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	response, err := s.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			s.errorResponse(c, http.StatusUnauthorized, "invalid credentials")
			return
		}
		s.errorResponse(c, http.StatusInternalServerError, "Failed to login: "+err.Error())
		return
	}

	s.successResponse(c, response)
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
