package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/config"
	"github.com/AI2HU/gego/internal/llm"
	"github.com/AI2HU/gego/internal/llm/anthropic"
	"github.com/AI2HU/gego/internal/llm/google"
	"github.com/AI2HU/gego/internal/llm/ollama"
	"github.com/AI2HU/gego/internal/llm/openai"
	"github.com/AI2HU/gego/internal/llm/perplexity"
	"github.com/AI2HU/gego/internal/models"
)

func (s *Server) getSchedulerStatus(c *gin.Context) {
	running, count, err := s.schedulerService.GetStatus(c.Request.Context())
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get scheduler status: "+err.Error())
		return
	}

	s.successResponse(c, models.SchedulerStatusResponse{
		Running:          running,
		EnabledSchedules: count,
	})
}

func (s *Server) startScheduler(c *gin.Context) {
	ctx := c.Request.Context()

	if err := s.initializeLLMProviders(ctx); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to initialize LLM providers: "+err.Error())
		return
	}

	if err := s.schedulerService.Start(ctx); err != nil {
		s.errorResponse(c, http.StatusConflict, err.Error())
		return
	}

	running, count, err := s.schedulerService.GetStatus(ctx)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get scheduler status: "+err.Error())
		return
	}

	s.successResponse(c, models.SchedulerStatusResponse{
		Running:          running,
		EnabledSchedules: count,
	})
}

func (s *Server) stopScheduler(c *gin.Context) {
	ctx := c.Request.Context()
	s.schedulerService.Stop()

	_, count, err := s.schedulerService.GetStatus(ctx)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get scheduler status: "+err.Error())
		return
	}

	s.successResponse(c, models.SchedulerStatusResponse{
		Running:          false,
		EnabledSchedules: count,
	})
}

func (s *Server) reloadScheduler(c *gin.Context) {
	ctx := c.Request.Context()

	if err := s.initializeLLMProviders(ctx); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to initialize LLM providers: "+err.Error())
		return
	}

	if err := s.schedulerService.Reload(ctx); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to reload scheduler: "+err.Error())
		return
	}

	running, count, err := s.schedulerService.GetStatus(ctx)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get scheduler status: "+err.Error())
		return
	}

	s.successResponse(c, models.SchedulerStatusResponse{
		Running:          running,
		EnabledSchedules: count,
	})
}

func (s *Server) runSchedule(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	if err := s.initializeLLMProviders(ctx); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to initialize LLM providers: "+err.Error())
		return
	}

	if err := s.schedulerService.ExecuteNow(ctx, id); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to execute schedule: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Schedule execution completed",
	})
}

func (s *Server) initializeLLMProviders(ctx context.Context) error {
	llms, err := s.db.ListLLMs(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list LLMs: %w", err)
	}

	geminiSystemInstruction := config.GetSystemInstruction(nil, config.ProviderGemini)
	chatGPTSystemInstruction := config.GetSystemInstruction(nil, config.ProviderChatGPT)

	for _, llmConfig := range llms {
		var provider llm.Provider

		switch llmConfig.Provider {
		case "openai":
			provider = openai.New(llmConfig.APIKey, llmConfig.BaseURL, chatGPTSystemInstruction)
		case "anthropic":
			provider = anthropic.New(llmConfig.APIKey, llmConfig.BaseURL)
		case "ollama":
			provider = ollama.New(llmConfig.BaseURL)
		case "google":
			provider = google.New(llmConfig.APIKey, llmConfig.BaseURL, geminiSystemInstruction)
		case "perplexity":
			provider = perplexity.New(llmConfig.APIKey, llmConfig.BaseURL)
		default:
			continue
		}

		s.llmRegistry.Register(provider)
	}

	return nil
}
