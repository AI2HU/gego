package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

// Prompt request/response structures are now defined in models package

// Prompt endpoints

// listPrompts handles GET /api/v1/prompts
func (s *Server) listPrompts(c *gin.Context) {
	enabled := shared.ParseEnabledFilter(c)

	page, limit := s.parsePagination(c)

	prompts, err := s.promptService.ListPrompts(c.Request.Context(), enabled)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to list prompts: "+err.Error())
		return
	}

	// Apply pagination
	total := len(prompts)
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		prompts = []*models.Prompt{}
	} else {
		if end > total {
			end = total
		}
		prompts = prompts[start:end]
	}

	// Convert to response format
	responses := make([]models.PromptResponse, len(prompts))
	for i, prompt := range prompts {
		responses[i] = models.PromptResponse{
			ID:        prompt.ID,
			Template:  prompt.Template,
			Tags:      prompt.Tags,
			Enabled:   prompt.Enabled,
			CreatedAt: prompt.CreatedAt,
			UpdatedAt: prompt.UpdatedAt,
		}
	}

	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data: responses,
		Pagination: models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      int64(total),
			TotalPages: totalPages,
		},
	})
}

// getPrompt handles GET /api/v1/prompts/:id
func (s *Server) getPrompt(c *gin.Context) {
	id := c.Param("id")

	prompt, err := s.promptService.GetPrompt(c.Request.Context(), id)
	if err != nil {
		s.errorResponse(c, http.StatusNotFound, "Prompt not found: "+err.Error())
		return
	}

	response := models.PromptResponse{
		ID:        prompt.ID,
		Template:  prompt.Template,
		Tags:      prompt.Tags,
		Enabled:   prompt.Enabled,
		CreatedAt: prompt.CreatedAt,
		UpdatedAt: prompt.UpdatedAt,
	}

	s.successResponse(c, response)
}

// createPrompt handles POST /api/v1/prompts
func (s *Server) createPrompt(c *gin.Context) {
	var req models.CreatePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Validate template length
	if len(req.Template) > 10000 {
		s.errorResponse(c, http.StatusBadRequest, "Template too long (max 10000 characters)")
		return
	}

	// Validate tags
	if len(req.Tags) > 20 {
		s.errorResponse(c, http.StatusBadRequest, "Too many tags (max 20)")
		return
	}

	for i, tag := range req.Tags {
		if len(tag) > 50 {
			s.errorResponse(c, http.StatusBadRequest, "Tag "+strconv.Itoa(i+1)+" too long (max 50 characters)")
			return
		}
	}

	prompt := &models.Prompt{
		ID:       uuid.New().String(),
		Template: req.Template,
		Tags:     req.Tags,
		Enabled:  req.Enabled,
	}

	if err := s.promptService.CreatePrompt(c.Request.Context(), prompt); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to create prompt: "+err.Error())
		return
	}

	response := models.PromptResponse{
		ID:        prompt.ID,
		Template:  prompt.Template,
		Tags:      prompt.Tags,
		Enabled:   prompt.Enabled,
		CreatedAt: prompt.CreatedAt,
		UpdatedAt: prompt.UpdatedAt,
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    response,
		Message: "Prompt created successfully",
	})
}

// updatePrompt handles PUT /api/v1/prompts/:id
func (s *Server) updatePrompt(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdatePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Get existing prompt
	prompt, err := s.promptService.GetPrompt(c.Request.Context(), id)
	if err != nil {
		s.errorResponse(c, http.StatusNotFound, "Prompt not found: "+err.Error())
		return
	}

	// Update fields if provided
	if req.Template != "" {
		if len(req.Template) > 10000 {
			s.errorResponse(c, http.StatusBadRequest, "Template too long (max 10000 characters)")
			return
		}
		prompt.Template = req.Template
	}
	if req.Tags != nil {
		if len(req.Tags) > 20 {
			s.errorResponse(c, http.StatusBadRequest, "Too many tags (max 20)")
			return
		}
		for i, tag := range req.Tags {
			if len(tag) > 50 {
				s.errorResponse(c, http.StatusBadRequest, "Tag "+strconv.Itoa(i+1)+" too long (max 50 characters)")
				return
			}
		}
		prompt.Tags = req.Tags
	}
	if req.Enabled != nil {
		prompt.Enabled = *req.Enabled
	}

	if err := s.promptService.UpdatePrompt(c.Request.Context(), prompt); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to update prompt: "+err.Error())
		return
	}

	response := models.PromptResponse{
		ID:        prompt.ID,
		Template:  prompt.Template,
		Tags:      prompt.Tags,
		Enabled:   prompt.Enabled,
		CreatedAt: prompt.CreatedAt,
		UpdatedAt: prompt.UpdatedAt,
	}

	s.successResponse(c, response)
}

// deletePrompt handles DELETE /api/v1/prompts/:id
func (s *Server) deletePrompt(c *gin.Context) {
	id := c.Param("id")

	if err := s.promptService.DeletePrompt(c.Request.Context(), id); err != nil {
		s.errorResponse(c, http.StatusNotFound, "Prompt not found: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Prompt deleted successfully",
	})
}
