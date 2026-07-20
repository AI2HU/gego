package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/services"
)

func (s *Server) getSMTPSettings(c *gin.Context) {
	if !s.requireAdminRole(c) {
		return
	}

	settings, err := s.mailService.GetSMTPSettings(c.Request.Context())
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to load SMTP settings: "+err.Error())
		return
	}
	s.successResponse(c, settings)
}

func (s *Server) updateSMTPSettings(c *gin.Context) {
	if !s.requireAdminRole(c) {
		return
	}

	var req models.UpdateSMTPSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	settings, err := s.mailService.UpdateSMTPSettings(c.Request.Context(), req)
	if err != nil {
		s.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	s.successResponse(c, settings)
}

func (s *Server) testSMTPSettings(c *gin.Context) {
	if !s.requireAdminRole(c) {
		return
	}

	var req models.TestSMTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := s.mailService.TestSMTP(c.Request.Context(), req); err != nil {
		if errors.Is(err, services.ErrSMTPNotConfigured) {
			s.errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		s.errorResponse(c, http.StatusBadRequest, "SMTP test failed: "+err.Error())
		return
	}

	s.successResponse(c, gin.H{"ok": true})
}
