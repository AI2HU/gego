package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

// listLLMs handles GET /api/v1/llms
func (s *Server) listLLMs(c *gin.Context) {
	enabled := shared.ParseEnabledFilter(c)

	llms, err := s.llmService.ListLLMs(c.Request.Context(), enabled)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to list LLMs: "+err.Error())
		return
	}

	responses := make([]models.LLMResponse, len(llms))
	for i, llm := range llms {
		responses[i] = models.LLMResponse{
			ID:        llm.ID,
			Name:      llm.Name,
			Provider:  llm.Provider,
			Model:     llm.Model,
			APIKey:    s.maskAPIKey(llm.APIKey),
			BaseURL:   llm.BaseURL,
			Config:    llm.Config,
			Enabled:   llm.Enabled,
			CreatedAt: llm.CreatedAt,
			UpdatedAt: llm.UpdatedAt,
		}
	}

	s.successResponse(c, responses)
}

// getLLM handles GET /api/v1/llms/:id
func (s *Server) getLLM(c *gin.Context) {
	id := c.Param("id")

	llm, err := s.llmService.GetLLM(c.Request.Context(), id)
	if err != nil {
		s.errorResponse(c, http.StatusNotFound, "LLM not found: "+err.Error())
		return
	}

	response := models.LLMResponse{
		ID:        llm.ID,
		Name:      llm.Name,
		Provider:  llm.Provider,
		Model:     llm.Model,
		APIKey:    s.maskAPIKey(llm.APIKey),
		BaseURL:   llm.BaseURL,
		Config:    llm.Config,
		Enabled:   llm.Enabled,
		CreatedAt: llm.CreatedAt,
		UpdatedAt: llm.UpdatedAt,
	}

	s.successResponse(c, response)
}

// createLLM handles POST /api/v1/llms
func (s *Server) createLLM(c *gin.Context) {
	var req models.CreateLLMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if !s.isValidProvider(req.Provider) {
		s.errorResponse(c, http.StatusBadRequest, "Invalid provider. Must be one of: openai, anthropic, ollama, google, perplexity")
		return
	}

	llm := &models.LLMConfig{
		ID:       uuid.New().String(),
		Name:     req.Name,
		Provider: req.Provider,
		Model:    req.Model,
		APIKey:   req.APIKey,
		BaseURL:  req.BaseURL,
		Config:   req.Config,
		Enabled:  req.Enabled,
	}

	if err := s.llmService.CreateLLM(c.Request.Context(), llm); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to create LLM: "+err.Error())
		return
	}

	response := models.LLMResponse{
		ID:        llm.ID,
		Name:      llm.Name,
		Provider:  llm.Provider,
		Model:     llm.Model,
		APIKey:    s.maskAPIKey(llm.APIKey),
		BaseURL:   llm.BaseURL,
		Config:    llm.Config,
		Enabled:   llm.Enabled,
		CreatedAt: llm.CreatedAt,
		UpdatedAt: llm.UpdatedAt,
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    response,
		Message: "LLM created successfully",
	})
}

// updateLLM handles PUT /api/v1/llms/:id
func (s *Server) updateLLM(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateLLMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	llm, err := s.llmService.GetLLM(c.Request.Context(), id)
	if err != nil {
		s.errorResponse(c, http.StatusNotFound, "LLM not found: "+err.Error())
		return
	}

	if req.Name != "" {
		llm.Name = req.Name
	}
	if req.Provider != "" {
		if !s.isValidProvider(req.Provider) {
			s.errorResponse(c, http.StatusBadRequest, "Invalid provider. Must be one of: openai, anthropic, ollama, google, perplexity")
			return
		}
		llm.Provider = req.Provider
	}
	if req.Model != "" {
		llm.Model = req.Model
	}
	if req.APIKey != "" {
		llm.APIKey = req.APIKey
	}
	if req.BaseURL != "" {
		llm.BaseURL = req.BaseURL
	}
	if req.Config != nil {
		llm.Config = req.Config
	}
	if req.Enabled != nil {
		llm.Enabled = *req.Enabled
	}

	if err := s.llmService.UpdateLLM(c.Request.Context(), llm); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to update LLM: "+err.Error())
		return
	}

	response := models.LLMResponse{
		ID:        llm.ID,
		Name:      llm.Name,
		Provider:  llm.Provider,
		Model:     llm.Model,
		APIKey:    s.maskAPIKey(llm.APIKey),
		BaseURL:   llm.BaseURL,
		Config:    llm.Config,
		Enabled:   llm.Enabled,
		CreatedAt: llm.CreatedAt,
		UpdatedAt: llm.UpdatedAt,
	}

	s.successResponse(c, response)
}

// deleteLLM handles DELETE /api/v1/llms/:id
func (s *Server) deleteLLM(c *gin.Context) {
	id := c.Param("id")

	if err := s.llmService.DeleteLLM(c.Request.Context(), id); err != nil {
		s.errorResponse(c, http.StatusNotFound, "LLM not found: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "LLM deleted successfully",
	})
}

// Helper functions for LLM endpoints
func (s *Server) isValidProvider(provider string) bool {
	validProviders := []string{"openai", "anthropic", "ollama", "google", "perplexity"}
	for _, valid := range validProviders {
		if provider == valid {
			return true
		}
	}
	return false
}

func (s *Server) maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return ""
	}
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}
