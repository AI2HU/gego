package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/services"
)

// Server represents the API server
type Server struct {
	db              db.Database
	llmService      *services.LLMService
	promptService   *services.PromptManagementService
	scheduleService *services.ScheduleService
	statsService    *services.StatsService
	searchService   *services.SearchService
	router          *gin.Engine
	corsOrigin      string
}

// NewServer creates a new API server
func NewServer(database db.Database, corsOrigin string) *Server {
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

	server := &Server{
		db:              database,
		llmService:      services.NewLLMService(database),
		promptService:   services.NewPromptManagementService(database),
		scheduleService: services.NewScheduleService(database),
		statsService:    services.NewStatsService(database),
		searchService:   services.NewSearchService(database),
		router:          router,
		corsOrigin:      corsOrigin,
	}

	server.setupRoutes()
	return server
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	api := s.router.Group("/api/v1")

	api.GET("/llms", s.listLLMs)
	api.GET("/llms/:id", s.getLLM)
	// api.POST("/llms", s.createLLM)
	// api.PUT("/llms/:id", s.updateLLM)
	// api.DELETE("/llms/:id", s.deleteLLM)

	api.GET("/prompts", s.listPrompts)
	api.GET("/prompts/:id", s.getPrompt)
	// api.POST("/prompts", s.createPrompt)
	// api.PUT("/prompts/:id", s.updatePrompt)
	// api.DELETE("/prompts/:id", s.deletePrompt)

	api.GET("/schedules", s.listSchedules)
	api.GET("/schedules/:id", s.getSchedule)
	// api.POST("/schedules", s.createSchedule)
	// api.PUT("/schedules/:id", s.updateSchedule)
	// api.DELETE("/schedules/:id", s.deleteSchedule)

	api.GET("/stats", s.getStats)

	api.POST("/search", s.search)

	api.GET("/health", s.healthCheck)
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
