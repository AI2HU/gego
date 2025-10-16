package models

import (
	"time"
)

// API Response structures

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

// LLM API Request/Response structures

// CreateLLMRequest represents the request to create a new LLM
type CreateLLMRequest struct {
	Name     string            `json:"name" binding:"required"`
	Provider string            `json:"provider" binding:"required"`
	Model    string            `json:"model" binding:"required"`
	APIKey   string            `json:"api_key,omitempty"`
	BaseURL  string            `json:"base_url,omitempty"`
	Config   map[string]string `json:"config,omitempty"`
	Enabled  bool              `json:"enabled"`
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

// Prompt API Request/Response structures

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

// Schedule API Request/Response structures

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
	NextRun     *time.Time `json:"next_run,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Stats API Response structures

// StatsResponse represents the response for statistics
type StatsResponse struct {
	TotalResponses int64             `json:"total_responses"`
	TotalPrompts   int64             `json:"total_prompts"`
	TotalLLMs      int64             `json:"total_llms"`
	TotalSchedules int64             `json:"total_schedules"`
	TopKeywords    []KeywordCount    `json:"top_keywords"`
	PromptStats    []*PromptStats    `json:"prompt_stats"`
	LLMStats       []*LLMStats       `json:"llm_stats"`
	ResponseTrends []TimeSeriesPoint `json:"response_trends"`
	LastUpdated    time.Time         `json:"last_updated"`
}

// Search API Request/Response structures

// SearchRequest represents the request to search responses
type SearchRequest struct {
	Keyword   string     `json:"keyword" binding:"required"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Limit     int        `json:"limit,omitempty"`
}

// SearchResponse represents the response for search operations
type SearchResponse struct {
	Keyword       string         `json:"keyword"`
	TotalMentions int            `json:"total_mentions"`
	UniquePrompts int            `json:"unique_prompts"`
	UniqueLLMs    int            `json:"unique_llms"`
	ByPrompt      map[string]int `json:"by_prompt"`
	ByLLM         map[string]int `json:"by_llm"`
	ByProvider    map[string]int `json:"by_provider"`
	FirstSeen     time.Time      `json:"first_seen"`
	LastSeen      time.Time      `json:"last_seen"`
	Responses     []*Response    `json:"responses,omitempty"`
}
