package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/services"
)

// getStats handles GET /api/v1/stats
func (s *Server) getStats(c *gin.Context) {
	limitStr := c.DefaultQuery("keyword_limit", "10")
	keywordLimit, _ := strconv.Atoi(limitStr)
	if keywordLimit <= 0 || keywordLimit > 100 {
		keywordLimit = 10
	}

	tags := parseTagsFromQuery(c)
	promptIDs, err := s.resolvePromptIDsByTags(c.Request.Context(), tags)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to resolve prompt tags: "+err.Error())
		return
	}

	stats, err := s.statsService.GetDashboardStats(c.Request.Context(), services.DashboardStatsOptions{
		PromptIDs:    promptIDs,
		TagFilter:    len(tags) > 0,
		KeywordLimit: keywordLimit,
	})
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get stats: "+err.Error())
		return
	}

	s.successResponse(c, stats)
}

type URLStatsResponse struct {
	TopURLs    []*services.URLMentionStats    `json:"top_urls"`
	TopDomains []*services.DomainMentionStats `json:"top_domains"`
}

func (s *Server) getURLStats(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	tags := parseTagsFromQuery(c)
	promptIDs, err := s.resolvePromptIDsByTags(c.Request.Context(), tags)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to resolve prompt tags: "+err.Error())
		return
	}

	if len(tags) > 0 && len(promptIDs) == 0 {
		s.successResponse(c, URLStatsResponse{
			TopURLs:    []*services.URLMentionStats{},
			TopDomains: []*services.DomainMentionStats{},
		})
		return
	}

	topURLs, err := s.statsService.GetTopURLsByCitations(c.Request.Context(), limit, promptIDs)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get top URLs: "+err.Error())
		return
	}

	topDomains, err := s.statsService.GetTopDomainsByCitations(c.Request.Context(), limit, promptIDs)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get top domains: "+err.Error())
		return
	}

	s.successResponse(c, URLStatsResponse{
		TopURLs:    topURLs,
		TopDomains: topDomains,
	})
}

func (s *Server) getQueryURLStats(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	stats, err := s.statsService.GetQueryURLRelationships(c.Request.Context(), limit)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get query URL relationships: "+err.Error())
		return
	}

	s.successResponse(c, stats)
}

func (s *Server) getKeywordDomainMatrix(c *gin.Context) {
	keywordLimitStr := c.DefaultQuery("keyword_limit", "20")
	domainLimitStr := c.DefaultQuery("domain_limit", "10")

	keywordLimit, _ := strconv.Atoi(keywordLimitStr)
	if keywordLimit <= 0 || keywordLimit > 200 {
		keywordLimit = 20
	}

	domainLimit, _ := strconv.Atoi(domainLimitStr)
	if domainLimit <= 0 || domainLimit > 100 {
		domainLimit = 10
	}

	stats, err := s.statsService.GetKeywordDomainMatrix(c.Request.Context(), keywordLimit, domainLimit)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get keyword-domain matrix: "+err.Error())
		return
	}

	s.successResponse(c, stats)
}

func (s *Server) getBrandCitationDomains(c *gin.Context) {
	brandID := strings.TrimSpace(c.Query("brand_id"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	if brandID == "" && keyword == "" {
		s.errorResponse(c, http.StatusBadRequest, "brand_id or keyword is required")
		return
	}
	if brandID != "" && keyword != "" {
		s.errorResponse(c, http.StatusBadRequest, "provide either brand_id or keyword, not both")
		return
	}

	limitStr := c.DefaultQuery("limit", "15")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 15
	}

	tags := parseTagsFromQuery(c)
	promptIDs, err := s.resolvePromptIDsByTags(c.Request.Context(), tags)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to resolve prompt tags: "+err.Error())
		return
	}

	if len(tags) > 0 && len(promptIDs) == 0 {
		empty := &services.BrandCitationDomainsResult{
			Domains: []*services.BrandCitationDomainStats{},
		}
		if brandID != "" {
			empty.BrandID = brandID
		} else {
			empty.BrandName = keyword
		}
		s.successResponse(c, empty)
		return
	}

	stats, err := s.statsService.GetBrandCitationDomains(
		c.Request.Context(),
		brandID,
		keyword,
		limit,
		promptIDs,
	)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.errorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get brand citation domains: "+err.Error())
		return
	}

	s.successResponse(c, stats)
}

// healthCheck handles GET /api/v1/health
func (s *Server) healthCheck(c *gin.Context) {
	if err := s.db.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Error:   "Database connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now(),
			"version":   "1.0.0",
		},
	})
}
