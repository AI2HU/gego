package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

// Stats request/response structures are now defined in models package

// Stats endpoint

// getStats handles GET /api/v1/stats
func (s *Server) getStats(c *gin.Context) {
	// Get basic counts
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

	// Get top keywords
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

	// Get prompt stats
	promptStats, err := s.statsService.GetPromptStats(c.Request.Context())
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get prompt stats: "+err.Error())
		return
	}

	// Get LLM stats
	llmStats, err := s.statsService.GetLLMStats(c.Request.Context())
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get LLM stats: "+err.Error())
		return
	}

	// Get response trends (last 30 days)
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

// Search endpoint

// search handles POST /api/v1/search
func (s *Server) search(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Validate keyword
	if len(req.Keyword) < 2 {
		s.errorResponse(c, http.StatusBadRequest, "Keyword must be at least 2 characters long")
		return
	}
	if len(req.Keyword) > 100 {
		s.errorResponse(c, http.StatusBadRequest, "Keyword must be no more than 100 characters long")
		return
	}

	// Set default limit
	if req.Limit <= 0 || req.Limit > 1000 {
		req.Limit = 100
	}

	// Perform keyword search
	keywordStats, err := s.searchService.SearchKeyword(c.Request.Context(), req.Keyword, req.StartTime, req.EndTime)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to search keyword: "+err.Error())
		return
	}

	// Get recent responses containing the keyword
	filter := shared.ResponseFilter{
		Keyword:   req.Keyword,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Limit:     req.Limit,
	}

	responses, err := s.searchService.ListResponses(c.Request.Context(), filter)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get responses: "+err.Error())
		return
	}

	response := models.SearchResponse{
		Keyword:       keywordStats.Keyword,
		TotalMentions: keywordStats.TotalMentions,
		UniquePrompts: keywordStats.UniquePrompts,
		UniqueLLMs:    keywordStats.UniqueLLMs,
		ByPrompt:      keywordStats.ByPrompt,
		ByLLM:         keywordStats.ByLLM,
		ByProvider:    keywordStats.ByProvider,
		FirstSeen:     keywordStats.FirstSeen,
		LastSeen:      keywordStats.LastSeen,
		Responses:     responses,
	}

	s.successResponse(c, response)
}

// Health check endpoint

// healthCheck handles GET /api/v1/health
func (s *Server) healthCheck(c *gin.Context) {
	// Test database connection
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
