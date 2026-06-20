package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

// listPrompts handles GET /api/v1/prompts
func (s *Server) listPrompts(c *gin.Context) {
	enabled := shared.ParseEnabledFilter(c)

	page, limit := s.parsePagination(c)

	prompts, err := s.promptService.ListPrompts(c.Request.Context(), enabled)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to list prompts: "+err.Error())
		return
	}

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

	s.successResponse(c, models.PaginatedResponse{
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

// createPrompt handles POST /api/v1/prompts (single or bulk)
func (s *Server) createPrompt(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	var header struct {
		Prompts json.RawMessage `json:"prompts"`
	}
	if err := json.Unmarshal(body, &header); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if header.Prompts != nil {
		var req models.BulkCreatePromptsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
			return
		}
		s.bulkCreatePrompts(c, req)
		return
	}

	var req models.CreatePromptRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := validateCreatePromptRequest(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, err.Error())
		return
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

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    promptToResponse(prompt),
		Message: "Prompt created successfully",
	})
}

func (s *Server) bulkCreatePrompts(c *gin.Context, req models.BulkCreatePromptsRequest) {
	templates := make([]string, 0, len(req.Prompts))
	for _, item := range req.Prompts {
		template := strings.TrimSpace(item.Template)
		if template != "" {
			templates = append(templates, template)
		}
	}

	if len(templates) == 0 {
		s.errorResponse(c, http.StatusBadRequest, "At least one prompt is required")
		return
	}

	saved, err := s.promptService.SaveGeneratedPrompts(c.Request.Context(), templates, req.Tags)
	if err != nil {
		s.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	responses := make([]models.PromptResponse, len(saved))
	for i, prompt := range saved {
		responses[i] = promptToResponse(prompt)
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data: models.BulkCreatePromptsResponse{
			Prompts:    responses,
			SavedCount: len(responses),
		},
		Message: "Prompts created successfully",
	})
}

func validateCreatePromptRequest(req *models.CreatePromptRequest) error {
	if strings.TrimSpace(req.Template) == "" {
		return fmt.Errorf("template is required")
	}

	if len(req.Template) > 10000 {
		return fmt.Errorf("template too long (max 10000 characters)")
	}

	if len(req.Tags) > 20 {
		return fmt.Errorf("too many tags (max 20)")
	}

	for i, tag := range req.Tags {
		if len(tag) > 50 {
			return fmt.Errorf("tag %s too long (max 50 characters)", strconv.Itoa(i+1))
		}
	}

	return nil
}

func promptToResponse(prompt *models.Prompt) models.PromptResponse {
	return models.PromptResponse{
		ID:        prompt.ID,
		Template:  prompt.Template,
		Tags:      prompt.Tags,
		Enabled:   prompt.Enabled,
		CreatedAt: prompt.CreatedAt,
		UpdatedAt: prompt.UpdatedAt,
	}
}

// updatePrompt handles PUT /api/v1/prompts/:id
func (s *Server) updatePrompt(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdatePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	prompt, err := s.promptService.GetPrompt(c.Request.Context(), id)
	if err != nil {
		s.errorResponse(c, http.StatusNotFound, "Prompt not found: "+err.Error())
		return
	}

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
