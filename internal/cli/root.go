package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/AI2HU/gego/internal/config"
	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/llm"
	"github.com/AI2HU/gego/internal/llm/anthropic"
	"github.com/AI2HU/gego/internal/llm/google"
	"github.com/AI2HU/gego/internal/llm/ollama"
	"github.com/AI2HU/gego/internal/llm/openai"
	"github.com/AI2HU/gego/internal/llm/perplexity"
	"github.com/AI2HU/gego/internal/logger"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/services"
	"github.com/AI2HU/gego/internal/shared"
)

var (
	cfgFile      string
	logLevel     string
	logFile      string
	cfg          *config.Config
	database     db.Database
	llmRegistry  *llm.Registry
	sched        *services.SchedulerService
	statsService *services.StatsService
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "gego",
	Short: "GEO tracker for LLM responses",
	Long: `Gego is a GEO tracker tool that schedules prompts across multiple LLMs
and analyzes brand mentions in their responses.

Track which brands appear most frequently, which prompts generate the most mentions,
and compare performance across different LLM providers.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeLogging(); err != nil {
			return fmt.Errorf("failed to initialize logging: %w", err)
		}

		if cmd.Name() == "init" || cmd.Name() == "api" {
			return nil
		}

		if cfgFile == "" {
			cfgFile = config.GetConfigPath()
		}

		if !config.Exists(cfgFile) {
			return fmt.Errorf("configuration file not found. Run 'gego init' to create one")
		}

		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg.KeywordsExclusionPath != "" {
			exclusionPath := cfg.KeywordsExclusionPath
			if !filepath.IsAbs(exclusionPath) {
				configDir := filepath.Dir(cfgFile)
				exclusionPath = filepath.Join(configDir, exclusionPath)
			}
			shared.SetExclusionFilePath(exclusionPath)
		}

		sqlConfig := &models.Config{
			Provider: cfg.SQLDatabase.Provider,
			URI:      cfg.SQLDatabase.URI,
			Database: cfg.SQLDatabase.Database,
			Options:  cfg.SQLDatabase.Options,
		}

		nosqlConfig := &models.Config{
			Provider: cfg.NoSQLDatabase.Provider,
			URI:      cfg.NoSQLDatabase.URI,
			Database: cfg.NoSQLDatabase.Database,
			Options:  cfg.NoSQLDatabase.Options,
		}

		database, err = db.New(sqlConfig, nosqlConfig)
		if err != nil {
			return fmt.Errorf("failed to create hybrid database: %w", err)
		}

		if err := database.Connect(context.Background()); err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}

		statsService = services.NewStatsService(database)

		llmRegistry = llm.NewRegistry()
		llmRegistry.Register(openai.New("", ""))
		llmRegistry.Register(anthropic.New("", ""))
		llmRegistry.Register(ollama.New(""))
		llmRegistry.Register(google.New("", ""))
		llmRegistry.Register(perplexity.New("", ""))

		sched = services.NewSchedulerService(database, llmRegistry)

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if database != nil {
			return database.Disconnect(context.Background())
		}
		return nil
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gego/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "INFO", "log level (DEBUG, INFO, WARNING, ERROR)")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "log file path (default: stdout)")

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(apiCmd)
	rootCmd.AddCommand(llmCmd)
	rootCmd.AddCommand(promptCmd)
	rootCmd.AddCommand(scheduleCmd)
	rootCmd.AddCommand(schedulerCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(runCmd)
}

// Helper function to initialize LLM providers from configs
func initializeLLMProviders(ctx context.Context) error {
	llms, err := database.ListLLMs(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list LLMs: %w", err)
	}

	for _, llmConfig := range llms {
		var provider llm.Provider

		switch llmConfig.Provider {
		case "openai":
			provider = openai.New(llmConfig.APIKey, llmConfig.BaseURL)
		case "anthropic":
			provider = anthropic.New(llmConfig.APIKey, llmConfig.BaseURL)
		case "ollama":
			provider = ollama.New(llmConfig.BaseURL)
		case "google":
			provider = google.New(llmConfig.APIKey, llmConfig.BaseURL)
		case "perplexity":
			provider = perplexity.New(llmConfig.APIKey, llmConfig.BaseURL)
		default:
			continue
		}

		llmRegistry.Register(provider)
	}

	return nil
}

// initializeLogging sets up the logging system based on command line flags
func initializeLogging() error {
	level := logger.ParseLogLevel(logLevel)

	var output io.Writer = os.Stdout
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file %s: %w", logFile, err)
		}
		output = file
	}

	logger.Init(level, output)

	logger.Info("Logging initialized - Level: %s", level.String())
	if logFile != "" {
		logger.Info("Logging to file: %s", logFile)
	} else {
		logger.Info("Logging to stdout")
	}

	return nil
}
