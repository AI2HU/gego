package db

import (
	"context"

	"github.com/AI2HU/gego/internal/models"
)

// SQLDatabase defines the interface for SQL database operations (LLMs and Schedules)
type SQLDatabase interface {
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

	// Schedule operations
	CreateSchedule(ctx context.Context, schedule *models.Schedule) error
	GetSchedule(ctx context.Context, id string) (*models.Schedule, error)
	ListSchedules(ctx context.Context, enabled *bool) ([]*models.Schedule, error)
	UpdateSchedule(ctx context.Context, schedule *models.Schedule) error
	DeleteSchedule(ctx context.Context, id string) error
	DeleteAllSchedules(ctx context.Context) (int, error)
}
