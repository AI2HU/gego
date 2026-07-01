package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/AI2HU/gego/internal/config"
	"github.com/AI2HU/gego/internal/db/upgrade"
)

const UpgradeSQLiteToPostgres = "sqlite_to_postgres"

type UpgradeRunOptions struct {
	SQLitePath    string
	PostgresURI   string
	ConfigPath    string
	MigrationsDir string
	DryRun        bool
	Force         bool
}

type UpgradeRunResult struct {
	UpgradeCode     string
	Status          string
	Message         string
	RestartRequired bool
}

type UpgradeHandler interface {
	Code() string
	Required(cfg *config.Config) (bool, error)
	Run(ctx context.Context, cfg *config.Config, opts UpgradeRunOptions) (*UpgradeRunResult, error)
}

type UpgradeService struct {
	cfg      *config.Config
	handlers map[string]UpgradeHandler
	mu       sync.Mutex
}

func NewUpgradeService(cfg *config.Config) *UpgradeService {
	s := &UpgradeService{
		cfg:      cfg,
		handlers: make(map[string]UpgradeHandler),
	}
	s.Register(&sqliteToPostgresHandler{})
	return s
}

func (s *UpgradeService) Register(handler UpgradeHandler) {
	s.handlers[handler.Code()] = handler
}

func (s *UpgradeService) ListRequired(ctx context.Context) ([]string, error) {
	_ = ctx
	var codes []string
	for _, handler := range s.handlers {
		required, err := handler.Required(s.cfg)
		if err != nil {
			return nil, err
		}
		if required {
			codes = append(codes, handler.Code())
		}
	}
	return codes, nil
}

func (s *UpgradeService) Run(ctx context.Context, code string, opts UpgradeRunOptions) (*UpgradeRunResult, error) {
	handler, ok := s.handlers[code]
	if !ok {
		return nil, fmt.Errorf("unknown upgrade code: %s", code)
	}

	if !s.mu.TryLock() {
		return nil, fmt.Errorf("upgrade already in progress")
	}
	defer s.mu.Unlock()

	return handler.Run(ctx, s.cfg, opts)
}

type sqliteToPostgresHandler struct{}

func (h *sqliteToPostgresHandler) Code() string {
	return UpgradeSQLiteToPostgres
}

func (h *sqliteToPostgresHandler) Required(cfg *config.Config) (bool, error) {
	return upgrade.SQLiteUpgradeRequired(cfg)
}

func (h *sqliteToPostgresHandler) Run(ctx context.Context, cfg *config.Config, opts UpgradeRunOptions) (*UpgradeRunResult, error) {
	result, err := upgrade.RunSQLiteToPostgres(ctx, cfg, upgrade.Options{
		SQLitePath:    opts.SQLitePath,
		PostgresURI:   opts.PostgresURI,
		ConfigPath:    opts.ConfigPath,
		MigrationsDir: opts.MigrationsDir,
		DryRun:        opts.DryRun,
		Force:         opts.Force,
	})
	if err != nil {
		return nil, err
	}

	return &UpgradeRunResult{
		UpgradeCode:     UpgradeSQLiteToPostgres,
		Status:          "completed",
		Message:         result.Message,
		RestartRequired: result.RestartRequired,
	}, nil
}
