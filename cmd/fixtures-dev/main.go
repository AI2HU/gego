package main

import (
	"context"
	"fmt"
	"os"

	appconfig "github.com/AI2HU/gego/internal/config"
	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/fixtures"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/services"
)

func main() {
	if os.Getenv(fixtures.EnvFixtures) != "dev" {
		fmt.Fprintf(os.Stderr, "refusing to load fixtures: set %s=dev\n", fixtures.EnvFixtures)
		os.Exit(1)
	}

	configPath := os.Getenv("GEGO_CONFIG_PATH")
	cfg, err := appconfig.ResolveConfig(configPath, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
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

	database, err := db.New(sqlConfig, nosqlConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create database: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := database.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer database.Disconnect(ctx)

	if err := database.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "database ping failed: %v\n", err)
		os.Exit(1)
	}

	if err := db.RunHybridMigrations(ctx, database); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	loader := &fixtures.Loader{Set: "dev"}

	fmt.Println("Cleaning gego SQL and MongoDB databases...")
	if err := loader.Reset(ctx, database); err != nil {
		fmt.Fprintf(os.Stderr, "failed to reset fixtures: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Loading dev fixtures...")
	summary, err := loader.Load(ctx, database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load fixtures: %v\n", err)
		os.Exit(1)
	}

	exclusionWordsService := services.NewExclusionWordsService(database)
	if err := exclusionWordsService.Initialize(ctx, cfg.ResolveKeywordsExclusionPath(configPath)); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize exclusion words cache: %v\n", err)
		os.Exit(1)
	}

	brandsService := services.NewBrandsService(database)
	if err := brandsService.Initialize(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize brands cache: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf(
		"Fixtures loaded: %d llms, %d prompts, %d schedules, %d brands, %d exclusion words, %d responses\n",
		summary.LLMs,
		summary.Prompts,
		summary.Schedules,
		summary.Brands,
		summary.ExclusionWords,
		summary.Responses,
	)
}
