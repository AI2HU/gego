package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/jobs"
	"github.com/AI2HU/gego/internal/llm"
	"github.com/AI2HU/gego/internal/logger"
	"github.com/AI2HU/gego/internal/models"
)

const (
	DefaultMaxRetries = 3
	DefaultRetryDelay = 30 * time.Second
)

type PromptExecutor struct {
	db          db.Database
	llmRegistry *llm.Registry
	rateLimiter jobs.RateLimiter
	maxRetries  int
	retryDelay  time.Duration
}

func NewPromptExecutor(database db.Database, registry *llm.Registry, rateLimiter jobs.RateLimiter) *PromptExecutor {
	return &PromptExecutor{
		db:          database,
		llmRegistry: registry,
		rateLimiter: rateLimiter,
		maxRetries:  DefaultMaxRetries,
		retryDelay:  DefaultRetryDelay,
	}
}

func (e *PromptExecutor) WithRetries(maxRetries int, retryDelay time.Duration) *PromptExecutor {
	if maxRetries > 0 {
		e.maxRetries = maxRetries
	}
	if retryDelay > 0 {
		e.retryDelay = retryDelay
	}
	return e
}

type PromptExecutionParams struct {
	ScheduleID  string
	RunID       string
	JobID       string
	Temperature float64
}

func (e *PromptExecutor) Execute(ctx context.Context, params PromptExecutionParams, prompt *models.Prompt, llmConfig *models.LLMConfig) (string, error) {
	var lastErr error

	for attempt := 1; attempt <= e.maxRetries; attempt++ {
		responseID, err := e.executeOnce(ctx, params, prompt, llmConfig)
		if err == nil {
			return responseID, nil
		}

		lastErr = err
		logger.Warning("job=%s run=%s attempt %d/%d failed: %v", params.JobID, params.RunID, attempt, e.maxRetries, err)

		retryDelay := e.retryDelay
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "quota") || strings.Contains(err.Error(), "rate limit") {
			retryDelay = 2 * time.Minute
		}

		if attempt < e.maxRetries {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(retryDelay):
			}
		}
	}

	return "", fmt.Errorf("failed after %d attempts: %w", e.maxRetries, lastErr)
}

func (e *PromptExecutor) executeOnce(ctx context.Context, params PromptExecutionParams, prompt *models.Prompt, llmConfig *models.LLMConfig) (string, error) {
	logger.Info("executing job=%s run=%s schedule=%s prompt=%s llm=%s provider=%s",
		params.JobID, params.RunID, params.ScheduleID, prompt.ID, llmConfig.ID, llmConfig.Provider)

	provider, ok := e.llmRegistry.Get(llmConfig.Provider)
	if !ok {
		return "", fmt.Errorf("provider not found: %s", llmConfig.Provider)
	}

	if e.rateLimiter != nil {
		if err := e.rateLimiter.Wait(ctx, llmConfig.Provider); err != nil {
			return "", fmt.Errorf("rate limiter wait failed: %w", err)
		}
	}

	llmConfigStruct := llm.Config{
		Model:       llmConfig.Model,
		Temperature: params.Temperature,
		MaxTokens:   1000,
	}
	applyLLMConfigOverrides(llmConfig, &llmConfigStruct)

	startTime := time.Now()
	resp, err := provider.Generate(ctx, prompt.Template, llmConfigStruct)
	duration := time.Since(startTime)

	responseID := uuid.New().String()
	response := &models.Response{
		ID:          responseID,
		PromptID:    prompt.ID,
		PromptText:  prompt.Template,
		LLMID:       llmConfig.ID,
		LLMName:     llmConfig.Name,
		LLMProvider: llmConfig.Provider,
		LLMModel:    llmConfig.Model,
		Temperature: params.Temperature,
		ScheduleID:  params.ScheduleID,
		RunID:       params.RunID,
		JobID:       params.JobID,
		CreatedAt:   time.Now(),
	}

	if err != nil {
		logger.Error("llm call failed after %v job=%s: %v", duration, params.JobID, err)
		response.Error = err.Error()
		if saveErr := e.db.CreateResponse(ctx, response); saveErr != nil {
			return "", saveErr
		}
		return "", err
	}

	response.ResponseText = resp.Text
	response.TokensUsed = resp.TokensUsed
	response.Error = resp.Error

	if len(resp.SearchURLs) > 0 {
		response.SearchURLs = make([]models.SearchURL, len(resp.SearchURLs))
		for i, url := range resp.SearchURLs {
			response.SearchURLs[i] = models.SearchURL{
				SearchQuery:   url.SearchQuery,
				URL:           url.URL,
				Title:         url.Title,
				CitationIndex: url.CitationIndex,
			}
		}
	}

	if err := e.db.CreateResponse(ctx, response); err != nil {
		return "", err
	}

	logger.Info("llm call succeeded after %v job=%s response=%s", duration, params.JobID, responseID)
	return responseID, nil
}

func applyLLMConfigOverrides(llmConfig *models.LLMConfig, cfg *llm.Config) {
	if llmConfig.Config == nil {
		return
	}
	if tempStr, ok := llmConfig.Config["temperature"]; ok {
		if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
			cfg.Temperature = temp
		}
	}
	if maxTokensStr, ok := llmConfig.Config["max_tokens"]; ok {
		if maxTokens, err := strconv.Atoi(maxTokensStr); err == nil && maxTokens >= 1 {
			cfg.MaxTokens = maxTokens
		}
	}
	if topPStr, ok := llmConfig.Config["top_p"]; ok {
		if topP, err := strconv.ParseFloat(topPStr, 64); err == nil {
			cfg.TopP = topP
		}
	}
	if topKStr, ok := llmConfig.Config["top_k"]; ok {
		if topK, err := strconv.Atoi(topKStr); err == nil {
			cfg.TopK = topK
		}
	}
	if streamStr, ok := llmConfig.Config["stream"]; ok {
		if stream, err := strconv.ParseBool(streamStr); err == nil {
			cfg.Stream = stream
		}
	}
}

func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "(not set)"
	}
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}
