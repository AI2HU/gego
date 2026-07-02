package models

import (
	"time"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// CreateLLMRequest represents the request to create a new LLM
type CreateLLMRequest struct {
	Name               string            `json:"name" binding:"required"`
	Provider           string            `json:"provider" binding:"required"`
	Model              string            `json:"model" binding:"required"`
	APIKey             string            `json:"api_key,omitempty"`
	ExistingKeyIndex   *int              `json:"existing_key_index,omitempty"`
	BaseURL            string            `json:"base_url,omitempty"`
	Config             map[string]string `json:"config,omitempty"`
	Enabled            bool              `json:"enabled"`
}

// UpdateLLMRequest represents the request to update an existing LLM
type UpdateLLMRequest struct {
	Name     string            `json:"name,omitempty"`
	Provider string            `json:"provider,omitempty"`
	Model    string            `json:"model,omitempty"`
	APIKey   string            `json:"api_key,omitempty"`
	BaseURL  string            `json:"base_url,omitempty"`
	Config   map[string]string `json:"config,omitempty"`
	Enabled  *bool             `json:"enabled,omitempty"`
}

// ProviderResponse represents an available LLM provider
type ProviderResponse struct {
	ID              string `json:"id"`
	DisplayName     string `json:"display_name"`
	ConsoleURL      string `json:"console_url,omitempty"`
	RequiresAPIKey  bool   `json:"requires_api_key"`
	RequiresBaseURL bool   `json:"requires_base_url"`
}

// ProviderAPIKeyResponse represents a masked existing API key
type ProviderAPIKeyResponse struct {
	Index  int    `json:"index"`
	Masked string `json:"masked"`
}

// ListProviderModelsRequest represents credentials for listing provider models
type ListProviderModelsRequest struct {
	APIKey             string `json:"api_key,omitempty"`
	ExistingKeyIndex   *int   `json:"existing_key_index,omitempty"`
	BaseURL            string `json:"base_url,omitempty"`
}

// ModelInfoResponse represents an available model from a provider
type ModelInfoResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	UsedInChat  bool   `json:"used_in_chat,omitempty"`
}

// TestLLMResponse represents the result of testing model access
type TestLLMResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// LLMResponse represents the response for LLM operations
type LLMResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Provider  string            `json:"provider"`
	Model     string            `json:"model"`
	APIKey    string            `json:"api_key,omitempty"`
	BaseURL   string            `json:"base_url,omitempty"`
	Config    map[string]string `json:"config,omitempty"`
	Enabled   bool              `json:"enabled"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// CreatePromptRequest represents the request to create a new prompt
type CreatePromptRequest struct {
	Template string   `json:"template" binding:"required"`
	Tags     []string `json:"tags,omitempty"`
	Enabled  bool     `json:"enabled"`
}

// UpdatePromptRequest represents the request to update an existing prompt
type UpdatePromptRequest struct {
	Template string   `json:"template,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Enabled  *bool    `json:"enabled,omitempty"`
}

// PromptResponse represents the response for prompt operations
type PromptResponse struct {
	ID        string    `json:"id"`
	Template  string    `json:"template"`
	Tags      []string  `json:"tags,omitempty"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GeneratePromptsRequest represents the request to generate prompts using an LLM
type GeneratePromptsRequest struct {
	LLMID        string `json:"llm_id" binding:"required"`
	LanguageCode string `json:"language_code" binding:"required"`
	UserInput    string `json:"user_input" binding:"required"`
	PromptCount  int    `json:"prompt_count"`
}

// GeneratePromptsResponse represents generated prompt templates
type GeneratePromptsResponse struct {
	Prompts []string `json:"prompts"`
}

// BulkCreatePromptsRequest represents a bulk prompt create request
type BulkCreatePromptsRequest struct {
	Prompts []BulkCreatePromptItem `json:"prompts" binding:"required,min=1,dive"`
	Tags    []string               `json:"tags,omitempty"`
}

// BulkCreatePromptItem represents one prompt in a bulk create request
type BulkCreatePromptItem struct {
	Template string `json:"template" binding:"required"`
}

// BulkCreatePromptsResponse represents saved prompts from a bulk create
type BulkCreatePromptsResponse struct {
	Prompts    []PromptResponse `json:"prompts"`
	SavedCount int              `json:"saved_count"`
}

// CreateScheduleRequest represents the request to create a new schedule
type CreateScheduleRequest struct {
	Name        string   `json:"name" binding:"required"`
	PromptIDs   []string `json:"prompt_ids" binding:"required"`
	LLMIDs      []string `json:"llm_ids" binding:"required"`
	CronExpr    string   `json:"cron_expr" binding:"required"`
	Temperature float64  `json:"temperature,omitempty"`
	Enabled     bool     `json:"enabled"`
}

// UpdateScheduleRequest represents the request to update an existing schedule
type UpdateScheduleRequest struct {
	Name        string   `json:"name,omitempty"`
	PromptIDs   []string `json:"prompt_ids,omitempty"`
	LLMIDs      []string `json:"llm_ids,omitempty"`
	CronExpr    string   `json:"cron_expr,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	Enabled     *bool    `json:"enabled,omitempty"`
}

