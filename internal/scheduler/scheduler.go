package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/llm"
	"github.com/AI2HU/gego/internal/logger"
	"github.com/AI2HU/gego/internal/models"
)

// Retry configuration constants
const (
	DefaultMaxRetries = 3
	DefaultRetryDelay = 30 * time.Second
)

// Scheduler manages scheduled prompt executions
type Scheduler struct {
	db          db.Database
	llmRegistry *llm.Registry
	cron        *cron.Cron
	running     bool
	mu          sync.RWMutex
}

// New creates a new scheduler
func New(database db.Database, llmRegistry *llm.Registry) *Scheduler {
	return &Scheduler{
		db:          database,
		llmRegistry: llmRegistry,
		cron:        cron.New(),
	}
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler already running")
	}

	// Load all enabled schedules
	schedules, err := s.db.ListSchedules(ctx, boolPtr(true))
	if err != nil {
		return fmt.Errorf("failed to load schedules: %w", err)
	}

	// Register each schedule with cron
	for _, schedule := range schedules {
		if err := s.registerSchedule(ctx, schedule); err != nil {
			logger.Error("Failed to register schedule %s: %v", schedule.ID, err)
		}
	}

	s.cron.Start()
	s.running = true

	logger.Info("Scheduler started")
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.cron.Stop()
	s.running = false

	logger.Info("Scheduler stopped")
}

// registerSchedule registers a schedule with cron
func (s *Scheduler) registerSchedule(ctx context.Context, schedule *models.Schedule) error {
	_, err := s.cron.AddFunc(schedule.CronExpr, func() {
		if err := s.executeSchedule(context.Background(), schedule); err != nil {
			logger.Error("Failed to execute schedule %s: %v", schedule.ID, err)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	logger.Info("Registered schedule %s with cron expression: %s", schedule.ID, schedule.CronExpr)
	return nil
}

// executeSchedule executes a schedule
func (s *Scheduler) executeSchedule(ctx context.Context, schedule *models.Schedule) error {
	logger.Info("Executing schedule: %s", schedule.ID)
	logger.Info("Schedule has %d prompts and %d LLMs", len(schedule.PromptIDs), len(schedule.LLMIDs))

	// Get prompts
	prompts := make([]*models.Prompt, 0, len(schedule.PromptIDs))
	for _, promptID := range schedule.PromptIDs {
		logger.Debug("Getting prompt: %s", promptID)
		prompt, err := s.db.GetPrompt(ctx, promptID)
		if err != nil {
			logger.Error("Failed to get prompt %s: %v", promptID, err)
			continue
		}
		logger.Debug("Retrieved prompt: %s (%s)", prompt.Template, prompt.ID)
		prompts = append(prompts, prompt)
	}

	// Get LLMs
	llms := make([]*models.LLMConfig, 0, len(schedule.LLMIDs))
	for _, llmID := range schedule.LLMIDs {
		logger.Debug("Getting LLM: %s", llmID)
		llmConfig, err := s.db.GetLLM(ctx, llmID)
		if err != nil {
			logger.Error("Failed to get LLM %s: %v", llmID, err)
			continue
		}
		if !llmConfig.Enabled {
			logger.Warning("LLM %s is disabled, skipping", llmConfig.Name)
			continue
		}
		logger.Debug("Retrieved LLM: %s (%s) - API Key: %s", llmConfig.Name, llmConfig.ID, maskAPIKey(llmConfig.APIKey))
		llms = append(llms, llmConfig)
	}

	logger.Info("Found %d prompts and %d enabled LLMs", len(prompts), len(llms))

	// Execute each prompt against each LLM
	var wg sync.WaitGroup
	executionCount := 0
	for _, prompt := range prompts {
		for _, llmConfig := range llms {
			wg.Add(1)
			executionCount++
			go func(p *models.Prompt, l *models.LLMConfig) {
				defer wg.Done()
				logger.Debug("Executing prompt '%s' with LLM '%s'", p.Template, l.Name)

				// Use retry mechanism: 3 attempts with 30-second delays
				if err := s.executePromptWithRetry(ctx, schedule.ID, p, l, DefaultMaxRetries, DefaultRetryDelay); err != nil {
					logger.Error("Failed to execute prompt %s with LLM %s after all retries: %v", p.ID, l.ID, err)
				} else {
					logger.Debug("Successfully executed prompt %s with LLM %s", p.ID, l.ID)
				}
			}(prompt, llmConfig)
		}
	}

	logger.Info("Starting %d concurrent executions", executionCount)
	wg.Wait()
	logger.Info("Completed %d executions", executionCount)

	// Update schedule last run time
	now := time.Now()
	schedule.LastRun = &now
	if err := s.db.UpdateSchedule(ctx, schedule); err != nil {
		logger.Error("Failed to update schedule last run: %v", err)
	}

	logger.Info("Completed schedule: %s", schedule.ID)
	return nil
}

// maskAPIKey masks the API key for logging (shows first 4 and last 4 characters)
func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "(not set)"
	}
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// executePromptWithRetry executes a prompt with retry mechanism
func (s *Scheduler) executePromptWithRetry(ctx context.Context, scheduleID string, prompt *models.Prompt, llmConfig *models.LLMConfig, maxRetries int, retryDelay time.Duration) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger.Debug("Attempt %d/%d for prompt '%s' with LLM '%s'", attempt, maxRetries, prompt.Template[:min(50, len(prompt.Template))]+"...", llmConfig.Name)

		err := s.executePromptWithLLM(ctx, scheduleID, prompt, llmConfig)
		if err == nil {
			if attempt > 1 {
				logger.Info("✅ Prompt execution succeeded on attempt %d after %d previous failures", attempt, attempt-1)
			}
			return nil
		}

		lastErr = err
		logger.Warning("❌ Attempt %d/%d failed for prompt '%s' with LLM '%s': %v", attempt, maxRetries, prompt.Template[:min(50, len(prompt.Template))]+"...", llmConfig.Name, err)

		// Don't wait after the last attempt
		if attempt < maxRetries {
			logger.Info("⏳ Waiting %v before retry attempt %d...", retryDelay, attempt+1)
			time.Sleep(retryDelay)
		}
	}

	logger.Error("💥 All %d attempts failed for prompt '%s' with LLM '%s'. Last error: %v", maxRetries, prompt.Template[:min(50, len(prompt.Template))]+"...", llmConfig.Name, lastErr)
	return fmt.Errorf("failed after %d attempts, last error: %w", maxRetries, lastErr)
}

