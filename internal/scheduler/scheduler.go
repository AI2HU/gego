package scheduler

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"golang.org/x/time/rate"

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

// Rate limiting configuration
const (
	// 6 requests per minute = 1 request every 10 seconds
	RequestsPerMinute = 6
	RateLimitBurst    = 1
)

// Scheduler manages scheduled prompt executions using robfig/cron
type Scheduler struct {
	db          db.Database
	llmRegistry *llm.Registry
	cron        *cron.Cron
	running     bool
	mu          sync.RWMutex
	// Rate limiters per LLM provider (keyed by provider name)
	rateLimiters map[string]*rate.Limiter
	rateMu       sync.RWMutex
	// Track registered schedule IDs for management
	scheduleEntries map[string]cron.EntryID
	entriesMu       sync.RWMutex
}

// New creates a new scheduler with proper cron configuration
func New(database db.Database, llmRegistry *llm.Registry) *Scheduler {
	// Create cron with proper configuration
	c := cron.New(
		cron.WithLocation(time.UTC),
		cron.WithLogger(cron.DefaultLogger),
		cron.WithChain(
			cron.Recover(cron.DefaultLogger), // Recover from panics
		),
	)

	return &Scheduler{
		db:              database,
		llmRegistry:     llmRegistry,
		cron:            c,
		rateLimiters:    make(map[string]*rate.Limiter),
		scheduleEntries: make(map[string]cron.EntryID),
	}
}

// Start starts the scheduler and loads all enabled schedules
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

	if len(schedules) == 0 {
		logger.Info("No enabled schedules found. Scheduler is running but will not execute any tasks.")
		logger.Info("Use 'gego schedule add' to create schedules or 'gego schedule list' to check existing schedules.")
	} else {
		logger.Info("Loaded %d enabled schedule(s)", len(schedules))
	}

	// Register each schedule with cron
	registeredCount := 0
	for _, schedule := range schedules {
		if err := s.registerSchedule(ctx, schedule); err != nil {
			logger.Error("Failed to register schedule %s: %v", schedule.ID, err)
		} else {
			registeredCount++
		}
	}

	if len(schedules) > 0 {
		logger.Info("Successfully registered %d schedule(s) with cron", registeredCount)
	}

	// Start the cron scheduler
	s.cron.Start()
	s.running = true

	logger.Info("Scheduler started successfully")
	return nil
}

// Stop stops the scheduler and removes all registered schedules
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	// Stop the cron scheduler
	s.cron.Stop()
	s.running = false

	// Clear all schedule entries
	s.entriesMu.Lock()
	s.scheduleEntries = make(map[string]cron.EntryID)
	s.entriesMu.Unlock()

	logger.Info("Scheduler stopped")
}

// GetStatus returns the current status of the scheduler
func (s *Scheduler) GetStatus(ctx context.Context) (bool, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running {
		return false, 0, nil
	}

	// Get count of enabled schedules from database
	schedules, err := s.db.ListSchedules(ctx, boolPtr(true))
	if err != nil {
		return s.running, 0, fmt.Errorf("failed to get schedule count: %w", err)
	}

	return s.running, len(schedules), nil
}

// registerSchedule registers a schedule with cron and stores the entry ID
func (s *Scheduler) registerSchedule(ctx context.Context, schedule *models.Schedule) error {
	// Create a job function that executes the schedule
	jobFunc := func() {
		logger.Info("Executing scheduled job: %s", schedule.Name)
		if err := s.executeSchedule(context.Background(), schedule); err != nil {
			logger.Error("Failed to execute schedule %s: %v", schedule.ID, err)
		}
	}

	// Add the job to cron and get the entry ID
	entryID, err := s.cron.AddFunc(schedule.CronExpr, jobFunc)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	// Store the entry ID for this schedule
	s.entriesMu.Lock()
	s.scheduleEntries[schedule.ID] = entryID
	s.entriesMu.Unlock()

	logger.Info("Registered schedule %s with cron expression: %s (Entry ID: %d)", schedule.ID, schedule.CronExpr, entryID)
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

				// Generate random temperature for each prompt if random was selected
				currentTemperature := schedule.Temperature
				if schedule.Temperature == -1.0 { // Special value indicating "random" was selected
					rand.Seed(time.Now().UnixNano())
					currentTemperature = rand.Float64()
					logger.Debug("Generated random temperature %.1f for prompt '%s'", currentTemperature, p.Template)
				}

				// Use retry mechanism: 3 attempts with 30-second delays
				if err := s.executePromptWithRetry(ctx, schedule.ID, p, l, currentTemperature, DefaultMaxRetries, DefaultRetryDelay); err != nil {
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
func (s *Scheduler) executePromptWithRetry(ctx context.Context, scheduleID string, prompt *models.Prompt, llmConfig *models.LLMConfig, temperature float64, maxRetries int, retryDelay time.Duration) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger.Debug("Attempt %d/%d for prompt '%s' with LLM '%s'", attempt, maxRetries, prompt.Template[:min(50, len(prompt.Template))]+"...", llmConfig.Name)

		err := s.executePromptWithLLM(ctx, scheduleID, prompt, llmConfig, temperature)
		if err == nil {
			if attempt > 1 {
				logger.Info("✅ Prompt execution succeeded on attempt %d after %d previous failures", attempt, attempt-1)
			}
			return nil
		}

		lastErr = err
		logger.Warning("❌ Attempt %d/%d failed for prompt '%s' with LLM '%s': %v", attempt, maxRetries, prompt.Template[:min(50, len(prompt.Template))]+"...", llmConfig.Name, err)

		// Check if it's a rate limiting error - use longer delay
		retryDelayToUse := retryDelay
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "quota") || strings.Contains(err.Error(), "rate limit") {
			retryDelayToUse = 2 * time.Minute // Wait 2 minutes for rate limit errors
			logger.Info("Rate limit detected, using extended retry delay: %v", retryDelayToUse)
		}

		// Don't wait after the last attempt
		if attempt < maxRetries {
			logger.Info("⏳ Waiting %v before retry attempt %d...", retryDelayToUse, attempt+1)
			time.Sleep(retryDelayToUse)
		}
	}

	logger.Error("💥 All %d attempts failed for prompt '%s' with LLM '%s'. Last error: %v", maxRetries, prompt.Template[:min(50, len(prompt.Template))]+"...", llmConfig.Name, lastErr)
	return fmt.Errorf("failed after %d attempts, last error: %w", maxRetries, lastErr)
}

