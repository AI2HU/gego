package db

import (
	"context"
	"time"

	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

// NoSQLDatabase defines the interface for NoSQL database operations (Prompts and Responses)
type NoSQLDatabase interface {
	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error

	// Prompt operations
	CreatePrompt(ctx context.Context, prompt *models.Prompt) error
	GetPrompt(ctx context.Context, id string) (*models.Prompt, error)
	ListPrompts(ctx context.Context, enabled *bool) ([]*models.Prompt, error)
	UpdatePrompt(ctx context.Context, prompt *models.Prompt) error
	DeletePrompt(ctx context.Context, id string) error
	DeleteAllPrompts(ctx context.Context) (int, error)

	// Response operations
	CreateResponse(ctx context.Context, response *models.Response) error
	GetResponse(ctx context.Context, id string) (*models.Response, error)
	ListResponses(ctx context.Context, filter shared.ResponseFilter) ([]*models.Response, error)
	DeleteAllResponses(ctx context.Context) (int, error)

	// Keyword search (on-demand, searches through response_text)
	SearchKeyword(ctx context.Context, keyword string, startTime, endTime *time.Time) (*models.KeywordStats, error)
	GetTopKeywords(ctx context.Context, limit int, startTime, endTime *time.Time) ([]models.KeywordCount, error)

	// Statistics operations
	GetPromptStats(ctx context.Context, promptID string) (*models.PromptStats, error)
	GetLLMStats(ctx context.Context, llmID string) (*models.LLMStats, error)
}
