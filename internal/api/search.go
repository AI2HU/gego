package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

// search handles POST /api/v1/search
func (s *Server) search(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if len(req.Keyword) < 2 {
		s.errorResponse(c, http.StatusBadRequest, "Keyword must be at least 2 characters long")
		return
	}
	if len(req.Keyword) > 100 {
		s.errorResponse(c, http.StatusBadRequest, "Keyword must be no more than 100 characters long")
		return
	}

	if req.Limit <= 0 || req.Limit > 1000 {
		req.Limit = 100
	}

	keywordStats, err := s.searchService.SearchKeyword(c.Request.Context(), req.Keyword, req.StartTime, req.EndTime)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to search keyword: "+err.Error())
		return
	}

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
