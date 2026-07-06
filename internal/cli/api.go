package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/AI2HU/gego/internal/api"
	"github.com/AI2HU/gego/internal/auth"
	appconfig "github.com/AI2HU/gego/internal/config"
	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/logger"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/services"
)

var (
	apiPort    string
	apiHost    string
	corsOrigin string
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the Gego REST API server",
	Long: `Start the Gego REST API server with full CRUD operations for:
- LLMs (Create, Read, Update, Delete)
- Prompts (Create, Read, Update, Delete)  
- Schedules (Create, Read, Update, Delete)
- Stats (Read-only)
- Search (POST endpoint for keyword search)

Authentication uses JWT Bearer tokens. Set GEGO_JWT_SECRET (min 32 characters)
before starting the server.

On first start (no users in database), create the initial admin via:
  GEGO_BOOTSTRAP_ADMIN_PASSWORD (required)
  GEGO_BOOTSTRAP_ADMIN_USERNAME (optional, defaults to admin)

Or create users manually: gego user create --username <name> --role admin`,
	RunE: runAPI,
}

func init() {
	apiCmd.Flags().StringVarP(&apiPort, "port", "p", "8989", "Port to run the API server on")
	apiCmd.Flags().StringVarP(&apiHost, "host", "H", "0.0.0.0", "Host to bind the API server to")
	apiCmd.Flags().StringVarP(&corsOrigin, "cors-origin", "c", "", "CORS origin to allow (overrides config file, use '*' for all origins)")
}

func runAPI(cmd *cobra.Command, args []string) error {
	var configPath string
	if cfgFile != "" {
		configPath = cfgFile
	} else if envPath := os.Getenv("GEGO_CONFIG_PATH"); envPath != "" {
		configPath = envPath
	} else {
		configPath = appconfig.GetConfigPath()
	}

	var cfg *appconfig.Config
	var err error
	if appconfig.Exists(configPath) {
		cfg, err = appconfig.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		cfg = appconfig.LoadFromEnv()
		fmt.Printf("No config file at %s — using environment variables\n", configPath)
	}

	selectedCORSOrigin := corsOrigin
	if selectedCORSOrigin == "" {
		if cfg.CORSOrigin != "" {
			selectedCORSOrigin = cfg.CORSOrigin
		} else {
			selectedCORSOrigin = "*"
		}
	}

	fmt.Printf("🚀 Starting Gego API Server\n")
	fmt.Printf("===========================\n")
	fmt.Printf("Host: %s\n", apiHost)
	fmt.Printf("Port: %s\n", apiPort)
	fmt.Printf("CORS Origin: %s\n", selectedCORSOrigin)
	fmt.Printf("URL: http://%s:%s/api/v1\n", apiHost, apiPort)
	fmt.Println()

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
		return fmt.Errorf("failed to create hybrid database: %w", err)
	}

	ctx := context.Background()
	if err := database.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Disconnect(ctx)

	if err := database.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	fmt.Println("✅ Database connection successful!")

	if cfg.SQLDatabase.Provider == "sqlite" {
		logger.Info("SQLite legacy mode — run database upgrade via UI or `gego db upgrade-from-sqlite`")
		fmt.Println("\n⚠️  SQLite legacy mode detected. Upgrade to PostgreSQL via the admin UI or CLI.")
	}

	fmt.Println("\n🔄 Running database migrations...")
	if err := runDatabaseMigrations(ctx, database); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	fmt.Println("✅ Database migrations completed successfully!")

	exclusionWordsService := services.NewExclusionWordsService(database)
	if err := exclusionWordsService.Initialize(ctx, cfg.ResolveKeywordsExclusionPath(configPath)); err != nil {
		return fmt.Errorf("failed to initialize exclusion words: %w", err)
	}

	brandsService := services.NewBrandsService(database)
	if err := brandsService.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize brands: %w", err)
	}

	fmt.Println("\n🔐 Checking API users...")
	bootstrappedUser, err := services.BootstrapAdminFromEnv(ctx, database)
	if err != nil {
		return err
	}
	if bootstrappedUser != nil {
		fmt.Printf("✅ Bootstrapped admin user %q from environment\n", bootstrappedUser.Username)
	} else {
		fmt.Println("✅ API users ready")
	}

	authConfig, err := auth.NewConfig(cfg)
	if err != nil {
		return err
	}

	server, err := api.NewServer(database, selectedCORSOrigin, authConfig, cfg, configPath, exclusionWordsService, brandsService)
	if err != nil {
		return fmt.Errorf("failed to create API server: %w", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n🛑 Shutting down API server...")
		_ = server.Close()
		_ = database.Disconnect(ctx)
		os.Exit(0)
	}()

	fmt.Println("🌐 API Server is running!")
	fmt.Println()
	fmt.Println("📚 Available Endpoints:")
	fmt.Println("  Models:")
	fmt.Println("    GET    /api/v1/models              - List all models")
	fmt.Println("    GET    /api/v1/models/:id          - Get specific model")
	fmt.Println("    POST   /api/v1/models              - Create new model")
	fmt.Println("    PUT    /api/v1/models/:id          - Update model")
	fmt.Println("    DELETE /api/v1/models/:id          - Delete model")
	fmt.Println()
	fmt.Println("  Providers:")
	fmt.Println("    GET    /api/v1/providers                      - List providers")
	fmt.Println("    GET    /api/v1/providers/:provider/api-keys     - List masked API keys")
	fmt.Println("    POST   /api/v1/providers/:provider/models     - Discover provider models")
	fmt.Println()
	fmt.Println("  Prompts:")
	fmt.Println("    GET    /api/v1/prompts           - List all prompts")
	fmt.Println("    GET    /api/v1/prompts/:id       - Get specific prompt")
	fmt.Println("    POST   /api/v1/prompts           - Create new prompt")
	fmt.Println("    PUT    /api/v1/prompts/:id       - Update prompt")
	fmt.Println("    DELETE /api/v1/prompts/:id       - Delete prompt")
	fmt.Println()
	fmt.Println("  Schedules:")
	fmt.Println("    GET    /api/v1/schedules         - List all schedules")
	fmt.Println("    GET    /api/v1/schedules/:id     - Get specific schedule")
	fmt.Println("    POST   /api/v1/schedules         - Create new schedule")
	fmt.Println("    PUT    /api/v1/schedules/:id     - Update schedule")
	fmt.Println("    DELETE /api/v1/schedules/:id     - Delete schedule")
	fmt.Println()
	fmt.Println("  Stats & Search:")
	fmt.Println("    GET    /api/v1/stats             - Get statistics")
	fmt.Println("    POST   /api/v1/search            - Search keywords")
	fmt.Println()
	fmt.Println("  Auth:")
	fmt.Println("    POST   /api/v1/auth/login        - Login (public)")
	fmt.Println("    POST   /api/v1/auth/refresh      - Refresh access token (public, cookie)")
	fmt.Println("    POST   /api/v1/auth/logout         - Logout (public, cookie)")
	fmt.Println("    GET    /api/v1/auth/me           - Current user profile")
	fmt.Println("    GET    /api/v1/health            - Health check (public)")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the server")

	address := fmt.Sprintf("%s:%s", apiHost, apiPort)
	return server.Run(address)
}
