package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/AI2HU/gego/internal/models"
)

// LLM request/response structures
type CreateLLMRequest struct {
	Name     string            `json:"name" binding:"required"`
	Provider string            `json:"provider" binding:"required"`
	Model    string            `json:"model" binding:"required"`
	APIKey   string            `json:"api_key,omitempty"`
	BaseURL  string            `json:"base_url,omitempty"`
	Config   map[string]string `json:"config,omitempty"`
	Enabled  bool              `json:"enabled"`
}

type UpdateLLMRequest struct {
	Name     string            `json:"name,omitempty"`
	Provider string            `json:"provider,omitempty"`
	Model    string            `json:"model,omitempty"`
	APIKey   string            `json:"api_key,omitempty"`
	BaseURL  string            `json:"base_url,omitempty"`
	Config   map[string]string `json:"config,omitempty"`
	Enabled  *bool             `json:"enabled,omitempty"`
}

type LLMResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Provider  string            `json:"provider"`
	Model     string            `json:"model"`
	APIKey    string            `json:"api_key,omitempty"`
	BaseURL   string            `json:"base_url,omitempty"`
	Config    map[string]string `json:"config,omitempty"`
	Enabled   bool              `json:"enabled"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// LLM endpoints

// listLLMs handles GET /api/v1/llms
func (s *Server) listLLMs(c *gin.Context) {
	enabledStr := c.Query("enabled")
	var enabled *bool

	if enabledStr != "" {
		if enabledStr == "true" {
			enabled = &[]bool{true}[0]
		} else if enabledStr == "false" {
			enabled = &[]bool{false}[0]
		}
	}

	llms, err := s.llmService.ListLLMs(c.Request.Context(), enabled)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to list LLMs: "+err.Error())
		return
	}

	// Convert to response format and mask API keys
	responses := make([]LLMResponse, len(llms))
	for i, llm := range llms {
		responses[i] = LLMResponse{
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

	response := LLMResponse{
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
	var req CreateLLMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Validate provider
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

	response := LLMResponse{
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

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    response,
		Message: "LLM created successfully",
	})
}

// updateLLM handles PUT /api/v1/llms/:id
func (s *Server) updateLLM(c *gin.Context) {
	id := c.Param("id")

	var req UpdateLLMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Get existing LLM
	llm, err := s.llmService.GetLLM(c.Request.Context(), id)
	if err != nil {
		s.errorResponse(c, http.StatusNotFound, "LLM not found: "+err.Error())
		return
	}

	// Update fields if provided
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

	response := LLMResponse{
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

	c.JSON(http.StatusOK, APIResponse{
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