// SchedulerStatusResponse represents the current scheduler state
type SchedulerStatusResponse struct {
	Running          bool `json:"running"`
	EnabledSchedules int  `json:"enabled_schedules"`
	IsLeader         bool `json:"is_leader"`
	PendingJobs      int  `json:"pending_jobs"`
	ActiveRuns       int  `json:"active_runs"`
	ActiveWorkers    int  `json:"active_workers"`
}

type ScheduleRunEnqueueResponse struct {
	RunID string `json:"run_id"`
}

type ScheduleRunResponse struct {
	ID            string     `json:"id"`
	ScheduleID    string     `json:"schedule_id"`
	Trigger       string     `json:"trigger"`
	Status        string     `json:"status"`
	TotalJobs     int        `json:"total_jobs"`
	CompletedJobs int        `json:"completed_jobs"`
	FailedJobs    int        `json:"failed_jobs"`
	CreatedAt     time.Time  `json:"created_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
}

type ScheduleJobResponse struct {
	ID          string     `json:"id"`
	RunID       string     `json:"run_id"`
	ScheduleID  string     `json:"schedule_id"`
	PromptID    string     `json:"prompt_id"`
	LLMID       string     `json:"llm_id"`
	Provider    string     `json:"provider"`
	Temperature float64    `json:"temperature"`
	Status      string     `json:"status"`
	Attempts    int        `json:"attempts"`
	MaxAttempts int        `json:"max_attempts"`
	WorkerID    string     `json:"worker_id,omitempty"`
	ResponseID  string     `json:"response_id,omitempty"`
	Error       string     `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ClaimedAt   *time.Time `json:"claimed_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type ScheduleRunListResponse struct {
	Data       []ScheduleRunResponse `json:"data"`
	NextCursor string                `json:"next_cursor,omitempty"`
}

type ScheduleRunDetailResponse struct {
	Run  ScheduleRunResponse  `json:"run"`
	Jobs []ScheduleJobResponse `json:"jobs"`
}

// ScheduleResponse represents the response for schedule operations
type ScheduleResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	PromptIDs   []string   `json:"prompt_ids"`
	LLMIDs      []string   `json:"llm_ids"`
	CronExpr    string     `json:"cron_expr"`
	Temperature float64    `json:"temperature"`
	Enabled     bool       `json:"enabled"`
	LastRun     *time.Time `json:"last_run,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// StatsResponse represents the response for statistics
type StatsResponse struct {
	TotalResponses int64             `json:"total_responses"`
	TotalPrompts   int64             `json:"total_prompts"`
	TotalLLMs      int64             `json:"total_llms"`
	TotalSchedules int64             `json:"total_schedules"`
	TopKeywords    []KeywordCount      `json:"top_keywords"`
	BrandTrends    []BrandTrendSeries  `json:"brand_trends"`
	PromptStats    []*PromptStats      `json:"prompt_stats"`
	LLMStats       []*LLMStats       `json:"llm_stats"`
	ResponseTrends []TimeSeriesPoint `json:"response_trends"`
	LastUpdated    time.Time         `json:"last_updated"`
}

// SearchRequest represents the request to search responses
type SearchRequest struct {
	Keyword   string     `json:"keyword" binding:"required"`
	Tags      []string   `json:"tags,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Limit     int        `json:"limit,omitempty"`
}

// SearchResponse represents the response for search operations
type SearchResponse struct {
	Keyword         string         `json:"keyword"`
	TotalResponses  int64          `json:"total_responses"`
	TotalMentions   int            `json:"total_mentions"`
	UniquePrompts   int            `json:"unique_prompts"`
	UniqueLLMs      int            `json:"unique_llms"`
	ByPrompt        map[string]int `json:"by_prompt"`
	ByLLM           map[string]int `json:"by_llm"`
	ByProvider      map[string]int `json:"by_provider"`
	FirstSeen       time.Time      `json:"first_seen"`
	LastSeen        time.Time      `json:"last_seen"`
	Responses       []*Response    `json:"responses,omitempty"`
}

// ErrorLogResponse represents a failed LLM call stored in responses
type ErrorLogResponse struct {
	ID          string    `json:"id"`
	PromptID    string    `json:"prompt_id"`
	PromptText  string    `json:"prompt_text"`
	LLMID       string    `json:"llm_id"`
	LLMName     string    `json:"llm_name"`
	LLMProvider string    `json:"llm_provider"`
	LLMModel    string    `json:"llm_model"`
	Error       string    `json:"error"`
	ScheduleID  string    `json:"schedule_id,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type UpgradeItem struct {
	Code   string `json:"code"`
	Severity string `json:"severity"`
}

type RunUpgradeRequest struct {
	UpgradeCode string `json:"upgrade_code" binding:"required"`
}

type RunUpgradeResponse struct {
	UpgradeCode     string `json:"upgrade_code"`
	Status          string `json:"status"`
	Message         string `json:"message"`
	RestartRequired bool   `json:"restart_required"`
}

// CreateExclusionWordRequest represents the request to create an exclusion word
type CreateExclusionWordRequest struct {
	Word string `json:"word" binding:"required"`
}

// ExclusionWordResponse represents an exclusion word in API responses
type ExclusionWordResponse struct {
	ID        string    `json:"id"`
	Word      string    `json:"word"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SuggestedBrandWordResponse represents a detected brand word suggestion
type SuggestedBrandWordResponse struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}
