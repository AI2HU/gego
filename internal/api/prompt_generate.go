package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/services"
)

func (s *Server) generatePrompts(c *gin.Context) {
	var req models.GeneratePromptsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	req.LanguageCode = strings.ToUpper(strings.TrimSpace(req.LanguageCode))
	if err := services.ValidateLanguageCode(req.LanguageCode); err != nil {
		s.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	req.UserInput = strings.TrimSpace(req.UserInput)
	if req.UserInput == "" {
		s.errorResponse(c, http.StatusBadRequest, "Description is required")
		return
	}

	if req.PromptCount <= 0 {
		req.PromptCount = 20
	}

	llmConfig, err := s.llmService.GetLLM(c.Request.Context(), req.LLMID)
	if err != nil {
		s.errorResponse(c, http.StatusBadRequest, "LLM not found: "+err.Error())
		return
	}

	registry, err := services.NewLLMRegistryForConfig(llmConfig, nil)
	if err != nil {
		s.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	existingPrompts, err := s.promptService.ListPrompts(c.Request.Context(), nil)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to list existing prompts: "+err.Error())
		return
	}

	existingTemplates := make([]string, 0, len(existingPrompts))
	for _, prompt := range existingPrompts {
		existingTemplates = append(existingTemplates, prompt.Template)
	}

	genService := services.NewPromptGenerationService(registry)
	prompts, err := genService.GeneratePrompts(c.Request.Context(), llmConfig, &services.GenerationConfig{
		LanguageCode:    req.LanguageCode,
		UserInput:       req.UserInput,
		PromptCount:     req.PromptCount,
		ExistingPrompts: existingTemplates,
	})
	if err != nil {
		s.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	s.successResponse(c, models.GeneratePromptsResponse{
		Prompts: prompts,
	})
}
