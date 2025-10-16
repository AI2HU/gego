package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/AI2HU/gego/internal/models"
)

// Schedule request/response structures
type CreateScheduleRequest struct {
	Name        string   `json:"name" binding:"required"`
	PromptIDs   []string `json:"prompt_ids" binding:"required"`
	LLMIDs      []string `json:"llm_ids" binding:"required"`
	CronExpr    string   `json:"cron_expr" binding:"required"`
	Temperature float64  `json:"temperature,omitempty"`
	Enabled     bool     `json:"enabled"`
}

type UpdateScheduleRequest struct {
	Name        string   `json:"name,omitempty"`
	PromptIDs   []string `json:"prompt_ids,omitempty"`
	LLMIDs      []string `json:"llm_ids,omitempty"`
	CronExpr    string   `json:"cron_expr,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	Enabled     *bool    `json:"enabled,omitempty"`
}

type ScheduleResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	PromptIDs   []string   `json:"prompt_ids"`
	LLMIDs      []string   `json:"llm_ids"`
	CronExpr    string     `json:"cron_expr"`
	Temperature float64    `json:"temperature"`
	Enabled     bool       `json:"enabled"`
	LastRun     *time.Time `json:"last_run,omitempty"`
	NextRun     *time.Time `json:"next_run,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Schedule endpoints

// listSchedules handles GET /api/v1/schedules
func (s *Server) listSchedules(c *gin.Context) {
	enabledStr := c.Query("enabled")
	var enabled *bool

	if enabledStr != "" {
		if enabledStr == "true" {
			enabled = &[]bool{true}[0]
		} else if enabledStr == "false" {
			enabled = &[]bool{false}[0]
		}
	}

	page, limit := s.parsePagination(c)

	schedules, err := s.scheduleService.ListSchedules(c.Request.Context(), enabled)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to list schedules: "+err.Error())
		return
	}

	// Apply pagination
	total := len(schedules)
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		schedules = []*models.Schedule{}
	} else {
		if end > total {
			end = total
		}
		schedules = schedules[start:end]
	}

	// Convert to response format
	responses := make([]ScheduleResponse, len(schedules))
	for i, schedule := range schedules {
		responses[i] = ScheduleResponse{
			ID:          schedule.ID,
			Name:        schedule.Name,
			PromptIDs:   schedule.PromptIDs,
			LLMIDs:      schedule.LLMIDs,
			CronExpr:    schedule.CronExpr,
			Temperature: schedule.Temperature,
			Enabled:     schedule.Enabled,
			LastRun:     schedule.LastRun,
			NextRun:     schedule.NextRun,
			CreatedAt:   schedule.CreatedAt,
			UpdatedAt:   schedule.UpdatedAt,
		}
	}

	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, PaginatedResponse{
		Data: responses,
		Pagination: Pagination{
			Page:       page,
			Limit:      limit,
			Total:      int64(total),
			TotalPages: totalPages,
		},
	})
}

// getSchedule handles GET /api/v1/schedules/:id
func (s *Server) getSchedule(c *gin.Context) {
	id := c.Param("id")

	schedule, err := s.scheduleService.GetSchedule(c.Request.Context(), id)
	if err != nil {
		s.errorResponse(c, http.StatusNotFound, "Schedule not found: "+err.Error())
		return
	}

	response := ScheduleResponse{
		ID:          schedule.ID,
		Name:        schedule.Name,
		PromptIDs:   schedule.PromptIDs,
		LLMIDs:      schedule.LLMIDs,
		CronExpr:    schedule.CronExpr,
		Temperature: schedule.Temperature,
		Enabled:     schedule.Enabled,
		LastRun:     schedule.LastRun,
		NextRun:     schedule.NextRun,
		CreatedAt:   schedule.CreatedAt,
		UpdatedAt:   schedule.UpdatedAt,
	}

	s.successResponse(c, response)
}

