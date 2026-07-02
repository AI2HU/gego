package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/services"
)

func (s *Server) listRequiredUpgrades(c *gin.Context) {
	_ = s.upgradeService.ReloadConfig(s.configPath)

	items, err := s.upgradeService.ListUpgrades(c.Request.Context())
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	s.successResponse(c, items)
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

	if code == services.UpgradeSQLiteToPostgres {
		if err := s.reconnectSQLFromConfig(c.Request.Context()); err != nil {
			s.errorResponse(c, http.StatusInternalServerError,
				fmt.Sprintf("upgrade completed but API failed to switch to PostgreSQL: %v", err))
			return
		}
		result.RestartRequired = false
		if result.Message != "" {
			result.Message += " "
		}
		result.Message += "API is now using PostgreSQL. Restart the worker process if it is running."
	}

	s.successResponse(c, models.RunUpgradeResponse{
		UpgradeCode:     result.UpgradeCode,
		Status:          result.Status,
		Message:         result.Message,
		RestartRequired: result.RestartRequired,
	})
}
