package models

import (
	"time"
)

// Config holds database configuration
type Config struct {
	Provider string            // sqlite, mongodb, cassandra
	URI      string            // Connection URI
	Database string            // Database name
	Options  map[string]string // Provider-specific options
}

// TimeSeriesPoint represents a point in time-series data
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

// KeywordCount represents a keyword and its mention count
type KeywordCount struct {
	Keyword string `json:"keyword"`
	Count   int    `json:"count"`
}

// LLMConfig represents an LLM provider configuration
type LLMConfig struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Provider  string            `json:"provider"` // openai, anthropic, ollama, custom
	Model     string            `json:"model"`
	APIKey    string            `json:"api_key,omitempty"`
	BaseURL   string            `json:"base_url,omitempty"`
	Config    map[string]string `json:"config,omitempty"` // Additional provider-specific config
	Enabled   bool              `json:"enabled"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// Prompt represents a prompt template
type Prompt struct {
	ID        string    `json:"id"`
	Template  string    `json:"template"`
	Tags      []string  `json:"tags,omitempty"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Schedule represents a scheduler configuration
type Schedule struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	PromptIDs   []string   `json:"prompt_ids"`
	LLMIDs      []string   `json:"llm_ids"`
	CronExpr    string     `json:"cron_expr"`             // Cron expression for scheduling
	Temperature float64    `json:"temperature,omitempty"` // Temperature for LLM generation (0-1, default 0.7)
	Enabled     bool       `json:"enabled"`
	LastRun     *time.Time `json:"last_run,omitempty"`
	NextRun     *time.Time `json:"next_run,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Response represents an LLM response to a prompt
type Response struct {
	ID           string                 `json:"id" bson:"_id"`
	PromptID     string                 `json:"prompt_id" bson:"prompt_id"`
	PromptText   string                 `json:"prompt_text" bson:"prompt_text"` // Actual prompt sent
	LLMID        string                 `json:"llm_id" bson:"llm_id"`
	LLMName      string                 `json:"llm_name" bson:"llm_name"`
	LLMProvider  string                 `json:"llm_provider" bson:"llm_provider"`
	LLMModel     string                 `json:"llm_model" bson:"llm_model"`
	ResponseText string                 `json:"response_text" bson:"response_text"`
	Temperature  float64                `json:"temperature,omitempty" bson:"temperature,omitempty"` // Temperature used for generation
	Metadata     map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`       // Additional metadata
	ScheduleID   string                 `json:"schedule_id,omitempty" bson:"schedule_id,omitempty"`
	TokensUsed   int                    `json:"tokens_used,omitempty" bson:"tokens_used,omitempty"`
	LatencyMs    int64                  `json:"latency_ms,omitempty" bson:"latency_ms,omitempty"`
	Error        string                 `json:"error,omitempty" bson:"error,omitempty"`
	CreatedAt    time.Time              `json:"created_at" bson:"created_at"`
}

// PromptStats represents aggregated statistics for a prompt
type PromptStats struct {
	PromptID       string         `json:"prompt_id"`
	TotalResponses int            `json:"total_responses"`
	UniqueLLMs     int            `json:"unique_llms"`
	LLMCounts      map[string]int `json:"llm_counts"`
	AvgTokens      float64        `json:"avg_tokens"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// LLMStats represents aggregated statistics for an LLM
type LLMStats struct {
	LLMID          string         `json:"llm_id"`
	TotalResponses int            `json:"total_responses"`
	UniquePrompts  int            `json:"unique_prompts"`
	PromptCounts   map[string]int `json:"prompt_counts"`
	AvgTokens      float64        `json:"avg_tokens"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// KeywordStats represents on-demand calculated statistics for a keyword search
type KeywordStats struct {
	Keyword       string         `json:"keyword"`
	TotalMentions int            `json:"total_mentions"`
	UniquePrompts int            `json:"unique_prompts"`
	UniqueLLMs    int            `json:"unique_llms"`
	ByPrompt      map[string]int `json:"by_prompt"`   // prompt_id -> count
	ByLLM         map[string]int `json:"by_llm"`      // llm_id -> count
	ByProvider    map[string]int `json:"by_provider"` // provider -> count
	FirstSeen     time.Time      `json:"first_seen"`
	LastSeen      time.Time      `json:"last_seen"`
}

// ModelInfo represents information about an available model from a provider
type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
