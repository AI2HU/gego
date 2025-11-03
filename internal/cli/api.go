package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/AI2HU/gego/internal/api"
	"github.com/AI2HU/gego/internal/config"
	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
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

The API runs on HTTP (no authentication required for now).`,
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
		configPath = config.GetConfigPath()
	}

	if !config.Exists(configPath) {
		return fmt.Errorf("configuration file not found at %s. Run 'gego init' to create one", configPath)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
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

	server := api.NewServer(database, selectedCORSOrigin)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n🛑 Shutting down API server...")
		database.Disconnect(ctx)
		os.Exit(0)
	}()

	fmt.Println("🌐 API Server is running!")
	fmt.Println()
	fmt.Println("📚 Available Endpoints:")
	fmt.Println("  LLMs:")
	fmt.Println("    GET    /api/v1/llms              - List all LLMs")
	fmt.Println("    GET    /api/v1/llms/:id          - Get specific LLM")
	fmt.Println("    POST   /api/v1/llms              - Create new LLM")
	fmt.Println("    PUT    /api/v1/llms/:id          - Update LLM")
	fmt.Println("    DELETE /api/v1/llms/:id          - Delete LLM")
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
	fmt.Println("    GET    /api/v1/health            - Health check")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the server")

	address := fmt.Sprintf("%s:%s", apiHost, apiPort)
	return server.Run(address)
}
