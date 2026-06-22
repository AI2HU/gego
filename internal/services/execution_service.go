package services

import (
	"context"
	"time"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/jobs"
	"github.com/AI2HU/gego/internal/llm"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

type ExecutionService struct {
	db          db.Database
	llmRegistry *llm.Registry
	executor    *PromptExecutor
}

func NewExecutionService(database db.Database, registry *llm.Registry, rateLimiter jobs.RateLimiter) *ExecutionService {
	return &ExecutionService{
		db:          database,
		llmRegistry: registry,
		executor:    NewPromptExecutor(database, registry, rateLimiter),
	}
}

type ExecutionConfig struct {
	Temperature float64       `json:"temperature"`
	MaxRetries  int           `json:"max_retries"`
	RetryDelay  time.Duration `json:"retry_delay"`
}

func DefaultExecutionConfig() *ExecutionConfig {
	return &ExecutionConfig{
		Temperature: 0.7,
		MaxRetries:  3,
		RetryDelay:  30 * time.Second,
	}
}

func (s *ExecutionService) ExecutePromptWithLLM(ctx context.Context, prompt *models.Prompt, llmConfig *models.LLMConfig, config *ExecutionConfig) (*models.Response, error) {
	if config == nil {
		config = DefaultExecutionConfig()
	}

	exec := s.executor.WithRetries(config.MaxRetries, config.RetryDelay)
	responseID, err := exec.Execute(ctx, PromptExecutionParams{
		Temperature: config.Temperature,
	}, prompt, llmConfig)
	if err != nil {
		return nil, err
	}
	return s.db.GetResponse(ctx, responseID)
}

func (s *ExecutionService) ListResponses(ctx context.Context, filter shared.ResponseFilter) ([]*models.Response, error) {
	return s.db.ListResponses(ctx, filter)
}

func (s *ExecutionService) GetResponse(ctx context.Context, id string) (*models.Response, error) {
	return s.db.GetResponse(ctx, id)
}

func (s *ExecutionService) DeleteAllResponses(ctx context.Context) (int, error) {
	return s.db.DeleteAllResponses(ctx)
}
