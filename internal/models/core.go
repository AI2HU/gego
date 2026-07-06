package models

import (
	"time"
)

// Core domain models

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

// SearchURL represents a URL from web search with metadata
type SearchURL struct {
	SearchQuery   string `json:"search_query" bson:"search_query"`
	URL           string `json:"url" bson:"url"`
	Title         string `json:"title" bson:"title"`
	CitationIndex int    `json:"citation_index" bson:"citation_index"`
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
	RunID        string                 `json:"run_id,omitempty" bson:"run_id,omitempty"`
	JobID        string                 `json:"job_id,omitempty" bson:"job_id,omitempty"`
	TokensUsed   int                    `json:"tokens_used,omitempty" bson:"tokens_used,omitempty"`
	Error        string                 `json:"error,omitempty" bson:"error,omitempty"`
	SearchURLs   []SearchURL            `json:"search_urls,omitempty" bson:"search_urls,omitempty"`
	CreatedAt    time.Time              `json:"created_at" bson:"created_at"`
}

// ExclusionWord represents a word excluded from brand/keyword tracking
type ExclusionWord struct {
	ID        string    `json:"id"`
	Word      string    `json:"word"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Brand represents a canonical brand name with optional aliases
type Brand struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Aliases   []*BrandAlias `json:"aliases,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// BrandAlias maps a detected token to a canonical brand name
type BrandAlias struct {
	ID            string    `json:"id"`
	BrandID       string    `json:"brand_id"`
	Alias         string    `json:"alias"`
	CaseSensitive bool      `json:"case_sensitive"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ModelInfo represents information about an available model from a provider
type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	UsedInChat  bool   `json:"used_in_chat,omitempty"`
}
