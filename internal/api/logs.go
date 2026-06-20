package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

func (s *Server) listErrorLogs(c *gin.Context) {
	page, limit := s.parsePagination(c)
	llmID := c.Query("llm_id")

	filter := shared.ResponseFilter{
		LLMID:      llmID,
		ErrorsOnly: true,
		Limit:      limit,
		Offset:     (page - 1) * limit,
	}

	total, err := s.searchService.CountResponses(c.Request.Context(), filter)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to count error logs: "+err.Error())
		return
	}

	responses, err := s.searchService.ListResponses(c.Request.Context(), filter)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to list error logs: "+err.Error())
		return
	}

	logs := make([]models.ErrorLogResponse, len(responses))
	for i, response := range responses {
		logs[i] = models.ErrorLogResponse{
			ID:          response.ID,
			PromptID:    response.PromptID,
			PromptText:  response.PromptText,
			LLMID:       response.LLMID,
			LLMName:     response.LLMName,
			LLMProvider: response.LLMProvider,
			LLMModel:    response.LLMModel,
			Error:       response.Error,
			ScheduleID:  response.ScheduleID,
			Temperature: response.Temperature,
			CreatedAt:   response.CreatedAt,
		}
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	s.successResponse(c, models.PaginatedResponse{
		Data: logs,
		Pagination: models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
