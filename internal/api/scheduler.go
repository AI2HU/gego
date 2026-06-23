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
	ctx, cancel := etcdRequestContext(c.Request.Context())
	defer cancel()

	enabled := true
	schedules, err := s.db.ListSchedules(ctx, &enabled)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to list schedules: "+err.Error())
		return
	}

	running, _, isLeader, pendingJobs, activeRuns, activeWorkers, err := s.schedulerService.GetStatus(ctx)
	if err != nil {
		s.errorResponse(c, http.StatusServiceUnavailable, "etcd unavailable: "+err.Error())
		return
	}

	s.successResponse(c, models.SchedulerStatusResponse{
		Running:          running,
		EnabledSchedules: len(schedules),
		IsLeader:         isLeader,
		PendingJobs:      pendingJobs,
		ActiveRuns:       activeRuns,
		ActiveWorkers:    activeWorkers,
	})
}

func (s *Server) getSchedulerHealth(c *gin.Context) {
	ctx, cancel := etcdRequestContext(c.Request.Context())
	defer cancel()
	if err := s.jobStore.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Error:   "etcd unavailable: " + err.Error(),
		})
		return
	}

	workers, err := s.jobStore.ListWorkers(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Error:   "failed to list workers: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"etcd":            "ok",
			"active_workers":  len(workers),
		},
	})
}

func (s *Server) startScheduler(c *gin.Context) {
	ctx := c.Request.Context()

	if err := s.initializeLLMProviders(ctx); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to initialize LLM providers: "+err.Error())
		return
	}

	listFn := func(ctx context.Context) ([]*models.Schedule, error) {
		enabled := true
		return s.db.ListSchedules(ctx, &enabled)
	}

	if err := s.schedulerService.Start(ctx, listFn); err != nil {
		s.errorResponse(c, http.StatusConflict, err.Error())
		return
	}

	s.getSchedulerStatus(c)
}

func (s *Server) stopScheduler(c *gin.Context) {
	s.schedulerService.Stop()
	s.getSchedulerStatus(c)
}

func (s *Server) reloadScheduler(c *gin.Context) {
	ctx := c.Request.Context()

	if err := s.initializeLLMProviders(ctx); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to initialize LLM providers: "+err.Error())
		return
	}

	listFn := func(ctx context.Context) ([]*models.Schedule, error) {
		enabled := true
		return s.db.ListSchedules(ctx, &enabled)
	}

	if err := s.schedulerService.Reload(ctx, listFn); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to reload scheduler: "+err.Error())
		return
	}

	s.getSchedulerStatus(c)
}

func (s *Server) runSchedule(c *gin.Context) {
	id := c.Param("id")
	ctx, cancel := etcdRequestContext(c.Request.Context())
	defer cancel()

	workers, err := s.jobStore.ListWorkers(ctx)
	if err != nil {
		s.errorResponse(c, http.StatusServiceUnavailable, "etcd unavailable: "+err.Error())
		return
	}
	if len(workers) == 0 {
		s.errorResponse(c, http.StatusConflict, "No workers connected. Start a worker with: gego worker start")
		return
	}

	runID, err := s.enqueueService.EnqueueSchedule(ctx, id, models.ScheduleRunTriggerManual, "", 0)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to enqueue schedule: "+err.Error())
		return
	}

	c.JSON(http.StatusAccepted, models.APIResponse{
		Success: true,
		Data:    models.ScheduleRunEnqueueResponse{RunID: runID},
		Message: "Schedule enqueued",
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
