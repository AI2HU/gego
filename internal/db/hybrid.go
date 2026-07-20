package db

import (
	"context"
	"fmt"
	"time"

	"github.com/AI2HU/gego/internal/db/mongodb"
	"github.com/AI2HU/gego/internal/db/postgres"
	"github.com/AI2HU/gego/internal/db/sqlite"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

// HybridDB implements the Database interface using PostgreSQL (or legacy SQLite) and MongoDB.
type HybridDB struct {
	sqlDB   SQLBackend
	nosqlDB NoSQLDatabase
}

// New creates a new hybrid database instance.
func New(sqlConfig, nosqlConfig *models.Config) (*HybridDB, error) {
	var sqlDB SQLBackend
	var nosqlDB NoSQLDatabase
	var err error

	switch sqlConfig.Provider {
	case "postgres":
		sqlDB, err = postgres.New(sqlConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create PostgreSQL database: %w", err)
		}
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

func (h *HybridDB) ReconnectSQL(ctx context.Context, sqlConfig *models.Config) error {
	if h.sqlDB != nil {
		_ = h.sqlDB.Disconnect(ctx)
	}

	var sqlDB SQLBackend
	var err error

	switch sqlConfig.Provider {
	case "postgres":
		sqlDB, err = postgres.New(sqlConfig)
		if err != nil {
			return fmt.Errorf("failed to create PostgreSQL database: %w", err)
		}
	case "sqlite":
		sqlDB, err = sqlite.New(sqlConfig)
		if err != nil {
			return fmt.Errorf("failed to create SQLite database: %w", err)
		}
	default:
		return fmt.Errorf("unsupported SQL database provider: %s", sqlConfig.Provider)
	}

	if err := sqlDB.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to SQL database: %w", err)
	}

	h.sqlDB = sqlDB
	return nil
}

func (h *HybridDB) GetSQLBackend() SQLBackend {
	return h.sqlDB
}

func (h *HybridDB) CleanAll(ctx context.Context) error {
	return h.nosqlDB.CleanAll(ctx)
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

func (h *HybridDB) CreateUser(ctx context.Context, user *models.User) error {
	return h.sqlDB.CreateUser(ctx, user)
}

func (h *HybridDB) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return h.sqlDB.GetUserByUsername(ctx, username)
}

func (h *HybridDB) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return h.sqlDB.GetUserByID(ctx, id)
}

func (h *HybridDB) ListUsers(ctx context.Context) ([]*models.User, error) {
	return h.sqlDB.ListUsers(ctx)
}

func (h *HybridDB) UpdateUser(ctx context.Context, user *models.User) error {
	return h.sqlDB.UpdateUser(ctx, user)
}

func (h *HybridDB) DeleteUser(ctx context.Context, id string) error {
	return h.sqlDB.DeleteUser(ctx, id)
}

func (h *HybridDB) CountUsersByRole(ctx context.Context, role models.Role) (int, error) {
	return h.sqlDB.CountUsersByRole(ctx, role)
}

func (h *HybridDB) CreateSession(ctx context.Context, session *models.UserSession) error {
	return h.sqlDB.CreateSession(ctx, session)
}

func (h *HybridDB) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.UserSession, error) {
	return h.sqlDB.GetSessionByTokenHash(ctx, tokenHash)
}

func (h *HybridDB) RevokeSession(ctx context.Context, id string) error {
	return h.sqlDB.RevokeSession(ctx, id)
}

func (h *HybridDB) RevokeSessionsByUserID(ctx context.Context, userID string) error {
	return h.sqlDB.RevokeSessionsByUserID(ctx, userID)
}

func (h *HybridDB) CreateExclusionWord(ctx context.Context, word *models.ExclusionWord) error {
	ew, err := exclusionWordBackend(h.sqlDB)
	if err != nil {
		return err
	}
	return ew.CreateExclusionWord(ctx, word)
}

func (h *HybridDB) GetExclusionWord(ctx context.Context, id string) (*models.ExclusionWord, error) {
	ew, err := exclusionWordBackend(h.sqlDB)
	if err != nil {
		return nil, err
	}
	return ew.GetExclusionWord(ctx, id)
}

func (h *HybridDB) GetExclusionWordByWord(ctx context.Context, word string) (*models.ExclusionWord, error) {
	ew, err := exclusionWordBackend(h.sqlDB)
	if err != nil {
		return nil, err
	}
	return ew.GetExclusionWordByWord(ctx, word)
}

func (h *HybridDB) ListExclusionWords(ctx context.Context) ([]*models.ExclusionWord, error) {
	ew, err := exclusionWordBackend(h.sqlDB)
	if err != nil {
		return nil, err
	}
	return ew.ListExclusionWords(ctx)
}

