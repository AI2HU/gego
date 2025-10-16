package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

// SearchService provides business logic for searching responses
type SearchService struct {
	db db.Database
}

// NewSearchService creates a new search service
func NewSearchService(database db.Database) *SearchService {
	return &SearchService{db: database}
}

// SearchKeyword searches for a keyword and returns statistics
func (s *SearchService) SearchKeyword(ctx context.Context, keyword string, startTime, endTime *time.Time) (*models.KeywordStats, error) {
	return s.db.SearchKeyword(ctx, keyword, startTime, endTime)
}

// ListResponses lists responses with filtering
func (s *SearchService) ListResponses(ctx context.Context, filter shared.ResponseFilter) ([]*models.Response, error) {
	return s.db.ListResponses(ctx, filter)
}

// SearchMatch represents a search match in a response
type SearchMatch struct {
	ResponseID  string    `json:"response_id"`
	PromptID    string    `json:"prompt_id"`
	PromptName  string    `json:"prompt_name"`
	FullPrompt  string    `json:"full_prompt"`
	LLMName     string    `json:"llm_name"`
	LLMProvider string    `json:"llm_provider"`
	Temperature float64   `json:"temperature"`
	Context     string    `json:"context"`
	CreatedAt   time.Time `json:"created_at"`
}

// SearchConfig represents configuration for search operations
type SearchConfig struct {
	Keyword       string `json:"keyword"`
	CaseSensitive bool   `json:"case_sensitive"`
	ContextLength int    `json:"context_length"`
	Limit         int    `json:"limit"`
}

// DefaultSearchConfig returns default search configuration
func DefaultSearchConfig() *SearchConfig {
	return &SearchConfig{
		CaseSensitive: false,
		ContextLength: 100,
		Limit:         50,
	}
}

// SearchResponses searches for keywords in responses
func (s *SearchService) SearchResponses(ctx context.Context, config *SearchConfig) ([]SearchMatch, error) {
	if config.Keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}

	// Use the database interface to search for responses containing the keyword
	filter := shared.ResponseFilter{
		Keyword: config.Keyword,
		Limit:   config.Limit * 10, // Get more responses to find matches
	}

	responses, err := s.db.ListResponses(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search responses: %w", err)
	}

	// Create regex for case-sensitive/insensitive matching
	var regex *regexp.Regexp
	if config.CaseSensitive {
		regex = regexp.MustCompile(regexp.QuoteMeta(config.Keyword))
	} else {
		regex = regexp.MustCompile("(?i)" + regexp.QuoteMeta(config.Keyword))
	}

	// Process matches
	var matches []SearchMatch
	for _, response := range responses {
		responseMatches := s.findMatches(response, regex, config.ContextLength)
		matches = append(matches, responseMatches...)
	}

	return matches, nil
}

// findMatches finds all matches in a response
func (s *SearchService) findMatches(response *models.Response, regex *regexp.Regexp, contextLength int) []SearchMatch {
	var matches []SearchMatch

	// Find all matches in the response text
	indices := regex.FindAllStringIndex(response.ResponseText, -1)

	for _, index := range indices {
		start := index[0]
		end := index[1]

		// Get context: contextLength chars before and after
		contextStart := start - contextLength
		if contextStart < 0 {
			contextStart = 0
		}
		contextEnd := end + contextLength
		if contextEnd > len(response.ResponseText) {
			contextEnd = len(response.ResponseText)
		}

		contextText := response.ResponseText[contextStart:contextEnd]

		// Get prompt name
		promptName := "Unknown Prompt"
		if prompt, err := s.db.GetPrompt(context.Background(), response.PromptID); err == nil {
			promptName = prompt.Template
		}

		matches = append(matches, SearchMatch{
			ResponseID:  response.ID,
			PromptID:    response.PromptID,
			PromptName:  promptName,
			FullPrompt:  response.PromptText,
			LLMName:     response.LLMName,
			LLMProvider: response.LLMProvider,
			Temperature: response.Temperature,
			Context:     contextText,
			CreatedAt:   response.CreatedAt,
		})
	}

	return matches
}

