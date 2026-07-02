package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/AI2HU/gego/internal/config"
	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
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

	fmt.Println("\n📊 Database Configuration")
	fmt.Println("--------------------------")
	fmt.Println("Gego uses a hybrid approach:")
	fmt.Println("  • PostgreSQL for LLMs, schedules, and users")
	fmt.Println("  • MongoDB for prompts and responses")
	fmt.Println()

	fmt.Println("🐘 PostgreSQL Configuration")
	postgresURI, err := promptOptional(reader, "PostgreSQL URI [postgres://localhost:5432/gego?sslmode=disable]: ", "postgres://localhost:5432/gego?sslmode=disable")
	if err != nil {
		return err
	}
	cfg.SQLDatabase.Provider = "postgres"
	cfg.SQLDatabase.URI = postgresURI
	cfg.SQLDatabase.Database = "gego"

	fmt.Println("\n🍃 MongoDB Configuration (for Prompts and Responses)")
	mongoURI, err2 := promptOptional(reader, "MongoDB URI [mongodb://localhost:27017]: ", "mongodb://localhost:27017")
	if err2 != nil {
		return err2
	}
	cfg.NoSQLDatabase.Provider = "mongodb"
	cfg.NoSQLDatabase.URI = mongoURI
	cfg.NoSQLDatabase.Database = "gego"

	fmt.Println("\n🔌 Testing database connections...")
	sqlConfig := &models.Config{
		Provider: cfg.SQLDatabase.Provider,
		URI:      cfg.SQLDatabase.URI,
		Database: cfg.SQLDatabase.Database,
	}

	nosqlConfig := &models.Config{
		Provider: cfg.NoSQLDatabase.Provider,
		URI:      cfg.NoSQLDatabase.URI,
		Database: cfg.NoSQLDatabase.Database,
	}

	testDB, dbErr := db.New(sqlConfig, nosqlConfig)
	if dbErr != nil {
		return fmt.Errorf("failed to create hybrid database: %w", dbErr)
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

	fmt.Println("\n🔄 Running database migrations...")
	if err := runDatabaseMigrations(ctx, testDB); err != nil {
		fmt.Printf("❌ Failed to run migrations: %v\n", err)
		fmt.Println("You may need to run migrations manually later.")
	} else {
		fmt.Println("✅ Database migrations completed successfully!")
	}

	fmt.Println("\n💾 Saving configuration...")
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Configuration saved to: %s\n", configPath)

	fmt.Println("\n📋 Configuration Summary")
	fmt.Println("========================")
	fmt.Printf("SQL Database: %s (%s)\n", cfg.SQLDatabase.Provider, cfg.SQLDatabase.URI)
	fmt.Printf("NoSQL Database: %s (%s)\n", cfg.NoSQLDatabase.Provider, cfg.NoSQLDatabase.URI)
	fmt.Printf("Database Name: %s\n", cfg.NoSQLDatabase.Database)
	fmt.Println()
	fmt.Println("🎉 Setup complete! You can now use gego.")
	fmt.Println()
	fmt.Println("ℹ️  Gego uses a hybrid database approach:")
	fmt.Println("   • PostgreSQL stores LLM configurations, schedules, and users")
	fmt.Println("   • MongoDB stores prompts and responses for keyword analysis")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Add LLM providers: gego llm add")
	fmt.Println("  2. Create prompts: gego prompt add")
	fmt.Println("  3. Set up schedules: gego schedule add")
	fmt.Println("  4. Start a worker: gego worker start")
	fmt.Println("  5. Start the scheduler: gego scheduler start")
	fmt.Println()
	if _, err := os.Stat(filepath.Join("internal", "db", "migrations")); err == nil {
		fmt.Println("Legacy SQLite installs can migrate with: gego db upgrade-from-sqlite")
	}

	return nil
}
