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
	"github.com/AI2HU/gego/internal/logger"
	"github.com/AI2HU/gego/internal/models"
)

const etcdUnavailablePrefix = "etcd unavailable: "

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
		s.errorResponse(c, http.StatusServiceUnavailable, etcdUnavailablePrefix+err.Error())
		return
	}

	settings, err := s.db.GetSchedulerSettings(ctx)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to load scheduler settings: "+err.Error())
		return
	}

	desiredRunning := false
	if settings != nil {
		desiredRunning = settings.DesiredRunning
	}

	s.successResponse(c, models.SchedulerStatusResponse{
		Running:          running,
		DesiredRunning:   desiredRunning,
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
			Error:   etcdUnavailablePrefix + err.Error(),
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
			"etcd":           "ok",
			"active_workers": len(workers),
		},
	})
}

func (s *Server) startScheduler(c *gin.Context) {
	ctx := context.Background()

	if err := s.initializeLLMProviders(ctx); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to initialize LLM providers: "+err.Error())
		return
	}

	if err := s.schedulerService.Start(ctx, s.listEnabledSchedules); err != nil {
		s.errorResponse(c, http.StatusConflict, err.Error())
		return
	}

	if err := s.persistSchedulerDesiredRunning(ctx, true); err != nil {
		logger.Error("Failed to persist scheduler desired running=true: %v", err)
	}

	s.getSchedulerStatus(c)
}

func (s *Server) stopScheduler(c *gin.Context) {
	ctx := context.Background()
	s.schedulerService.Stop()

	if err := s.persistSchedulerDesiredRunning(ctx, false); err != nil {
		logger.Error("Failed to persist scheduler desired running=false: %v", err)
	}

	s.getSchedulerStatus(c)
}

func (s *Server) reloadScheduler(c *gin.Context) {
	ctx := context.Background()

	if err := s.initializeLLMProviders(ctx); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to initialize LLM providers: "+err.Error())
		return
	}

	if err := s.schedulerService.Reload(ctx, s.listEnabledSchedules); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to reload scheduler: "+err.Error())
		return
	}

	if err := s.persistSchedulerDesiredRunning(ctx, true); err != nil {
		logger.Error("Failed to persist scheduler desired running=true: %v", err)
	}

	s.getSchedulerStatus(c)
}

func (s *Server) RestoreScheduler(ctx context.Context) {
	settings, err := s.db.GetSchedulerSettings(ctx)
	if err != nil {
		logger.Warning("Skipping scheduler restore: %v", err)
		return
	}
	if settings == nil || !settings.DesiredRunning {
		return
	}

	if err := s.initializeLLMProviders(ctx); err != nil {
		logger.Error("Failed to restore scheduler (LLM init): %v", err)
		return
	}

	if err := s.schedulerService.Start(ctx, s.listEnabledSchedules); err != nil {
		logger.Error("Failed to restore scheduler: %v", err)
		return
	}

	logger.Info("Scheduler restored from desired_running=true")
}

func (s *Server) listEnabledSchedules(ctx context.Context) ([]*models.Schedule, error) {
	enabled := true
	return s.db.ListSchedules(ctx, &enabled)
}

func (s *Server) persistSchedulerDesiredRunning(ctx context.Context, desiredRunning bool) error {
	return s.db.SetSchedulerDesiredRunning(ctx, desiredRunning)
}

func (s *Server) runSchedule(c *gin.Context) {
	id := c.Param("id")
	ctx, cancel := etcdRequestContext(c.Request.Context())
	defer cancel()

	workers, err := s.jobStore.ListWorkers(ctx)
	if err != nil {
		s.errorResponse(c, http.StatusServiceUnavailable, etcdUnavailablePrefix+err.Error())
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
