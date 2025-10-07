package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/AI2HU/gego/internal/config"
	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/db/mongodb"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize gego configuration",
	Long:  `Interactive wizard to set up gego configuration including database and brand list.`,
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🚀 Welcome to Gego - GEO Tracker Setup")
	fmt.Println("======================================")
	fmt.Println()

	// Check if config already exists
	configPath := config.GetConfigPath()
	if config.Exists(configPath) {
		fmt.Printf("Configuration file already exists at: %s\n", configPath)
		confirmed, err := promptYesNo(reader, "Do you want to overwrite it? (y/N): ")
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Setup cancelled.")
			return nil
		}
	}

	cfg := config.DefaultConfig()

	// Database configuration
	fmt.Println("\n📊 Database Configuration")
	fmt.Println("--------------------------")

	provider, err := promptOptional(reader, "Database provider (mongodb/cassandra) [mongodb]: ", "mongodb")
	if err != nil {
		return err
	}
	cfg.Database.Provider = provider

	uri, err2 := promptOptional(reader, "Database URI [mongodb://localhost:27017]: ", "mongodb://localhost:27017")
	if err2 != nil {
		return err2
	}
	cfg.Database.URI = uri

	dbName, err3 := promptOptional(reader, "Database name [gego]: ", "gego")
	if err3 != nil {
		return err3
	}
	cfg.Database.Database = dbName

	// Test database connection
	fmt.Println("\n🔌 Testing database connection...")
	dbConfig := &db.Config{
		Provider: cfg.Database.Provider,
		URI:      cfg.Database.URI,
		Database: cfg.Database.Database,
	}

	var testDB db.Database
	var dbErr error

	switch cfg.Database.Provider {
	case "mongodb":
		testDB, dbErr = mongodb.New(dbConfig)
	default:
		return fmt.Errorf("unsupported database provider: %s", cfg.Database.Provider)
	}

	if dbErr != nil {
		return fmt.Errorf("failed to create database: %w", dbErr)
	}

	ctx := context.Background()
	if err := testDB.Connect(ctx); err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		fmt.Println("\nPlease check your database configuration and try again.")
		return err
	}
	defer testDB.Disconnect(ctx)

	if err := testDB.Ping(ctx); err != nil {
		fmt.Printf("❌ Failed to ping database: %v\n", err)
		return err
	}

	fmt.Println("✅ Database connection successful!")

	// Save configuration
	fmt.Println("\n💾 Saving configuration...")
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Configuration saved to: %s\n", configPath)

	// Summary
	fmt.Println("\n📋 Configuration Summary")
	fmt.Println("========================")
	fmt.Printf("Database: %s\n", cfg.Database.Provider)
	fmt.Printf("URI: %s\n", cfg.Database.URI)
	fmt.Printf("Database Name: %s\n", cfg.Database.Database)
	fmt.Println()
	fmt.Println("🎉 Setup complete! You can now use gego.")
	fmt.Println()
	fmt.Println("ℹ️  Gego automatically extracts keywords from LLM responses.")
	fmt.Println("   No predefined keyword list needed!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Add LLM providers: gego llm add")
	fmt.Println("  2. Create prompts: gego prompt add")
	fmt.Println("  3. Set up schedules: gego schedule add")
	fmt.Println("  4. Start scheduler: gego run")

	return nil
}
