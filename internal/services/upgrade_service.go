package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/AI2HU/gego/internal/config"
	"github.com/AI2HU/gego/internal/db/upgrade"
	"github.com/AI2HU/gego/internal/models"
)

const UpgradeSQLiteToPostgres = "sqlite_to_postgres"

type UpgradeSeverity string

const (
	UpgradeSeverityMajor UpgradeSeverity = "major"
	UpgradeSeverityMinor UpgradeSeverity = "minor"
)

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
	Severity() UpgradeSeverity
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

func (s *UpgradeService) ReloadConfig(path string) error {
	if path == "" || !config.Exists(path) {
		return nil
	}
	cfg, err := config.Load(path)
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}
	s.cfg = cfg
	return nil
}

func (s *UpgradeService) ListUpgrades(ctx context.Context) ([]models.UpgradeItem, error) {
	_ = ctx
	items := []models.UpgradeItem{}
	for _, handler := range s.handlers {
		required, err := handler.Required(s.cfg)
		if err != nil {
			return nil, err
		}
		if !required {
			continue
		}
		items = append(items, models.UpgradeItem{
			Code:   handler.Code(),
			Severity: string(handler.Severity()),
		})
	}
	return items, nil
}

func (s *UpgradeService) ListRequired(ctx context.Context) ([]string, error) {
	items, err := s.ListUpgrades(ctx)
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(items))
	for _, item := range items {
		codes = append(codes, item.Code)
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

	result, err := handler.Run(ctx, s.cfg, opts)
	if err != nil {
		return nil, err
	}

	if opts.ConfigPath != "" {
		_ = s.ReloadConfig(opts.ConfigPath)
	}

	return result, nil
}

type sqliteToPostgresHandler struct{}

func (h *sqliteToPostgresHandler) Code() string {
	return UpgradeSQLiteToPostgres
}

func (h *sqliteToPostgresHandler) Severity() UpgradeSeverity {
	return UpgradeSeverityMajor
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
