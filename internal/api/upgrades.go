package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/services"
)

func (s *Server) listRequiredUpgrades(c *gin.Context) {
	codes, err := s.upgradeService.ListRequired(c.Request.Context())
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	if codes == nil {
		codes = []string{}
	}
	s.successResponse(c, models.UpgradesStatusResponse{
		RequiredUpgradeCodes: codes,
	})
}

func (s *Server) runUpgrade(c *gin.Context) {
	var req models.RunUpgradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	code := strings.TrimSpace(req.UpgradeCode)
	if code == "" {
		s.errorResponse(c, http.StatusBadRequest, "upgrade_code is required")
		return
	}

	result, err := s.upgradeService.Run(c.Request.Context(), code, services.UpgradeRunOptions{
		ConfigPath: s.configPath,
	})
	if err != nil {
		status := http.StatusInternalServerError
		msg := err.Error()
		switch {
		case strings.Contains(msg, "unknown upgrade code"):
			status = http.StatusBadRequest
		case strings.Contains(msg, "upgrade already in progress"):
			status = http.StatusConflict
		case strings.Contains(msg, "postgres URI is required"):
			status = http.StatusPreconditionFailed
		}
		s.errorResponse(c, status, msg)
		return
	}

	s.successResponse(c, models.RunUpgradeResponse{
		UpgradeCode:     result.UpgradeCode,
		Status:          result.Status,
		Message:         result.Message,
		RestartRequired: result.RestartRequired,
	})
}