// createSchedule handles POST /api/v1/schedules
func (s *Server) createSchedule(c *gin.Context) {
	var req CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Validate temperature
	if req.Temperature < 0.0 || req.Temperature > 1.0 {
		s.errorResponse(c, http.StatusBadRequest, "Temperature must be between 0.0 and 1.0")
		return
	}

	// Validate arrays
	if len(req.PromptIDs) == 0 {
		s.errorResponse(c, http.StatusBadRequest, "At least one prompt ID is required")
		return
	}
	if len(req.LLMIDs) == 0 {
		s.errorResponse(c, http.StatusBadRequest, "At least one LLM ID is required")
		return
	}
	if len(req.PromptIDs) > 50 {
		s.errorResponse(c, http.StatusBadRequest, "Maximum 50 prompts allowed per schedule")
		return
	}
	if len(req.LLMIDs) > 50 {
		s.errorResponse(c, http.StatusBadRequest, "Maximum 50 LLMs allowed per schedule")
		return
	}

	// Validate cron expression (basic validation)
	if len(req.CronExpr) == 0 {
		s.errorResponse(c, http.StatusBadRequest, "Cron expression is required")
		return
	}

	// Validate that referenced prompts and LLMs exist
	if err := s.validateScheduleReferences(c.Request.Context(), req.PromptIDs, req.LLMIDs); err != nil {
		s.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	schedule := &models.Schedule{
		ID:          uuid.New().String(),
		Name:        req.Name,
		PromptIDs:   req.PromptIDs,
		LLMIDs:      req.LLMIDs,
		CronExpr:    req.CronExpr,
		Temperature: req.Temperature,
		Enabled:     req.Enabled,
	}

	if err := s.scheduleService.CreateSchedule(c.Request.Context(), schedule); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to create schedule: "+err.Error())
		return
	}

	response := ScheduleResponse{
		ID:          schedule.ID,
		Name:        schedule.Name,
		PromptIDs:   schedule.PromptIDs,
		LLMIDs:      schedule.LLMIDs,
		CronExpr:    schedule.CronExpr,
		Temperature: schedule.Temperature,
		Enabled:     schedule.Enabled,
		LastRun:     schedule.LastRun,
		NextRun:     schedule.NextRun,
		CreatedAt:   schedule.CreatedAt,
		UpdatedAt:   schedule.UpdatedAt,
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    response,
		Message: "Schedule created successfully",
	})
}

// updateSchedule handles PUT /api/v1/schedules/:id
func (s *Server) updateSchedule(c *gin.Context) {
	id := c.Param("id")

	var req UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Get existing schedule
	schedule, err := s.scheduleService.GetSchedule(c.Request.Context(), id)
	if err != nil {
		s.errorResponse(c, http.StatusNotFound, "Schedule not found: "+err.Error())
		return
	}

	// Update fields if provided
	if req.Name != "" {
		schedule.Name = req.Name
	}
	if req.PromptIDs != nil {
		if len(req.PromptIDs) == 0 {
			s.errorResponse(c, http.StatusBadRequest, "At least one prompt ID is required")
			return
		}
		if len(req.PromptIDs) > 50 {
			s.errorResponse(c, http.StatusBadRequest, "Maximum 50 prompts allowed per schedule")
			return
		}
		schedule.PromptIDs = req.PromptIDs
	}
	if req.LLMIDs != nil {
		if len(req.LLMIDs) == 0 {
			s.errorResponse(c, http.StatusBadRequest, "At least one LLM ID is required")
			return
		}
		if len(req.LLMIDs) > 50 {
			s.errorResponse(c, http.StatusBadRequest, "Maximum 50 LLMs allowed per schedule")
			return
		}
		schedule.LLMIDs = req.LLMIDs
	}
	if req.CronExpr != "" {
		schedule.CronExpr = req.CronExpr
	}
	if req.Temperature != nil {
		if *req.Temperature < 0.0 || *req.Temperature > 1.0 {
			s.errorResponse(c, http.StatusBadRequest, "Temperature must be between 0.0 and 1.0")
			return
		}
		schedule.Temperature = *req.Temperature
	}
	if req.Enabled != nil {
		schedule.Enabled = *req.Enabled
	}

	// Validate references if they were updated
	if req.PromptIDs != nil || req.LLMIDs != nil {
		if err := s.validateScheduleReferences(c.Request.Context(), schedule.PromptIDs, schedule.LLMIDs); err != nil {
			s.errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
	}

	if err := s.scheduleService.UpdateSchedule(c.Request.Context(), schedule); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to update schedule: "+err.Error())
		return
	}

	response := ScheduleResponse{
		ID:          schedule.ID,
		Name:        schedule.Name,
		PromptIDs:   schedule.PromptIDs,
		LLMIDs:      schedule.LLMIDs,
		CronExpr:    schedule.CronExpr,
		Temperature: schedule.Temperature,
		Enabled:     schedule.Enabled,
		LastRun:     schedule.LastRun,
		NextRun:     schedule.NextRun,
		CreatedAt:   schedule.CreatedAt,
		UpdatedAt:   schedule.UpdatedAt,
	}

	s.successResponse(c, response)
}

// deleteSchedule handles DELETE /api/v1/schedules/:id
func (s *Server) deleteSchedule(c *gin.Context) {
	id := c.Param("id")

	if err := s.scheduleService.DeleteSchedule(c.Request.Context(), id); err != nil {
		s.errorResponse(c, http.StatusNotFound, "Schedule not found: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Schedule deleted successfully",
	})
}

// validateScheduleReferences validates that all referenced prompts and LLMs exist
func (s *Server) validateScheduleReferences(ctx context.Context, promptIDs, llmIDs []string) error {
	// Validate prompts exist
	for _, promptID := range promptIDs {
		if _, err := s.promptService.GetPrompt(ctx, promptID); err != nil {
			return fmt.Errorf("prompt not found: %s", promptID)
		}
	}

	// Validate LLMs exist
	for _, llmID := range llmIDs {
		if _, err := s.llmService.GetLLM(ctx, llmID); err != nil {
			return fmt.Errorf("LLM not found: %s", llmID)
		}
	}

	return nil
}
