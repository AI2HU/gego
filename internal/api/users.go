package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/auth"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/services"
)

func (s *Server) listUsers(c *gin.Context) {
	users, err := s.authService.ListUserProfiles(c.Request.Context())
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to list users: "+err.Error())
		return
	}
	s.successResponse(c, users)
}

func (s *Server) createUser(c *gin.Context) {
	if !s.requireAdminRole(c) {
		return
	}

	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	user, err := s.authService.CreateUser(c.Request.Context(), req.Username, req.Password, req.Role)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			s.errorResponse(c, http.StatusConflict, "username already exists")
			return
		}
		s.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	s.successResponse(c, models.ToAuthUserResponse(user))
}

func (s *Server) updateUser(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if req.Role == nil && req.Password == nil {
		s.errorResponse(c, http.StatusBadRequest, "role or password is required")
		return
	}

	user, err := s.authService.UpdateUser(c.Request.Context(), id, req.Role, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			s.errorResponse(c, http.StatusNotFound, "user not found")
			return
		}
		if errors.Is(err, services.ErrLastAdmin) {
			s.errorResponse(c, http.StatusConflict, err.Error())
			return
		}
		s.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	s.successResponse(c, models.ToAuthUserResponse(user))
}

func (s *Server) deleteUser(c *gin.Context) {
	if !s.requireAdminRole(c) {
		return
	}

	id := c.Param("id")
	actorID, err := auth.GetUserID(c)
	if err != nil {
		s.errorResponse(c, http.StatusUnauthorized, "authentication required")
		return
	}

	if err := s.authService.DeleteUser(c.Request.Context(), id, actorID); err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			s.errorResponse(c, http.StatusNotFound, "user not found")
			return
		}
		if errors.Is(err, services.ErrCannotDeleteSelf) || errors.Is(err, services.ErrLastAdmin) {
			s.errorResponse(c, http.StatusConflict, err.Error())
			return
		}
		s.errorResponse(c, http.StatusInternalServerError, "Failed to delete user: "+err.Error())
		return
	}

	s.successResponse(c, gin.H{"id": id})
}

func (s *Server) requireAdminRole(c *gin.Context) bool {
	role, err := auth.GetRole(c)
	if err != nil {
		s.errorResponse(c, http.StatusUnauthorized, "authentication required")
		return false
	}
	if role != models.RoleAdmin {
		s.errorResponse(c, http.StatusForbidden, "admin role required")
		return false
	}
	return true
}
