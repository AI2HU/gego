package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/auth"
	"github.com/AI2HU/gego/internal/config"
	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/llm"
	"github.com/AI2HU/gego/internal/llm/anthropic"
	"github.com/AI2HU/gego/internal/llm/google"
	"github.com/AI2HU/gego/internal/llm/ollama"
	"github.com/AI2HU/gego/internal/llm/openai"
	"github.com/AI2HU/gego/internal/llm/perplexity"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/services"
)

// Server represents the API server
type Server struct {
	db              db.Database
	llmService      *services.LLMService
	promptService   *services.PromptManagementService
	scheduleService  *services.ScheduleService
	schedulerService *services.SchedulerService
	statsService     *services.StatsService
	searchService   *services.SearchService
	authService     *services.AuthService
	authMiddleware  *auth.Middleware
	llmRegistry     *llm.Registry
	router          *gin.Engine
	corsOrigin      string
}

// NewServer creates a new API server
func NewServer(database db.Database, corsOrigin string, authConfig auth.Config) (*Server, error) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	allowedOrigins := parseAllowedOrigins(corsOrigin)

	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowedOrigin := getAllowedOrigin(origin, allowedOrigins, corsOrigin)

		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	authMW, err := auth.NewMiddleware(authConfig)
	if err != nil {
		return nil, err
	}

	llmRegistry := llm.NewRegistry()
	llmRegistry.Register(openai.New("", "", config.GetSystemInstruction(nil, config.ProviderChatGPT)))
	llmRegistry.Register(anthropic.New("", ""))
	llmRegistry.Register(ollama.New(""))
	llmRegistry.Register(google.New("", "", config.GetSystemInstruction(nil, config.ProviderGemini)))
	llmRegistry.Register(perplexity.New("", ""))

	server := &Server{
		db:              database,
		llmService:      services.NewLLMService(database),
		promptService:   services.NewPromptManagementService(database),
		scheduleService:  services.NewScheduleService(database),
		schedulerService: services.NewSchedulerService(database, llmRegistry),
		statsService:     services.NewStatsService(database),
		searchService:   services.NewSearchService(database),
		authService:     services.NewAuthService(database, authConfig),
		authMiddleware:  authMW,
		llmRegistry:     llmRegistry,
		router:          router,
		corsOrigin:      corsOrigin,
	}

	server.setupRoutes()
	return server, nil
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	api := s.router.Group("/api/v1")

	api.GET("/health", s.healthCheck)
	api.POST("/auth/login", s.login)
	api.POST("/auth/refresh", s.refresh)
	api.POST("/auth/logout", s.logout)

	protected := api.Group("")
	protected.Use(s.authMiddleware.Authenticate())

	protected.GET("/providers", s.requirePerm(auth.PermLLMsRead), s.listProviders)
	protected.GET("/providers/:provider/api-keys", s.requirePerm(auth.PermLLMsRead), s.listProviderAPIKeys)
	protected.POST("/providers/:provider/models", s.requirePerm(auth.PermLLMsWrite), s.listProviderModels)
	protected.GET("/models", s.requirePerm(auth.PermLLMsRead), s.listLLMs)
	protected.GET("/models/:id", s.requirePerm(auth.PermLLMsRead), s.getLLM)
	protected.POST("/models", s.requirePerm(auth.PermLLMsWrite), s.createLLM)
	protected.PUT("/models/:id", s.requirePerm(auth.PermLLMsWrite), s.updateLLM)
	protected.DELETE("/models/:id", s.requirePerm(auth.PermLLMsWrite), s.deleteLLM)

	protected.GET("/prompts", s.requirePerm(auth.PermPromptsRead), s.listPrompts)
	protected.POST("/prompts/generate", s.requirePerm(auth.PermPromptsWrite), s.generatePrompts)
	protected.GET("/prompts/:id", s.requirePerm(auth.PermPromptsRead), s.getPrompt)
	protected.POST("/prompts", s.requirePerm(auth.PermPromptsWrite), s.createPrompt)
	protected.PUT("/prompts/:id", s.requirePerm(auth.PermPromptsWrite), s.updatePrompt)
	protected.DELETE("/prompts/:id", s.requirePerm(auth.PermPromptsWrite), s.deletePrompt)

	protected.GET("/schedules", s.requirePerm(auth.PermSchedulesRead), s.listSchedules)
	protected.GET("/schedules/:id", s.requirePerm(auth.PermSchedulesRead), s.getSchedule)
	protected.POST("/schedules", s.requirePerm(auth.PermSchedulesWrite), s.createSchedule)
	protected.PUT("/schedules/:id", s.requirePerm(auth.PermSchedulesWrite), s.updateSchedule)
	protected.DELETE("/schedules/:id", s.requirePerm(auth.PermSchedulesWrite), s.deleteSchedule)
	protected.POST("/schedules/:id/run", s.requirePerm(auth.PermSchedulesWrite), s.runSchedule)

	protected.GET("/scheduler/status", s.requirePerm(auth.PermSchedulesRead), s.getSchedulerStatus)
	protected.POST("/scheduler/start", s.requirePerm(auth.PermSchedulesWrite), s.startScheduler)
	protected.POST("/scheduler/stop", s.requirePerm(auth.PermSchedulesWrite), s.stopScheduler)
	protected.POST("/scheduler/reload", s.requirePerm(auth.PermSchedulesWrite), s.reloadScheduler)

	protected.GET("/stats", s.requirePerm(auth.PermStatsRead), s.getStats)
	protected.GET("/stats/urls", s.requirePerm(auth.PermStatsRead), s.getURLStats)
	protected.GET("/stats/query-urls", s.requirePerm(auth.PermStatsRead), s.getQueryURLStats)
	protected.GET("/stats/keyword-domains", s.requirePerm(auth.PermStatsRead), s.getKeywordDomainMatrix)

	protected.POST("/search", s.requirePerm(auth.PermSearchExecute), s.search)
	protected.GET("/auth/me", s.requirePerm(auth.PermAuthProfile), s.me)
}

// Run starts the API server
func (s *Server) Run(address string) error {
	return s.router.Run(address)
}

// Helper functions
func (s *Server) successResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    data,
	})
}

func (s *Server) errorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, models.APIResponse{
		Success: false,
		Error:   message,
	})
}

func (s *Server) parsePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	return page, limit
}

func parseAllowedOrigins(corsOrigin string) []string {
	if corsOrigin == "" || corsOrigin == "*" {
		return nil
	}

	origins := strings.Split(corsOrigin, ",")
	allowed := make([]string, 0, len(origins))
	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			allowed = append(allowed, trimmed)
		}
	}
	return allowed
}

func getAllowedOrigin(requestOrigin string, allowedOrigins []string, corsOrigin string) string {
	if corsOrigin == "*" {
		return "*"
	}

	if requestOrigin == "" {
		return ""
	}

	for _, allowed := range allowedOrigins {
		if requestOrigin == allowed {
			return allowed
		}
	}

	return ""
}
