package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

// StatsService provides business logic for statistics
type StatsService struct {
	db db.Database
}

// NewStatsService creates a new stats service
func NewStatsService(database db.Database) *StatsService {
	return &StatsService{db: database}
}

// GetTopKeywords returns the top keywords by mention count
func (s *StatsService) GetTopKeywords(ctx context.Context, limit int, startTime, endTime *time.Time) ([]models.KeywordCount, error) {
	return s.db.GetTopKeywords(ctx, limit, startTime, endTime)
}

// SearchKeyword returns statistics for a specific keyword
func (s *StatsService) SearchKeyword(ctx context.Context, keyword string, startTime, endTime *time.Time) (*models.KeywordStats, error) {
	return s.db.SearchKeyword(ctx, keyword, startTime, endTime)
}

// GetKeywordTrends returns keyword trends over time - placeholder for future implementation
func (s *StatsService) GetKeywordTrends(ctx context.Context, keyword string, startTime, endTime time.Time) ([]models.TimeSeriesPoint, error) {
	// TODO: Implement GetKeywordTrends in database interface
	return nil, fmt.Errorf("keyword trends not implemented")
}

// GetPromptStats returns statistics for prompts - placeholder for future implementation
func (s *StatsService) GetPromptStats(ctx context.Context) ([]*models.PromptStats, error) {
	// TODO: Implement GetPromptStats in database interface
	return nil, fmt.Errorf("prompt stats not implemented")
}

// GetLLMStats returns statistics for LLMs - placeholder for future implementation
func (s *StatsService) GetLLMStats(ctx context.Context) ([]*models.LLMStats, error) {
	// TODO: Implement GetLLMStats in database interface
	return nil, fmt.Errorf("LLM stats not implemented")
}

// GetOverallStats returns overall system statistics
func (s *StatsService) GetOverallStats(ctx context.Context) (*OverallStats, error) {
	// Get basic counts
	prompts, err := s.db.ListPrompts(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompts: %w", err)
	}

	llms, err := s.db.ListLLMs(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLMs: %w", err)
	}

	schedules, err := s.db.ListSchedules(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedules: %w", err)
	}

	responses, err := s.db.ListResponses(ctx, shared.ResponseFilter{Limit: 1})
	if err != nil {
		return nil, fmt.Errorf("failed to get responses: %w", err)
	}

	// Count enabled items
	enabledPrompts := 0
	for _, prompt := range prompts {
		if prompt.Enabled {
			enabledPrompts++
		}
	}

	enabledLLMs := 0
	for _, llm := range llms {
		if llm.Enabled {
			enabledLLMs++
		}
	}

	enabledSchedules := 0
	for _, schedule := range schedules {
		if schedule.Enabled {
			enabledSchedules++
		}
	}

	return &OverallStats{
		TotalPrompts:     len(prompts),
		EnabledPrompts:   enabledPrompts,
		TotalLLMs:        len(llms),
		EnabledLLMs:      enabledLLMs,
		TotalSchedules:   len(schedules),
		EnabledSchedules: enabledSchedules,
		TotalResponses:   len(responses),
	}, nil
}

// OverallStats represents overall system statistics
type OverallStats struct {
	TotalPrompts     int `json:"total_prompts"`
	EnabledPrompts   int `json:"enabled_prompts"`
	TotalLLMs        int `json:"total_llms"`
	EnabledLLMs      int `json:"enabled_llms"`
	TotalSchedules   int `json:"total_schedules"`
	EnabledSchedules int `json:"enabled_schedules"`
	TotalResponses   int `json:"total_responses"`
}

// GetProviderStats returns statistics by provider
func (s *StatsService) GetProviderStats(ctx context.Context) (map[string]*ProviderStats, error) {
	responses, err := s.db.ListResponses(ctx, shared.ResponseFilter{Limit: 10000})
	if err != nil {
		return nil, fmt.Errorf("failed to get responses: %w", err)
	}

	providerStats := make(map[string]*ProviderStats)

	for _, response := range responses {
		if providerStats[response.LLMProvider] == nil {
			providerStats[response.LLMProvider] = &ProviderStats{
				Provider:       response.LLMProvider,
				TotalResponses: 0,
				TotalTokens:    0,
				TotalLatency:   0,
				UniquePrompts:  make(map[string]bool),
				UniqueLLMs:     make(map[string]bool),
			}
		}

		stats := providerStats[response.LLMProvider]
		stats.TotalResponses++
		stats.TotalTokens += response.TokensUsed
		stats.TotalLatency += response.LatencyMs
		stats.UniquePrompts[response.PromptID] = true
		stats.UniqueLLMs[response.LLMID] = true
	}

	// Calculate averages
	for _, stats := range providerStats {
		if stats.TotalResponses > 0 {
			stats.AvgTokens = float64(stats.TotalTokens) / float64(stats.TotalResponses)
			stats.AvgLatency = float64(stats.TotalLatency) / float64(stats.TotalResponses)
		}
		stats.UniquePromptCount = len(stats.UniquePrompts)
		stats.UniqueLLMCount = len(stats.UniqueLLMs)
	}

	return providerStats, nil
}

