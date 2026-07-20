package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
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

	result, err := s.authService.CreateInvitedUser(c.Request.Context(), req.Email, req.Role)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			s.errorResponse(c, http.StatusConflict, "email already exists")
			return
		}
		s.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	inviteURL := buildInviteURL(c, result.Token)
	emailSent := s.sendInviteEmail(c, result.User.Username, inviteURL)

	s.successResponse(c, models.CreateUserResponse{
		User:      models.ToAuthUserResponse(result.User),
		InviteURL: inviteURL,
		EmailSent: emailSent,
	})
}

func (s *Server) inviteUser(c *gin.Context) {
	if !s.requireAdminRole(c) {
		return
	}

	id := c.Param("id")
	result, err := s.authService.CreatePasswordInviteForUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			s.errorResponse(c, http.StatusNotFound, "user not found")
			return
		}
		s.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	inviteURL := buildInviteURL(c, result.Token)
	emailSent := s.sendInviteEmail(c, result.User.Username, inviteURL)

	s.successResponse(c, models.InviteUserResponse{
		User:      models.ToAuthUserResponse(result.User),
		InviteURL: inviteURL,
		EmailSent: emailSent,
	})
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

func (s *Server) sendInviteEmail(c *gin.Context, toEmail, inviteURL string) bool {
	if s.mailService == nil {
		return false
	}

	err := s.mailService.Send(c.Request.Context(), models.SendEmailRequest{
		To:      []string{toEmail},
		Subject: "You're invited to Gego — set your password",
		Body: fmt.Sprintf(`Hello,

You've been invited to join Gego.

To get started, set your password using the link below. This link expires in 1 week:

%s

If you did not expect this invitation, you can ignore this email.

—
Gego — See what AI says about your brand.
Gego helps teams track how AI assistants mention and describe their brand across models and prompts.
`, inviteURL),
	})
	if err != nil {
		return false
	}
	return true
}

func buildInviteURL(c *gin.Context, token string) string {
	base := strings.TrimRight(os.Getenv("GEGO_PUBLIC_URL"), "/")
	if base == "" {
		base = strings.TrimRight(c.GetHeader("Origin"), "/")
	}
	if base == "" {
		if referer := c.GetHeader("Referer"); referer != "" {
			if u, err := url.Parse(referer); err == nil && u.Scheme != "" && u.Host != "" {
				base = u.Scheme + "://" + u.Host
			}
		}
	}
	if base == "" {
		return "/set-password?token=" + url.QueryEscape(token)
	}
	return base + "/set-password?token=" + url.QueryEscape(token)
}
