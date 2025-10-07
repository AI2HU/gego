package db

import (
	"context"
	"time"

	"github.com/AI2HU/gego/internal/models"
)

// Database defines the interface for database operations
type Database interface {
	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error

	// LLM operations
	CreateLLM(ctx context.Context, llm *models.LLMConfig) error
	GetLLM(ctx context.Context, id string) (*models.LLMConfig, error)
	ListLLMs(ctx context.Context, enabled *bool) ([]*models.LLMConfig, error)
	UpdateLLM(ctx context.Context, llm *models.LLMConfig) error
	DeleteLLM(ctx context.Context, id string) error
	DeleteAllLLMs(ctx context.Context) (int, error)

	// Prompt operations
	CreatePrompt(ctx context.Context, prompt *models.Prompt) error
	GetPrompt(ctx context.Context, id string) (*models.Prompt, error)
	ListPrompts(ctx context.Context, enabled *bool) ([]*models.Prompt, error)
	UpdatePrompt(ctx context.Context, prompt *models.Prompt) error
	DeletePrompt(ctx context.Context, id string) error
	DeleteAllPrompts(ctx context.Context) (int, error)

	// Schedule operations
	CreateSchedule(ctx context.Context, schedule *models.Schedule) error
	GetSchedule(ctx context.Context, id string) (*models.Schedule, error)
	ListSchedules(ctx context.Context, enabled *bool) ([]*models.Schedule, error)
	UpdateSchedule(ctx context.Context, schedule *models.Schedule) error
	DeleteSchedule(ctx context.Context, id string) error
	DeleteAllSchedules(ctx context.Context) (int, error)

	// Response operations
	CreateResponse(ctx context.Context, response *models.Response) error
	GetResponse(ctx context.Context, id string) (*models.Response, error)
	ListResponses(ctx context.Context, filter ResponseFilter) ([]*models.Response, error)

	// Keyword search (on-demand, searches through response_text)
	SearchKeyword(ctx context.Context, keyword string, startTime, endTime *time.Time) (*models.KeywordStats, error)
	GetTopKeywords(ctx context.Context, limit int, startTime, endTime *time.Time) ([]KeywordCount, error)
}

// ResponseFilter provides filtering options for listing responses
type ResponseFilter struct {
	PromptID   string
	LLMID      string
	ScheduleID string
	Keyword    string
	StartTime  *time.Time
	EndTime    *time.Time
	Limit      int
	Offset     int
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

// Config holds database configuration
type Config struct {
	Provider string            // mongodb, cassandra
	URI      string            // Connection URI
	Database string            // Database name
	Options  map[string]string // Provider-specific options
}