// SearchByPrompt searches responses by prompt ID
func (s *SearchService) SearchByPrompt(ctx context.Context, promptID string, limit int) ([]*models.Response, error) {
	filter := shared.ResponseFilter{
		PromptID: promptID,
		Limit:    limit,
	}
	return s.db.ListResponses(ctx, filter)
}

// SearchByLLM searches responses by LLM ID
func (s *SearchService) SearchByLLM(ctx context.Context, llmID string, limit int) ([]*models.Response, error) {
	filter := shared.ResponseFilter{
		LLMID: llmID,
		Limit: limit,
	}
	return s.db.ListResponses(ctx, filter)
}

// SearchByDateRange searches responses within a date range
func (s *SearchService) SearchByDateRange(ctx context.Context, startTime, endTime time.Time, limit int) ([]*models.Response, error) {
	filter := shared.ResponseFilter{
		StartTime: &startTime,
		EndTime:   &endTime,
		Limit:     limit,
	}
	return s.db.ListResponses(ctx, filter)
}

// SearchBySchedule searches responses by schedule ID
func (s *SearchService) SearchBySchedule(ctx context.Context, scheduleID string, limit int) ([]*models.Response, error) {
	filter := shared.ResponseFilter{
		ScheduleID: scheduleID,
		Limit:      limit,
	}
	return s.db.ListResponses(ctx, filter)
}

// GetResponseStats returns statistics about responses
func (s *SearchService) GetResponseStats(ctx context.Context) (*ResponseStats, error) {
	// Get all responses for basic stats
	responses, err := s.db.ListResponses(ctx, shared.ResponseFilter{Limit: 10000})
	if err != nil {
		return nil, err
	}

	stats := &ResponseStats{
		TotalResponses: len(responses),
		ByProvider:     make(map[string]int),
		ByPrompt:       make(map[string]int),
		ByLLM:          make(map[string]int),
		TotalTokens:    0,
		AvgLatency:     0,
	}

	var totalLatency int64
	for _, response := range responses {
		stats.ByProvider[response.LLMProvider]++
		stats.ByPrompt[response.PromptID]++
		stats.ByLLM[response.LLMID]++
		stats.TotalTokens += response.TokensUsed
		totalLatency += response.LatencyMs
	}

	if len(responses) > 0 {
		stats.AvgLatency = float64(totalLatency) / float64(len(responses))
	}

	return stats, nil
}

// ResponseStats represents statistics about responses
type ResponseStats struct {
	TotalResponses int            `json:"total_responses"`
	ByProvider     map[string]int `json:"by_provider"`
	ByPrompt       map[string]int `json:"by_prompt"`
	ByLLM          map[string]int `json:"by_llm"`
	TotalTokens    int            `json:"total_tokens"`
	AvgLatency     float64        `json:"avg_latency"`
}

// HighlightKeyword highlights a keyword in text
func HighlightKeyword(text, keyword string, caseSensitive bool) string {
	if caseSensitive {
		return strings.ReplaceAll(text, keyword, fmt.Sprintf("**%s**", keyword))
	}

	// Case insensitive replacement
	regex := regexp.MustCompile("(?i)" + regexp.QuoteMeta(keyword))
	return regex.ReplaceAllStringFunc(text, func(match string) string {
		return fmt.Sprintf("**%s**", match)
	})
}

// ValidateSearchConfig validates search configuration
func ValidateSearchConfig(config *SearchConfig) error {
	if config.Keyword == "" {
		return fmt.Errorf("keyword is required")
	}
	if config.ContextLength < 0 {
		return fmt.Errorf("context length must be non-negative")
	}
	if config.Limit <= 0 {
		return fmt.Errorf("limit must be positive")
	}
	return nil
}
