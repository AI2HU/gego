package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
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
	fmt.Println("  • SQLite for LLMs and Schedules (structured data)")
	fmt.Println("  • MongoDB for Prompts and Responses (unstructured data)")
	fmt.Println()

	fmt.Println("🗄️  SQLite Configuration (for LLMs and Schedules)")
	sqlitePath, err := promptOptional(reader, "SQLite database path [gego.db]: ", "gego.db")
	if err != nil {
		return err
	}
	cfg.SQLDatabase.Provider = "sqlite"
	cfg.SQLDatabase.URI = sqlitePath
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
	if err := runMigrations(sqlitePath); err != nil {
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
	fmt.Printf("SQLite Database: %s (%s)\n", cfg.SQLDatabase.Provider, cfg.SQLDatabase.URI)
	fmt.Printf("NoSQL Database: %s (%s)\n", cfg.NoSQLDatabase.Provider, cfg.NoSQLDatabase.URI)
	fmt.Printf("Database Name: %s\n", cfg.NoSQLDatabase.Database)
	fmt.Println()
	fmt.Println("🎉 Setup complete! You can now use gego.")
	fmt.Println()
	fmt.Println("ℹ️  Gego uses a hybrid database approach:")
	fmt.Println("   • SQLite stores LLM configurations and schedules")
	fmt.Println("   • MongoDB stores prompts and responses for keyword analysis")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Add LLM providers: gego llm add")
	fmt.Println("  2. Create prompts: gego prompt add")
	fmt.Println("  3. Set up schedules: gego schedule add")
	fmt.Println("  4. Start scheduler: gego run")
	fmt.Println()
	fmt.Println("Migration commands:")
	fmt.Println("  • Run migrations: gego migrate up")
	fmt.Println("  • Check status: gego migrate status")

	return nil
}

// runMigrations executes database migrations using gomigrate
func runMigrations(sqlitePath string) error {
	if _, err := exec.LookPath("migrate"); err != nil {
		return fmt.Errorf("migrate command not found. Please install golang-migrate: https://github.com/golang-migrate/migrate")
	}

	migrationsDir := filepath.Join("internal", "db", "migrations")

	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory not found: %s", migrationsDir)
	}

	absSQLitePath, err := filepath.Abs(sqlitePath)
	if err != nil {
		return fmt.Errorf("failed to resolve SQLite path: %w", err)
	}

	dbURL := fmt.Sprintf("sqlite3://%s", absSQLitePath)

	cmd := exec.Command("migrate",
		"-path", migrationsDir,
		"-database", dbURL,
		"up")

	cmd.Dir = "."

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("migration failed: %w\nOutput: %s", err, string(output))
	}

	if len(output) > 0 {
		fmt.Printf("Migration output: %s", string(output))
	}

	return nil
}
