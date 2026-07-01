package api

import (
	"context"
	"fmt"

	"github.com/AI2HU/gego/internal/config"
	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/services"
)

func (s *Server) reconnectSQLFromConfig(ctx context.Context) error {
	s.dbMu.Lock()
	defer s.dbMu.Unlock()

	cfg, err := config.Load(s.configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if cfg.SQLDatabase.Provider != "postgres" {
		return nil
	}

	hybrid, ok := s.db.(*db.HybridDB)
	if !ok {
		return fmt.Errorf("database is not a HybridDB instance")
	}

	sqlConfig := &models.Config{
		Provider: cfg.SQLDatabase.Provider,
		URI:      cfg.SQLDatabase.URI,
		Database: cfg.SQLDatabase.Database,
		Options:  cfg.SQLDatabase.Options,
	}

	if err := hybrid.ReconnectSQL(ctx, sqlConfig); err != nil {
		return fmt.Errorf("reconnect SQL database: %w", err)
	}

	if err := db.RunHybridMigrations(ctx, hybrid); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return s.rebuildServices()
}

func (s *Server) rebuildServices() error {
	s.llmService = services.NewLLMService(s.db)
	s.promptService = services.NewPromptManagementService(s.db)
	s.scheduleService = services.NewScheduleService(s.db)
	s.statsService = services.NewStatsService(s.db)
	s.searchService = services.NewSearchService(s.db)
	s.authService = services.NewAuthService(s.db, s.authConfig)

	if s.runtime != nil {
		_ = s.runtime.Close()
	}

	runtime, err := services.NewRuntime(s.db, s.llmRegistry)
	if err != nil {
		return fmt.Errorf("reinitialize runtime: %w", err)
	}

	s.runtime = runtime
	s.schedulerService = runtime.Scheduler
	s.enqueueService = runtime.Enqueue
	s.jobStore = runtime.Store
	return nil
}
