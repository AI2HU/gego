package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/models"
)

// getStats handles GET /api/v1/stats
func (s *Server) getStats(c *gin.Context) {
	totalResponses, err := s.statsService.GetTotalResponses(c.Request.Context())
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get total responses: "+err.Error())
		return
	}

	totalPrompts, err := s.statsService.GetTotalPrompts(c.Request.Context())
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get total prompts: "+err.Error())
		return
	}

	totalLLMs, err := s.statsService.GetTotalLLMs(c.Request.Context())
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get total LLMs: "+err.Error())
		return
	}

	totalSchedules, err := s.statsService.GetTotalSchedules(c.Request.Context())
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get total schedules: "+err.Error())
		return
	}

	limitStr := c.DefaultQuery("keyword_limit", "10")
	keywordLimit, _ := strconv.Atoi(limitStr)
	if keywordLimit <= 0 || keywordLimit > 100 {
		keywordLimit = 10
	}

	topKeywords, err := s.statsService.GetTopKeywords(c.Request.Context(), keywordLimit, nil, nil)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get top keywords: "+err.Error())
		return
	}

	promptStats, err := s.statsService.GetAllPromptStats(c.Request.Context())
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get prompt stats: "+err.Error())
		return
	}

	llmStats, err := s.statsService.GetAllLLMStats(c.Request.Context())
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get LLM stats: "+err.Error())
		return
	}

	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -30)
	responseTrends, err := s.statsService.GetResponseTrends(c.Request.Context(), startTime, endTime)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get response trends: "+err.Error())
		return
	}

	response := models.StatsResponse{
		TotalResponses: totalResponses,
		TotalPrompts:   totalPrompts,
		TotalLLMs:      totalLLMs,
		TotalSchedules: totalSchedules,
		TopKeywords:    topKeywords,
		PromptStats:    promptStats,
		LLMStats:       llmStats,
		ResponseTrends: responseTrends,
		LastUpdated:    time.Now(),
	}

	s.successResponse(c, response)
}

// healthCheck handles GET /api/v1/health
func (s *Server) healthCheck(c *gin.Context) {
	if err := s.db.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Error:   "Database connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now(),
			"version":   "1.0.0",
		},
	})
}
