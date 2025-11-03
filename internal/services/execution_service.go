package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/llm"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

// ExecutionService provides business logic for prompt execution
type ExecutionService struct {
	db          db.Database
	llmRegistry *llm.Registry
}

// NewExecutionService creates a new execution service
func NewExecutionService(database db.Database, registry *llm.Registry) *ExecutionService {
	return &ExecutionService{
		db:          database,
		llmRegistry: registry,
	}
}

// ExecutionConfig represents configuration for prompt execution
type ExecutionConfig struct {
	Temperature float64       `json:"temperature"`
	MaxRetries  int           `json:"max_retries"`
	RetryDelay  time.Duration `json:"retry_delay"`
}

// DefaultExecutionConfig returns default execution configuration
func DefaultExecutionConfig() *ExecutionConfig {
	return &ExecutionConfig{
		Temperature: 0.7,
		MaxRetries:  3,
		RetryDelay:  30 * time.Second,
	}
}

// ExecutePromptWithLLM executes a prompt with a specific LLM
func (s *ExecutionService) ExecutePromptWithLLM(ctx context.Context, prompt *models.Prompt, llmConfig *models.LLMConfig, config *ExecutionConfig) (*models.Response, error) {
	if config == nil {
		config = DefaultExecutionConfig()
	}

	provider, ok := s.llmRegistry.Get(llmConfig.Provider)
	if !ok {
		return nil, fmt.Errorf("LLM provider %s not found", llmConfig.Provider)
	}

	var lastErr error
	for attempt := 1; attempt <= config.MaxRetries; attempt++ {
		response, err := provider.Generate(ctx, prompt.Template, llm.Config{
			Model:       llmConfig.Model,
			Temperature: config.Temperature,
			MaxTokens:   1000,
		})

		if err != nil {
			lastErr = fmt.Errorf("failed to generate response: %w", err)
			if attempt < config.MaxRetries {
				time.Sleep(config.RetryDelay)
				continue
			}
			return nil, lastErr
		}

		if response.Error != "" {
			lastErr = fmt.Errorf("LLM error: %s", response.Error)
			if attempt < config.MaxRetries {
				time.Sleep(config.RetryDelay)
				continue
			}
			return nil, lastErr
		}

		responseModel := &models.Response{
			ID:           uuid.New().String(),
			PromptID:     prompt.ID,
			LLMID:        llmConfig.ID,
			PromptText:   prompt.Template,
			ResponseText: response.Text,
			LLMName:      llmConfig.Name,
			LLMProvider:  llmConfig.Provider,
			LLMModel:     llmConfig.Model,
			Temperature:  config.Temperature,
			TokensUsed:   response.TokensUsed,
			LatencyMs:    response.LatencyMs,
			CreatedAt:    time.Now(),
		}

		if err := s.db.CreateResponse(ctx, responseModel); err != nil {
			return nil, fmt.Errorf("failed to save response: %w", err)
		}

		return responseModel, nil
	}

	return nil, fmt.Errorf("all %d attempts failed. Last error: %w", config.MaxRetries, lastErr)
}