// executePromptWithLLM executes a single prompt with a single LLM
func (s *Scheduler) executePromptWithLLM(ctx context.Context, scheduleID string, prompt *models.Prompt, llmConfig *models.LLMConfig, temperature float64) error {
	logger.Info("Starting execution: prompt='%s' LLM='%s' provider='%s' temperature=%.2f", prompt.Template, llmConfig.Name, llmConfig.Provider, temperature)

	// Get LLM provider
	provider, ok := s.llmRegistry.Get(llmConfig.Provider)
	if !ok {
		logger.Error("Provider not found: %s", llmConfig.Provider)
		return fmt.Errorf("provider not found: %s", llmConfig.Provider)
	}
	logger.Debug("Found provider for: %s", llmConfig.Provider)

	// Get rate limiter for this provider
	rateLimiter := s.getRateLimiter(llmConfig.Provider)

	// Wait for rate limiter before making the request
	logger.Debug("Waiting for rate limiter for provider: %s", llmConfig.Provider)
	if err := rateLimiter.Wait(ctx); err != nil {
		logger.Error("Rate limiter wait failed: %v", err)
		return fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// Prepare config
	config := make(map[string]interface{})
	config["model"] = llmConfig.Model
	config["temperature"] = temperature
	config["api_key"] = llmConfig.APIKey
	if llmConfig.BaseURL != "" {
		config["base_url"] = llmConfig.BaseURL
	}
	for k, v := range llmConfig.Config {
		config[k] = v
	}
	logger.Debug("Prepared config for LLM: model=%s temperature=%.2f api_key=%s base_url=%s", llmConfig.Model, temperature, maskAPIKey(llmConfig.APIKey), llmConfig.BaseURL)

	// Generate response
	logger.Debug("[%s] Calling LLM provider with prompt: %s", llmConfig.Name, prompt.Template[:min(50, len(prompt.Template))]+"...")
	startTime := time.Now()
	resp, err := provider.Generate(ctx, prompt.Template, config)
	duration := time.Since(startTime)

	if err != nil {
		logger.Error("[%s] LLM call failed after %v: %v", llmConfig.Name, duration, err)
		// Store error response
		response := &models.Response{
			ID:          uuid.New().String(),
			PromptID:    prompt.ID,
			PromptText:  prompt.Template,
			LLMID:       llmConfig.ID,
			LLMName:     llmConfig.Name,
			LLMProvider: llmConfig.Provider,
			LLMModel:    llmConfig.Model,
			Temperature: temperature,
			Error:       err.Error(),
			ScheduleID:  scheduleID,
			LatencyMs:   time.Since(startTime).Milliseconds(),
			CreatedAt:   time.Now(),
		}
		return s.db.CreateResponse(ctx, response)
	}

	logger.Info("[%s] LLM call succeeded after %v, response length: %d", llmConfig.Name, duration, len(resp.Text))

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
		Temperature:  temperature,
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
			if err := s.executePromptWithRetry(ctx, "", prompt, l, 0.7, DefaultMaxRetries, DefaultRetryDelay); err != nil {
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

// getRateLimiter gets or creates a rate limiter for the given provider
func (s *Scheduler) getRateLimiter(provider string) *rate.Limiter {
	s.rateMu.RLock()
	limiter, exists := s.rateLimiters[provider]
	s.rateMu.RUnlock()

	if exists {
		return limiter
	}

	// Create new rate limiter
	s.rateMu.Lock()
	defer s.rateMu.Unlock()

	// Double-check in case another goroutine created it
	if limiter, exists := s.rateLimiters[provider]; exists {
		return limiter
	}

	// Create rate limiter: 6 requests per minute = 1 request every 10 seconds
	limiter = rate.NewLimiter(rate.Every(time.Minute/RequestsPerMinute), RateLimitBurst)
	s.rateLimiters[provider] = limiter
	return limiter
}

func boolPtr(b bool) *bool {
	return &b
}
