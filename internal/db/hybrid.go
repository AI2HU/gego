package db

import (
	"context"
	"fmt"
	"time"

	"github.com/AI2HU/gego/internal/db/mongodb"
	"github.com/AI2HU/gego/internal/db/sqlite"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

// HybridDB implements the Database interface using both SQLite and NoSQL
type HybridDB struct {
	sqlDB   SQLDatabase   // SQLite for LLMs and Schedules
	nosqlDB NoSQLDatabase // MongoDB for Prompts and Responses
}

// New creates a new hybrid database instance
func New(sqlConfig, nosqlConfig *models.Config) (*HybridDB, error) {
	var sqlDB SQLDatabase
	var nosqlDB NoSQLDatabase
	var err error

	switch sqlConfig.Provider {
	case "sqlite":
		sqlDB, err = sqlite.New(sqlConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create SQLite database: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported SQL database provider: %s", sqlConfig.Provider)
	}

	switch nosqlConfig.Provider {
	case "mongodb":
		nosqlDB, err = mongodb.New(nosqlConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create NoSQL database: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported NoSQL database provider: %s", nosqlConfig.Provider)
	}

	return &HybridDB{
		sqlDB:   sqlDB,
		nosqlDB: nosqlDB,
	}, nil
}

// Connection management
func (h *HybridDB) Connect(ctx context.Context) error {
	if err := h.sqlDB.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to SQL database: %w", err)
	}

	if err := h.nosqlDB.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to NoSQL database: %w", err)
	}

	return nil
}

func (h *HybridDB) Disconnect(ctx context.Context) error {
	var errs []error

	if err := h.sqlDB.Disconnect(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to disconnect from SQL database: %w", err))
	}

	if err := h.nosqlDB.Disconnect(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to disconnect from NoSQL database: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("disconnect errors: %v", errs)
	}

	return nil
}

func (h *HybridDB) Ping(ctx context.Context) error {
	if err := h.sqlDB.Ping(ctx); err != nil {
		return fmt.Errorf("SQL database ping failed: %w", err)
	}

	if err := h.nosqlDB.Ping(ctx); err != nil {
		return fmt.Errorf("NoSQL database ping failed: %w", err)
	}

	return nil
}

func (h *HybridDB) CreateLLM(ctx context.Context, llm *models.LLMConfig) error {
	return h.sqlDB.CreateLLM(ctx, llm)
}

func (h *HybridDB) GetLLM(ctx context.Context, id string) (*models.LLMConfig, error) {
	return h.sqlDB.GetLLM(ctx, id)
}

func (h *HybridDB) ListLLMs(ctx context.Context, enabled *bool) ([]*models.LLMConfig, error) {
	return h.sqlDB.ListLLMs(ctx, enabled)
}

func (h *HybridDB) UpdateLLM(ctx context.Context, llm *models.LLMConfig) error {
	return h.sqlDB.UpdateLLM(ctx, llm)
}

func (h *HybridDB) DeleteLLM(ctx context.Context, id string) error {
	return h.sqlDB.DeleteLLM(ctx, id)
}

func (h *HybridDB) DeleteAllLLMs(ctx context.Context) (int, error) {
	return h.sqlDB.DeleteAllLLMs(ctx)
}

func (h *HybridDB) CreateSchedule(ctx context.Context, schedule *models.Schedule) error {
	return h.sqlDB.CreateSchedule(ctx, schedule)
}

func (h *HybridDB) GetSchedule(ctx context.Context, id string) (*models.Schedule, error) {
	return h.sqlDB.GetSchedule(ctx, id)
}

func (h *HybridDB) ListSchedules(ctx context.Context, enabled *bool) ([]*models.Schedule, error) {
	return h.sqlDB.ListSchedules(ctx, enabled)
}

func (h *HybridDB) UpdateSchedule(ctx context.Context, schedule *models.Schedule) error {
	return h.sqlDB.UpdateSchedule(ctx, schedule)
}

func (h *HybridDB) DeleteSchedule(ctx context.Context, id string) error {
	return h.sqlDB.DeleteSchedule(ctx, id)
}

func (h *HybridDB) DeleteAllSchedules(ctx context.Context) (int, error) {
	return h.sqlDB.DeleteAllSchedules(ctx)
}

// Prompt operations - Use NoSQL
func (h *HybridDB) CreatePrompt(ctx context.Context, prompt *models.Prompt) error {
	return h.nosqlDB.CreatePrompt(ctx, prompt)
}

func (h *HybridDB) GetPrompt(ctx context.Context, id string) (*models.Prompt, error) {
	return h.nosqlDB.GetPrompt(ctx, id)
}

func (h *HybridDB) ListPrompts(ctx context.Context, enabled *bool) ([]*models.Prompt, error) {
	return h.nosqlDB.ListPrompts(ctx, enabled)
}

func (h *HybridDB) UpdatePrompt(ctx context.Context, prompt *models.Prompt) error {
	return h.nosqlDB.UpdatePrompt(ctx, prompt)
}

func (h *HybridDB) DeletePrompt(ctx context.Context, id string) error {
	return h.nosqlDB.DeletePrompt(ctx, id)
}

func (h *HybridDB) DeleteAllPrompts(ctx context.Context) (int, error) {
	return h.nosqlDB.DeleteAllPrompts(ctx)
}

func (h *HybridDB) CreateResponse(ctx context.Context, response *models.Response) error {
	return h.nosqlDB.CreateResponse(ctx, response)
}

func (h *HybridDB) GetResponse(ctx context.Context, id string) (*models.Response, error) {
	return h.nosqlDB.GetResponse(ctx, id)
}

func (h *HybridDB) ListResponses(ctx context.Context, filter shared.ResponseFilter) ([]*models.Response, error) {
	return h.nosqlDB.ListResponses(ctx, filter)
}

func (h *HybridDB) DeleteAllResponses(ctx context.Context) (int, error) {
	return h.nosqlDB.DeleteAllResponses(ctx)
}

func (h *HybridDB) SearchKeyword(ctx context.Context, keyword string, startTime, endTime *time.Time) (*models.KeywordStats, error) {
	return h.nosqlDB.SearchKeyword(ctx, keyword, startTime, endTime)
}

func (h *HybridDB) GetTopKeywords(ctx context.Context, limit int, startTime, endTime *time.Time) ([]models.KeywordCount, error) {
	return h.nosqlDB.GetTopKeywords(ctx, limit, startTime, endTime)
}

func (h *HybridDB) GetPromptStats(ctx context.Context, promptID string) (*models.PromptStats, error) {
	return h.nosqlDB.GetPromptStats(ctx, promptID)
}

func (h *HybridDB) GetLLMStats(ctx context.Context, llmID string) (*models.LLMStats, error) {
	return h.nosqlDB.GetLLMStats(ctx, llmID)
}

func (h *HybridDB) GetNoSQLDatabase() *mongodb.MongoDB {
	if mongoDB, ok := h.nosqlDB.(*mongodb.MongoDB); ok {
		return mongoDB
	}
	return nil
}

func (h *HybridDB) GetSQLiteDatabase() *sqlite.SQLite {
	if sqliteDB, ok := h.sqlDB.(*sqlite.SQLite); ok {
		return sqliteDB
	}
	return nil
}