// ExecuteSchedule executes all prompts in a schedule with all LLMs
func (s *ExecutionService) ExecuteSchedule(ctx context.Context, scheduleID string, config *ExecutionConfig) (*ExecutionResult, error) {
	scheduleService := NewScheduleService(s.db)
	plan, err := scheduleService.GetScheduleExecutionPlan(ctx, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution plan: %w", err)
	}

	result := &ExecutionResult{
		ScheduleID:           scheduleID,
		ScheduleName:         plan.ScheduleName,
		TotalExecutions:      plan.CalculateTotalExecutions(),
		SuccessfulExecutions: 0,
		FailedExecutions:     0,
		Responses:            make([]*models.Response, 0),
		Errors:               make([]ExecutionError, 0),
		StartTime:            time.Now(),
	}

	for _, prompt := range plan.Prompts {
		for _, llmConfig := range plan.LLMs {
			temperature := plan.Temperature
			if config != nil {
				temperature = config.Temperature
			}

			execConfig := &ExecutionConfig{
				Temperature: temperature,
				MaxRetries:  config.MaxRetries,
				RetryDelay:  config.RetryDelay,
			}

			response, err := s.ExecutePromptWithLLM(ctx, prompt, llmConfig, execConfig)
			if err != nil {
				result.FailedExecutions++
				result.Errors = append(result.Errors, ExecutionError{
					PromptID: prompt.ID,
					LLMID:    llmConfig.ID,
					Error:    err.Error(),
				})
			} else {
				result.SuccessfulExecutions++
				result.Responses = append(result.Responses, response)
			}
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if err := scheduleService.UpdateLastRun(ctx, scheduleID, result.StartTime); err != nil {
		fmt.Printf("Warning: failed to update schedule last run time: %v\n", err)
	}

	return result, nil
}

// ExecuteAllEnabledPrompts executes all enabled prompts with all enabled LLMs
func (s *ExecutionService) ExecuteAllEnabledPrompts(ctx context.Context, config *ExecutionConfig) (*ExecutionResult, error) {
	promptService := NewPromptManagementService(s.db)
	llmService := NewLLMService(s.db)

	prompts, err := promptService.GetEnabledPrompts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled prompts: %w", err)
	}

	llms, err := llmService.GetEnabledLLMs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled LLMs: %w", err)
	}

	if len(prompts) == 0 {
		return nil, fmt.Errorf("no enabled prompts found")
	}

	if len(llms) == 0 {
		return nil, fmt.Errorf("no enabled LLMs found")
	}

	result := &ExecutionResult{
		ScheduleID:           "manual-execution",
		ScheduleName:         "Manual Execution",
		TotalExecutions:      len(prompts) * len(llms),
		SuccessfulExecutions: 0,
		FailedExecutions:     0,
		Responses:            make([]*models.Response, 0),
		Errors:               make([]ExecutionError, 0),
		StartTime:            time.Now(),
	}

	for _, prompt := range prompts {
		for _, llmConfig := range llms {
			response, err := s.ExecutePromptWithLLM(ctx, prompt, llmConfig, config)
			if err != nil {
				result.FailedExecutions++
				result.Errors = append(result.Errors, ExecutionError{
					PromptID: prompt.ID,
					LLMID:    llmConfig.ID,
					Error:    err.Error(),
				})
			} else {
				result.SuccessfulExecutions++
				result.Responses = append(result.Responses, response)
			}
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

// ExecutionResult represents the result of an execution
type ExecutionResult struct {
	ScheduleID           string             `json:"schedule_id"`
	ScheduleName         string             `json:"schedule_name"`
	TotalExecutions      int                `json:"total_executions"`
	SuccessfulExecutions int                `json:"successful_executions"`
	FailedExecutions     int                `json:"failed_executions"`
	Responses            []*models.Response `json:"responses"`
	Errors               []ExecutionError   `json:"errors"`
	StartTime            time.Time          `json:"start_time"`
	EndTime              time.Time          `json:"end_time"`
	Duration             time.Duration      `json:"duration"`
}

// ExecutionError represents an error during execution
type ExecutionError struct {
	PromptID string `json:"prompt_id"`
	LLMID    string `json:"llm_id"`
	Error    string `json:"error"`
}

// ValidateTemperature validates temperature value
func ValidateTemperature(temperature float64) error {
	if temperature < 0.0 || temperature > 1.0 {
		return fmt.Errorf("temperature must be between 0.0 and 1.0, got: %.2f", temperature)
	}
	return nil
}

// ListResponses lists responses with filtering
func (s *ExecutionService) ListResponses(ctx context.Context, filter shared.ResponseFilter) ([]*models.Response, error) {
	return s.db.ListResponses(ctx, filter)
}

// GetResponse retrieves a response by ID
func (s *ExecutionService) GetResponse(ctx context.Context, id string) (*models.Response, error) {
	return s.db.GetResponse(ctx, id)
}

// DeleteResponse deletes a response - placeholder for future implementation
func (s *ExecutionService) DeleteResponse(ctx context.Context, id string) error {
	// TODO: Implement DeleteResponse in database interface
	return fmt.Errorf("delete response not implemented")
}

// DeleteAllResponses deletes all responses
func (s *ExecutionService) DeleteAllResponses(ctx context.Context) (int, error) {
	return s.db.DeleteAllResponses(ctx)
}
