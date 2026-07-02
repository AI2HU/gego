package db

import (
	"context"

	"github.com/AI2HU/gego/internal/models"
)

// SQLDatabase defines the interface for SQL database operations (LLMs and Schedules).
// Implement new methods in internal/db/postgres/ only; SQLite is legacy (see internal/db/README.md).
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

	// User operations
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	ListUsers(ctx context.Context) ([]*models.User, error)

	// Session operations
	CreateSession(ctx context.Context, session *models.UserSession) error
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.UserSession, error)
	RevokeSession(ctx context.Context, id string) error

	// Exclusion word operations
	CreateExclusionWord(ctx context.Context, word *models.ExclusionWord) error
	GetExclusionWord(ctx context.Context, id string) (*models.ExclusionWord, error)
	GetExclusionWordByWord(ctx context.Context, word string) (*models.ExclusionWord, error)
	ListExclusionWords(ctx context.Context) ([]*models.ExclusionWord, error)
	DeleteExclusionWord(ctx context.Context, id string) error
	CountExclusionWords(ctx context.Context) (int, error)
}