func (h *HybridDB) DeleteExclusionWord(ctx context.Context, id string) error {
	ew, err := exclusionWordBackend(h.sqlDB)
	if err != nil {
		return err
	}
	return ew.DeleteExclusionWord(ctx, id)
}

func (h *HybridDB) CountExclusionWords(ctx context.Context) (int, error) {
	ew, err := exclusionWordBackend(h.sqlDB)
	if err != nil {
		return 0, err
	}
	return ew.CountExclusionWords(ctx)
}

func (h *HybridDB) DeleteAllExclusionWords(ctx context.Context) (int, error) {
	ew, err := exclusionWordBackend(h.sqlDB)
	if err != nil {
		return 0, err
	}
	return ew.DeleteAllExclusionWords(ctx)
}

func (h *HybridDB) CreateBrand(ctx context.Context, brand *models.Brand) error {
	b, err := brandBackend(h.sqlDB)
	if err != nil {
		return err
	}
	return b.CreateBrand(ctx, brand)
}

func (h *HybridDB) GetBrand(ctx context.Context, id string) (*models.Brand, error) {
	b, err := brandBackend(h.sqlDB)
	if err != nil {
		return nil, err
	}
	return b.GetBrand(ctx, id)
}

func (h *HybridDB) GetBrandByName(ctx context.Context, name string) (*models.Brand, error) {
	b, err := brandBackend(h.sqlDB)
	if err != nil {
		return nil, err
	}
	return b.GetBrandByName(ctx, name)
}

func (h *HybridDB) ListBrands(ctx context.Context) ([]*models.Brand, error) {
	b, err := brandBackend(h.sqlDB)
	if err != nil {
		return nil, err
	}
	return b.ListBrands(ctx)
}

func (h *HybridDB) UpdateBrand(ctx context.Context, brand *models.Brand) error {
	b, err := brandBackend(h.sqlDB)
	if err != nil {
		return err
	}
	return b.UpdateBrand(ctx, brand)
}

func (h *HybridDB) DeleteBrand(ctx context.Context, id string) error {
	b, err := brandBackend(h.sqlDB)
	if err != nil {
		return err
	}
	return b.DeleteBrand(ctx, id)
}

func (h *HybridDB) CreateBrandAlias(ctx context.Context, alias *models.BrandAlias) error {
	b, err := brandBackend(h.sqlDB)
	if err != nil {
		return err
	}
	return b.CreateBrandAlias(ctx, alias)
}

func (h *HybridDB) GetBrandAlias(ctx context.Context, id string) (*models.BrandAlias, error) {
	b, err := brandBackend(h.sqlDB)
	if err != nil {
		return nil, err
	}
	return b.GetBrandAlias(ctx, id)
}

func (h *HybridDB) GetBrandAliasByAlias(ctx context.Context, alias string) (*models.BrandAlias, error) {
	b, err := brandBackend(h.sqlDB)
	if err != nil {
		return nil, err
	}
	return b.GetBrandAliasByAlias(ctx, alias)
}

func (h *HybridDB) ListBrandAliasesByBrandID(ctx context.Context, brandID string) ([]*models.BrandAlias, error) {
	b, err := brandBackend(h.sqlDB)
	if err != nil {
		return nil, err
	}
	return b.ListBrandAliasesByBrandID(ctx, brandID)
}

func (h *HybridDB) UpdateBrandAlias(ctx context.Context, alias *models.BrandAlias) error {
	b, err := brandBackend(h.sqlDB)
	if err != nil {
		return err
	}
	return b.UpdateBrandAlias(ctx, alias)
}

func (h *HybridDB) DeleteBrandAlias(ctx context.Context, id string) error {
	b, err := brandBackend(h.sqlDB)
	if err != nil {
		return err
	}
	return b.DeleteBrandAlias(ctx, id)
}

func (h *HybridDB) DeleteAllBrands(ctx context.Context) (int, error) {
	b, err := brandBackend(h.sqlDB)
	if err != nil {
		return 0, err
	}
	return b.DeleteAllBrands(ctx)
}

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

func (h *HybridDB) CountResponses(ctx context.Context, filter shared.ResponseFilter) (int64, error) {
	return h.nosqlDB.CountResponses(ctx, filter)
}

func (h *HybridDB) DeleteAllResponses(ctx context.Context) (int, error) {
	return h.nosqlDB.DeleteAllResponses(ctx)
}

func (h *HybridDB) SearchKeyword(ctx context.Context, keyword string, startTime, endTime *time.Time, promptIDs []string) (*models.KeywordStats, error) {
	return h.nosqlDB.SearchKeyword(ctx, keyword, startTime, endTime, promptIDs)
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