// executePromptWithLLM executes a single prompt with a single LLM
func (s *Scheduler) executePromptWithLLM(ctx context.Context, scheduleID string, prompt *models.Prompt, llmConfig *models.LLMConfig) error {
	logger.Info("Starting execution: prompt='%s' LLM='%s' provider='%s'", prompt.Template, llmConfig.Name, llmConfig.Provider)

	// Get LLM provider
	provider, ok := s.llmRegistry.Get(llmConfig.Provider)
	if !ok {
		logger.Error("Provider not found: %s", llmConfig.Provider)
		return fmt.Errorf("provider not found: %s", llmConfig.Provider)
	}
	logger.Debug("Found provider for: %s", llmConfig.Provider)

	// Prepare config
	config := make(map[string]interface{})
	config["model"] = llmConfig.Model
	config["api_key"] = llmConfig.APIKey
	if llmConfig.BaseURL != "" {
		config["base_url"] = llmConfig.BaseURL
	}
	for k, v := range llmConfig.Config {
		config[k] = v
	}
	logger.Debug("Prepared config for LLM: model=%s api_key=%s base_url=%s", llmConfig.Model, maskAPIKey(llmConfig.APIKey), llmConfig.BaseURL)

	// Generate response
	logger.Debug("Calling LLM provider with prompt: %s", prompt.Template[:min(50, len(prompt.Template))]+"...")
	startTime := time.Now()
	resp, err := provider.Generate(ctx, prompt.Template, config)
	duration := time.Since(startTime)

	if err != nil {
		logger.Error("LLM call failed after %v: %v", duration, err)
		// Store error response
		response := &models.Response{
			ID:          uuid.New().String(),
			PromptID:    prompt.ID,
			PromptText:  prompt.Template,
			LLMID:       llmConfig.ID,
			LLMName:     llmConfig.Name,
			LLMProvider: llmConfig.Provider,
			LLMModel:    llmConfig.Model,
			Error:       err.Error(),
			ScheduleID:  scheduleID,
			LatencyMs:   time.Since(startTime).Milliseconds(),
			CreatedAt:   time.Now(),
		}
		return s.db.CreateResponse(ctx, response)
	}

	logger.Info("LLM call succeeded after %v, response length: %d", duration, len(resp.Text))

	// Create response record
	response := &models.Response{
		ID:           uuid.New().String(),
		PromptID:     prompt.ID,
		PromptText:   prompt.Template,
		LLMID:        llmConfig.ID,
		LLMName:      llmConfig.Name,
		LLMProvider:  llmConfig.Provider,
		LLMModel:     llmConfig.Model,
		ResponseText: resp.Text,
		ScheduleID:   scheduleID,
		TokensUsed:   resp.TokensUsed,
		LatencyMs:    resp.LatencyMs,
		Error:        resp.Error,
		CreatedAt:    time.Now(),
	}

	return s.db.CreateResponse(ctx, response)
}

// ExecuteNow executes a schedule immediately
func (s *Scheduler) ExecuteNow(ctx context.Context, scheduleID string) error {
	schedule, err := s.db.GetSchedule(ctx, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	return s.executeSchedule(ctx, schedule)
}

// ExecutePrompt executes a single prompt with specified LLMs
func (s *Scheduler) ExecutePrompt(ctx context.Context, promptID string, llmIDs []string) error {
	prompt, err := s.db.GetPrompt(ctx, promptID)
	if err != nil {
		return fmt.Errorf("failed to get prompt: %w", err)
	}

	llms := make([]*models.LLMConfig, 0, len(llmIDs))
	for _, llmID := range llmIDs {
		llmConfig, err := s.db.GetLLM(ctx, llmID)
		if err != nil {
			logger.Error("Failed to get LLM %s: %v", llmID, err)
			continue
		}
		llms = append(llms, llmConfig)
	}

	var wg sync.WaitGroup
	for _, llmConfig := range llms {
		wg.Add(1)
		go func(l *models.LLMConfig) {
			defer wg.Done()
			// Use retry mechanism: 3 attempts with 30-second delays
			if err := s.executePromptWithRetry(ctx, "", prompt, l, DefaultMaxRetries, DefaultRetryDelay); err != nil {
				logger.Error("Failed to execute prompt %s with LLM %s after all retries: %v", prompt.ID, l.ID, err)
			}
		}(llmConfig)
	}

	wg.Wait()
	return nil
}

// Reload reloads all schedules
func (s *Scheduler) Reload(ctx context.Context) error {
	s.Stop()
	time.Sleep(100 * time.Millisecond) // Give it time to stop
	return s.Start(ctx)
}

func boolPtr(b bool) *bool {
	return &b
}
