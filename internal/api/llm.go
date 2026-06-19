package api

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/services"
	"github.com/AI2HU/gego/internal/shared"
)

// listProviders handles GET /api/v1/providers
func (s *Server) listProviders(c *gin.Context) {
	providers := services.AllProviders()
	responses := make([]models.ProviderResponse, len(providers))
	for i, provider := range providers {
		responses[i] = models.ProviderResponse{
			ID:              provider.String(),
			DisplayName:     provider.DisplayName(),
			ConsoleURL:      provider.GetConsoleURL(),
			RequiresAPIKey:  provider != services.Ollama,
			RequiresBaseURL: provider == services.Ollama,
		}
	}
	s.successResponse(c, responses)
}

// listProviderAPIKeys handles GET /api/v1/providers/:provider/api-keys
func (s *Server) listProviderAPIKeys(c *gin.Context) {
	provider := c.Param("provider")
	if !s.isValidProvider(provider) {
		s.errorResponse(c, http.StatusBadRequest, "Invalid provider")
		return
	}

	keys, err := s.llmService.GetExistingAPIKeysForProvider(c.Request.Context(), provider)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to list API keys: "+err.Error())
		return
	}

	responses := make([]models.ProviderAPIKeyResponse, len(keys))
	for i, key := range keys {
		responses[i] = models.ProviderAPIKeyResponse{
			Index:  i,
			Masked: s.maskAPIKey(key),
		}
	}
	s.successResponse(c, responses)
}

// listProviderModels handles POST /api/v1/providers/:provider/models
func (s *Server) listProviderModels(c *gin.Context) {
	providerName := c.Param("provider")
	if !s.isValidProvider(providerName) {
		s.errorResponse(c, http.StatusBadRequest, "Invalid provider")
		return
	}

	var req models.ListProviderModelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	apiKey := req.APIKey
	baseURL := req.BaseURL

	if req.ExistingKeyIndex != nil {
		keys, err := s.llmService.GetExistingAPIKeysForProvider(c.Request.Context(), providerName)
		if err != nil {
			s.errorResponse(c, http.StatusInternalServerError, "Failed to resolve API key: "+err.Error())
			return
		}
		idx := *req.ExistingKeyIndex
		if idx < 0 || idx >= len(keys) {
			s.errorResponse(c, http.StatusBadRequest, "Invalid existing_key_index")
			return
		}
		apiKey = keys[idx]
	}

	provider, ok := s.llmRegistry.Get(providerName)
	if !ok {
		s.errorResponse(c, http.StatusBadRequest, "Provider not available: "+providerName)
		return
	}

	providerEnum := services.FromString(providerName)
	if providerEnum != services.Ollama && apiKey == "" {
		s.errorResponse(c, http.StatusBadRequest, "API key is required for "+providerEnum.DisplayName())
		return
	}

	if providerEnum == services.Ollama && baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	modelsList, err := provider.ListModels(c.Request.Context(), apiKey, baseURL)
	if err != nil {
		s.errorResponse(c, http.StatusBadGateway, "Failed to list models: "+err.Error())
		return
	}

	responses := make([]models.ModelInfoResponse, len(modelsList))
	for i, model := range modelsList {
		responses[i] = models.ModelInfoResponse{
			ID:          model.ID,
			Name:        model.Name,
			Description: model.Description,
			UsedInChat:  model.UsedInChat,
		}
	}
	s.successResponse(c, responses)
}

// listLLMs handles GET /api/v1/models
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

// getLLM handles GET /api/v1/models/:id
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

// createLLM handles POST /api/v1/models
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

	apiKey := req.APIKey
	if req.ExistingKeyIndex != nil {
		keys, err := s.llmService.GetExistingAPIKeysForProvider(c.Request.Context(), req.Provider)
		if err != nil {
			s.errorResponse(c, http.StatusInternalServerError, "Failed to resolve API key: "+err.Error())
			return
		}
		idx := *req.ExistingKeyIndex
		if idx < 0 || idx >= len(keys) {
			s.errorResponse(c, http.StatusBadRequest, "Invalid existing_key_index")
			return
		}
		apiKey = keys[idx]
	}

	llm := &models.LLMConfig{
		ID:       uuid.New().String(),
		Name:     req.Name,
		Provider: req.Provider,
		Model:    req.Model,
		APIKey:   apiKey,
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

// updateLLM handles PUT /api/v1/models/:id
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

// deleteLLM handles DELETE /api/v1/models/:id
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
	return slices.Contains(validProviders, provider)
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