// ProviderStats represents statistics for a provider
type ProviderStats struct {
	Provider          string          `json:"provider"`
	TotalResponses    int             `json:"total_responses"`
	TotalTokens       int             `json:"total_tokens"`
	TotalLatency      int64           `json:"total_latency"`
	AvgTokens         float64         `json:"avg_tokens"`
	AvgLatency        float64         `json:"avg_latency"`
	UniquePromptCount int             `json:"unique_prompt_count"`
	UniqueLLMCount    int             `json:"unique_llm_count"`
	UniquePrompts     map[string]bool `json:"-"`
	UniqueLLMs        map[string]bool `json:"-"`
}

// GetTopPromptsByMentions returns prompts ranked by keyword mentions
func (s *StatsService) GetTopPromptsByMentions(ctx context.Context, limit int) ([]*PromptMentionStats, error) {
	// Get all responses
	responses, err := s.db.ListResponses(ctx, shared.ResponseFilter{Limit: 10000})
	if err != nil {
		return nil, fmt.Errorf("failed to get responses: %w", err)
	}

	// Count mentions per prompt
	promptMentions := make(map[string]int)
	promptNames := make(map[string]string)

	for _, response := range responses {
		promptMentions[response.PromptID]++

		// Get prompt name for display
		if _, exists := promptNames[response.PromptID]; !exists {
			if prompt, err := s.db.GetPrompt(ctx, response.PromptID); err == nil {
				promptNames[response.PromptID] = prompt.Template
			} else {
				promptNames[response.PromptID] = fmt.Sprintf("Unknown Prompt (%s)", response.PromptID[:8])
			}
		}
	}

	// Convert to slice and sort
	var results []*PromptMentionStats
	for promptID, count := range promptMentions {
		results = append(results, &PromptMentionStats{
			PromptID:   promptID,
			PromptName: promptNames[promptID],
			Mentions:   count,
		})
	}

	// Sort by mentions (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Mentions > results[j].Mentions
	})

	// Limit results
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// PromptMentionStats represents mention statistics for a prompt
type PromptMentionStats struct {
	PromptID   string `json:"prompt_id"`
	PromptName string `json:"prompt_name"`
	Mentions   int    `json:"mentions"`
}

// GetTopLLMsByMentions returns LLMs ranked by keyword mentions
func (s *StatsService) GetTopLLMsByMentions(ctx context.Context, limit int) ([]*LLMMentionStats, error) {
	// Get all responses
	responses, err := s.db.ListResponses(ctx, shared.ResponseFilter{Limit: 10000})
	if err != nil {
		return nil, fmt.Errorf("failed to get responses: %w", err)
	}

	// Count mentions per LLM
	llmMentions := make(map[string]int)
	llmNames := make(map[string]string)

	for _, response := range responses {
		llmMentions[response.LLMID]++

		// Get LLM name for display
		if _, exists := llmNames[response.LLMID]; !exists {
			if llm, err := s.db.GetLLM(ctx, response.LLMID); err == nil {
				llmNames[response.LLMID] = fmt.Sprintf("%s (%s)", llm.Name, llm.Provider)
			} else {
				llmNames[response.LLMID] = fmt.Sprintf("Unknown LLM (%s)", response.LLMID[:8])
			}
		}
	}

	// Convert to slice and sort
	var results []*LLMMentionStats
	for llmID, count := range llmMentions {
		results = append(results, &LLMMentionStats{
			LLMID:    llmID,
			LLMName:  llmNames[llmID],
			Mentions: count,
		})
	}

	// Sort by mentions (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Mentions > results[j].Mentions
	})

	// Limit results
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// LLMMentionStats represents mention statistics for an LLM
type LLMMentionStats struct {
	LLMID    string `json:"llm_id"`
	LLMName  string `json:"llm_name"`
	Mentions int    `json:"mentions"`
}

// ResetAllStats resets all statistics by clearing all responses
func (s *StatsService) ResetAllStats(ctx context.Context) (int, error) {
	return s.db.DeleteAllResponses(ctx)
}
