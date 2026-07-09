package services

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/llm"
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

// GetTotalResponses returns the total number of responses
func (s *StatsService) GetTotalResponses(ctx context.Context) (int64, error) {
	return s.db.CountResponses(ctx, shared.ResponseFilter{})
}

// GetTotalPrompts returns the total number of prompts
func (s *StatsService) GetTotalPrompts(ctx context.Context) (int64, error) {
	prompts, err := s.db.ListPrompts(ctx, nil)
	if err != nil {
		return 0, err
	}
	return int64(len(prompts)), nil
}

// GetTotalLLMs returns the total number of LLMs
func (s *StatsService) GetTotalLLMs(ctx context.Context) (int64, error) {
	llms, err := s.db.ListLLMs(ctx, nil)
	if err != nil {
		return 0, err
	}
	return int64(len(llms)), nil
}

// GetTotalSchedules returns the total number of schedules
func (s *StatsService) GetTotalSchedules(ctx context.Context) (int64, error) {
	schedules, err := s.db.ListSchedules(ctx, nil)
	if err != nil {
		return 0, err
	}
	return int64(len(schedules)), nil
}

// GetResponseTrends returns response trends over time
func (s *StatsService) GetResponseTrends(ctx context.Context, startTime, endTime time.Time) ([]models.TimeSeriesPoint, error) {
	// This is a placeholder implementation
	// In a real implementation, you would query the database for responses within the time range
	// and group them by time intervals (e.g., daily, hourly)
	return []models.TimeSeriesPoint{}, nil
}

// GetTopKeywords returns the top keywords by mention count
func (s *StatsService) GetTopKeywords(ctx context.Context, limit int, startTime, endTime *time.Time, promptIDs []string) ([]models.KeywordCount, error) {
	if len(promptIDs) > 0 {
		responses, err := s.db.ListResponses(ctx, shared.ResponseFilter{PromptIDs: promptIDs, Limit: 10000})
		if err != nil {
			return nil, err
		}
		return aggregateTopKeywordsFromResponses(responses, limit), nil
	}

	return s.db.GetTopKeywords(ctx, limit, startTime, endTime)
}

// SearchKeyword returns statistics for a specific keyword
func (s *StatsService) SearchKeyword(ctx context.Context, keyword string, startTime, endTime *time.Time, promptIDs []string) (*models.KeywordStats, error) {
	return s.db.SearchKeyword(ctx, keyword, startTime, endTime, promptIDs)
}

// GetKeywordTrends returns keyword trends over time - placeholder for future implementation
func (s *StatsService) GetKeywordTrends(ctx context.Context, keyword string, startTime, endTime time.Time) ([]models.TimeSeriesPoint, error) {
	// TODO: Implement GetKeywordTrends in database interface
	return nil, fmt.Errorf("keyword trends not implemented")
}

// GetPromptStats returns statistics for a specific prompt
func (s *StatsService) GetPromptStats(ctx context.Context, promptID string) (*models.PromptStats, error) {
	return s.db.GetPromptStats(ctx, promptID)
}

// GetLLMStats returns statistics for a specific LLM
func (s *StatsService) GetLLMStats(ctx context.Context, llmID string) (*models.LLMStats, error) {
	return s.db.GetLLMStats(ctx, llmID)
}

// GetAllPromptStats returns statistics for all prompts
func (s *StatsService) GetAllPromptStats(ctx context.Context) ([]*models.PromptStats, error) {
	prompts, err := s.db.ListPrompts(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompts: %w", err)
	}

	var allStats []*models.PromptStats
	for _, prompt := range prompts {
		stats, err := s.db.GetPromptStats(ctx, prompt.ID)
		if err != nil {
			continue
		}
		allStats = append(allStats, stats)
	}

	return allStats, nil
}

// GetAllLLMStats returns statistics for all LLMs
func (s *StatsService) GetAllLLMStats(ctx context.Context) ([]*models.LLMStats, error) {
	llms, err := s.db.ListLLMs(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLMs: %w", err)
	}

	var allStats []*models.LLMStats
	for _, llm := range llms {
		stats, err := s.db.GetLLMStats(ctx, llm.ID)
		if err != nil {
			continue
		}
		allStats = append(allStats, stats)
	}

	return allStats, nil
}

// GetOverallStats returns overall system statistics
func (s *StatsService) GetOverallStats(ctx context.Context) (*OverallStats, error) {
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
				UniquePrompts:  make(map[string]bool),
				UniqueLLMs:     make(map[string]bool),
			}
		}

		stats := providerStats[response.LLMProvider]
		stats.TotalResponses++
		stats.TotalTokens += response.TokensUsed
		stats.UniquePrompts[response.PromptID] = true
		stats.UniqueLLMs[response.LLMID] = true
	}

	for _, stats := range providerStats {
		if stats.TotalResponses > 0 {
			stats.AvgTokens = float64(stats.TotalTokens) / float64(stats.TotalResponses)
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
	AvgTokens         float64         `json:"avg_tokens"`
	UniquePromptCount int             `json:"unique_prompt_count"`
	UniqueLLMCount    int             `json:"unique_llm_count"`
	UniquePrompts     map[string]bool `json:"-"`
	UniqueLLMs        map[string]bool `json:"-"`
}

// GetTopURLsByCitations returns URLs ranked by how often they are cited
func (s *StatsService) GetTopURLsByCitations(ctx context.Context, limit int, promptIDs []string) ([]*URLMentionStats, error) {
	filter := shared.ResponseFilter{Limit: 10000}
	if len(promptIDs) > 0 {
		filter.PromptIDs = promptIDs
	}

	responses, err := s.db.ListResponses(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get responses: %w", err)
	}

	urlStats := make(map[string]*URLMentionStats)

	for _, response := range responses {
		for _, url := range response.SearchURLs {
			if url.URL == "" {
				continue
			}

			stats, exists := urlStats[url.URL]
			if !exists {
				stats = &URLMentionStats{
					URL: url.URL,
				}
				if url.Title != "" {
					stats.Title = url.Title
				}
				if url.SearchQuery != "" {
					stats.SearchQuery = url.SearchQuery
				}
				urlStats[url.URL] = stats
			}

			stats.Citations++
		}
	}

	var results []*URLMentionStats
	for _, stats := range urlStats {
		results = append(results, stats)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Citations == results[j].Citations {
			return results[i].URL < results[j].URL
		}
		return results[i].Citations > results[j].Citations
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	if results == nil {
		return []*URLMentionStats{}, nil
	}

	return results, nil
}

// URLMentionStats represents citation statistics for a URL
type URLMentionStats struct {
	URL         string `json:"url"`
	Title       string `json:"title,omitempty"`
	SearchQuery string `json:"search_query,omitempty"`
	Citations   int    `json:"citations"`
}

type DomainMentionStats struct {
	Domain         string `json:"domain"`
	Citations      int    `json:"citations"`
	UniqueURLCount int    `json:"unique_url_count"`
}

type QueryURLItem struct {
	URL       string `json:"url"`
	Title     string `json:"title,omitempty"`
	Citations int    `json:"citations"`
}

type QueryURLStats struct {
	Query          string         `json:"query"`
	TotalCitations int            `json:"total_citations"`
	URLs           []QueryURLItem `json:"urls"`
}

type DomainCount struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

type KeywordDomainStats struct {
	Keyword string        `json:"keyword"`
	Total   int           `json:"total"`
	Domains []DomainCount `json:"domains"`
}

func (s *StatsService) GetTopDomainsByCitations(ctx context.Context, limit int, promptIDs []string) ([]*DomainMentionStats, error) {
	filter := shared.ResponseFilter{Limit: 10000}
	if len(promptIDs) > 0 {
		filter.PromptIDs = promptIDs
	}

	responses, err := s.db.ListResponses(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get responses: %w", err)
	}

	domainStats := make(map[string]*DomainMentionStats)
	domainURLs := make(map[string]map[string]bool)

	for _, response := range responses {
		for _, u := range response.SearchURLs {
			if u.URL == "" {
				continue
			}

			domain := extractDomain(u.URL)
			if domain == "" {
				continue
			}

			if domainStats[domain] == nil {
				domainStats[domain] = &DomainMentionStats{
					Domain:    domain,
					Citations: 0,
				}
				domainURLs[domain] = make(map[string]bool)
			}

			stats := domainStats[domain]
			stats.Citations++
			domainURLs[domain][u.URL] = true
		}
	}

	var results []*DomainMentionStats
	for domain, stats := range domainStats {
		stats.UniqueURLCount = len(domainURLs[domain])
		results = append(results, stats)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Citations == results[j].Citations {
			return results[i].Domain < results[j].Domain
		}
		return results[i].Citations > results[j].Citations
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	if results == nil {
		return []*DomainMentionStats{}, nil
	}

	return results, nil
}

func (s *StatsService) GetQueryURLRelationships(ctx context.Context, limit int) ([]*QueryURLStats, error) {
	responses, err := s.db.ListResponses(ctx, shared.ResponseFilter{Limit: 10000})
	if err != nil {
		return nil, fmt.Errorf("failed to get responses: %w", err)
	}

	queryURLItems := make(map[string]map[string]*QueryURLItem)
	queryTotals := make(map[string]int)

	for _, response := range responses {
		for _, u := range response.SearchURLs {
			if u.URL == "" {
				continue
			}
			if u.SearchQuery == "" || u.SearchQuery == llm.UnknownSearchQuery {
				continue
			}

			query := u.SearchQuery

			if queryURLItems[query] == nil {
				queryURLItems[query] = make(map[string]*QueryURLItem)
			}

			urlItems := queryURLItems[query]
			item, exists := urlItems[u.URL]
			if !exists {
				item = &QueryURLItem{
					URL:   u.URL,
					Title: u.Title,
				}
				urlItems[u.URL] = item
			}

			item.Citations++
			queryTotals[query]++
		}
	}

	var results []*QueryURLStats
	for query, urls := range queryURLItems {
		items := make([]QueryURLItem, 0, len(urls))
		for _, item := range urls {
			items = append(items, *item)
		}

		sort.Slice(items, func(i, j int) bool {
			if items[i].Citations == items[j].Citations {
				return items[i].URL < items[j].URL
			}
			return items[i].Citations > items[j].Citations
		})

		results = append(results, &QueryURLStats{
			Query:          query,
			TotalCitations: queryTotals[query],
			URLs:           items,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].TotalCitations == results[j].TotalCitations {
			return results[i].Query < results[j].Query
		}
		return results[i].TotalCitations > results[j].TotalCitations
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

func (s *StatsService) GetKeywordDomainMatrix(ctx context.Context, keywordLimit, domainLimit int) ([]*KeywordDomainStats, error) {
	responses, err := s.db.ListResponses(ctx, shared.ResponseFilter{Limit: 10000})
	if err != nil {
		return nil, fmt.Errorf("failed to get responses: %w", err)
	}

	matrix := make(map[string]map[string]int)

	for _, response := range responses {
		if len(response.SearchURLs) == 0 {
			continue
		}

		keywords := shared.ExtractCapitalizedWords(response.ResponseText)
		if len(keywords) == 0 {
			continue
		}

		keywordSet := make(map[string]bool)
		for _, kw := range keywords {
			kw = strings.TrimSpace(kw)
			if kw == "" {
				continue
			}
			keywordSet[kw] = true
		}

		if len(keywordSet) == 0 {
			continue
		}

		domainSet := make(map[string]bool)
		for _, u := range response.SearchURLs {
			if u.URL == "" {
				continue
			}
			domain := extractDomain(u.URL)
			if domain == "" {
				continue
			}
			domainSet[domain] = true
		}

		if len(domainSet) == 0 {
			continue
		}

		for kw := range keywordSet {
			if matrix[kw] == nil {
				matrix[kw] = make(map[string]int)
			}
			for domain := range domainSet {
				matrix[kw][domain]++
			}
		}
	}

	var results []*KeywordDomainStats
	for kw, domains := range matrix {
		var (
			domainCounts []DomainCount
			total        int
		)

		for domain, count := range domains {
			total += count
			domainCounts = append(domainCounts, DomainCount{
				Domain: domain,
				Count:  count,
			})
		}

		sort.Slice(domainCounts, func(i, j int) bool {
			if domainCounts[i].Count == domainCounts[j].Count {
				return domainCounts[i].Domain < domainCounts[j].Domain
			}
			return domainCounts[i].Count > domainCounts[j].Count
		})

		if domainLimit > 0 && len(domainCounts) > domainLimit {
			domainCounts = domainCounts[:domainLimit]
		}

		results = append(results, &KeywordDomainStats{
			Keyword: kw,
			Total:   total,
			Domains: domainCounts,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Total == results[j].Total {
			return results[i].Keyword < results[j].Keyword
		}
		return results[i].Total > results[j].Total
	})

	if keywordLimit > 0 && len(results) > keywordLimit {
		results = results[:keywordLimit]
	}

	return results, nil
}

type BrandCitationDomainStats struct {
	Domain         string `json:"domain"`
	Citations      int    `json:"citations"`
	UniqueURLCount int    `json:"unique_url_count"`
}

type BrandCitationDomainsResult struct {
	BrandID   string                      `json:"brand_id"`
	BrandName string                      `json:"brand_name"`
	TotalHits int                         `json:"total_hits"`
	Domains   []*BrandCitationDomainStats `json:"domains"`
}

const brandCitationDomainsBatchSize = 500

func (s *StatsService) GetBrandCitationDomains(
	ctx context.Context,
	brandID string,
	keyword string,
	limit int,
	promptIDs []string,
) (*BrandCitationDomainsResult, error) {
	if brandID != "" {
		brand, err := s.db.GetBrand(ctx, brandID)
		if err != nil {
			return nil, err
		}
		return s.getBrandCitationDomainsForKeyword(ctx, brand.ID, brand.Name, limit, promptIDs)
	}

	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, fmt.Errorf("brand_id or keyword is required")
	}

	return s.getBrandCitationDomainsForKeyword(ctx, "", keyword, limit, promptIDs)
}

func (s *StatsService) getBrandCitationDomainsForKeyword(
	ctx context.Context,
	brandID string,
	keyword string,
	limit int,
	promptIDs []string,
) (*BrandCitationDomainsResult, error) {
	searchTerms := shared.ExpandBrandSearchTerms(keyword)
	if len(searchTerms) == 0 {
		return &BrandCitationDomainsResult{
			BrandID:   brandID,
			BrandName: keyword,
			Domains:   []*BrandCitationDomainStats{},
		}, nil
	}

	attributor := shared.NewBrandCitationAttributor(searchTerms)
	domainStats := make(map[string]*BrandCitationDomainStats)
	domainURLs := make(map[string]map[string]bool)
	totalHits := 0

	for offset := 0; ; offset += brandCitationDomainsBatchSize {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		batch, err := s.db.ListResponses(ctx, shared.ResponseFilter{
			Keyword:       keyword,
			HasSearchURLs: true,
			PromptIDs:     promptIDs,
			Limit:         brandCitationDomainsBatchSize,
			Offset:        offset,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get responses: %w", err)
		}
		if len(batch) == 0 {
			break
		}

		for _, response := range batch {
			if len(response.SearchURLs) == 0 || response.ResponseText == "" {
				continue
			}

			urlInputs := make([]shared.SearchURLInput, len(response.SearchURLs))
			for i, source := range response.SearchURLs {
				urlInputs[i] = shared.SearchURLInput{
					URL:           source.URL,
					CitationIndex: source.CitationIndex,
				}
			}

			hits := attributor.Attribute(response.ResponseText, urlInputs)
			for _, hit := range hits {
				if hit.Domain == "" {
					continue
				}

				if domainStats[hit.Domain] == nil {
					domainStats[hit.Domain] = &BrandCitationDomainStats{Domain: hit.Domain}
					domainURLs[hit.Domain] = make(map[string]bool)
				}

				domainStats[hit.Domain].Citations++
				if hit.URL != "" {
					domainURLs[hit.Domain][hit.URL] = true
				}
				totalHits++
			}
		}

		if len(batch) < brandCitationDomainsBatchSize {
			break
		}
	}

	results := make([]*BrandCitationDomainStats, 0, len(domainStats))
	for domain, stats := range domainStats {
		stats.UniqueURLCount = len(domainURLs[domain])
		results = append(results, stats)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Citations == results[j].Citations {
			return results[i].Domain < results[j].Domain
		}
		return results[i].Citations > results[j].Citations
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	if results == nil {
		results = []*BrandCitationDomainStats{}
	}

	return &BrandCitationDomainsResult{
		BrandID:   brandID,
		BrandName: keyword,
		TotalHits: totalHits,
		Domains:   results,
	}, nil
}

func extractDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	host := strings.ToLower(u.Host)
	if host == "" {
		return ""
	}

	if strings.HasPrefix(host, "www.") {
		host = host[4:]
	}

	return host
}

// GetTopPromptsByMentions returns prompts ranked by keyword mentions
func (s *StatsService) GetTopPromptsByMentions(ctx context.Context, limit int) ([]*PromptMentionStats, error) {
	responses, err := s.db.ListResponses(ctx, shared.ResponseFilter{Limit: 10000})
	if err != nil {
		return nil, fmt.Errorf("failed to get responses: %w", err)
	}

	promptMentions := make(map[string]int)
	promptNames := make(map[string]string)

	for _, response := range responses {
		promptMentions[response.PromptID]++

		if _, exists := promptNames[response.PromptID]; !exists {
			if prompt, err := s.db.GetPrompt(ctx, response.PromptID); err == nil {
				promptNames[response.PromptID] = prompt.Template
			} else {
				promptNames[response.PromptID] = fmt.Sprintf("Unknown Prompt (%s)", response.PromptID[:8])
			}
		}
	}

	var results []*PromptMentionStats
	for promptID, count := range promptMentions {
		results = append(results, &PromptMentionStats{
			PromptID:   promptID,
			PromptName: promptNames[promptID],
			Mentions:   count,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Mentions > results[j].Mentions
	})

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
	responses, err := s.db.ListResponses(ctx, shared.ResponseFilter{Limit: 10000})
	if err != nil {
		return nil, fmt.Errorf("failed to get responses: %w", err)
	}

	llmMentions := make(map[string]int)
	llmNames := make(map[string]string)

	for _, response := range responses {
		llmMentions[response.LLMID]++

		if _, exists := llmNames[response.LLMID]; !exists {
			if llm, err := s.db.GetLLM(ctx, response.LLMID); err == nil {
				llmNames[response.LLMID] = fmt.Sprintf("%s (%s)", llm.Name, llm.Provider)
			} else {
				llmNames[response.LLMID] = fmt.Sprintf("Unknown LLM (%s)", response.LLMID[:8])
			}
		}
	}

	var results []*LLMMentionStats
	for llmID, count := range llmMentions {
		results = append(results, &LLMMentionStats{
			LLMID:    llmID,
			LLMName:  llmNames[llmID],
			Mentions: count,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Mentions > results[j].Mentions
	})

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

type DashboardStatsOptions struct {
	PromptIDs    []string
	TagFilter    bool
	KeywordLimit int
}

func (s *StatsService) GetDashboardStats(ctx context.Context, opts DashboardStatsOptions) (*models.StatsResponse, error) {
	if opts.TagFilter && len(opts.PromptIDs) == 0 {
		return emptyDashboardStats(), nil
	}

	if !opts.TagFilter {
		totalResponses, err := s.GetTotalResponses(ctx)
		if err != nil {
			return nil, err
		}

		totalPrompts, err := s.GetTotalPrompts(ctx)
		if err != nil {
			return nil, err
		}

		totalLLMs, err := s.GetTotalLLMs(ctx)
		if err != nil {
			return nil, err
		}

		totalSchedules, err := s.GetTotalSchedules(ctx)
		if err != nil {
			return nil, err
		}

		topKeywords, err := s.GetTopKeywords(ctx, opts.KeywordLimit, nil, nil, nil)
		if err != nil {
			return nil, err
		}

		promptStats, err := s.GetAllPromptStats(ctx)
		if err != nil {
			return nil, err
		}

		llmStats, err := s.GetAllLLMStats(ctx)
		if err != nil {
			return nil, err
		}

		endTime := time.Now()
		startTime := endTime.AddDate(0, 0, -30)
		responseTrends, err := s.GetResponseTrends(ctx, startTime, endTime)
		if err != nil {
			return nil, err
		}

		trendResponses, err := s.db.ListResponses(ctx, shared.ResponseFilter{
			Limit: 10000,
		})
		if err != nil {
			return nil, err
		}
		brandTrends := aggregateBrandTrendsFromResponses(trendResponses, topKeywords, 30, opts.KeywordLimit)

		return &models.StatsResponse{
			TotalResponses: totalResponses,
			TotalPrompts:   totalPrompts,
			TotalLLMs:      totalLLMs,
			TotalSchedules: totalSchedules,
			TopKeywords:    topKeywords,
			BrandTrends:    brandTrends,
			PromptStats:    promptStats,
			LLMStats:       llmStats,
			ResponseTrends: responseTrends,
			LastUpdated:    time.Now(),
		}, nil
	}

	filter := shared.ResponseFilter{PromptIDs: opts.PromptIDs}
	totalResponses, err := s.db.CountResponses(ctx, filter)
	if err != nil {
		return nil, err
	}

	responses, err := s.db.ListResponses(ctx, shared.ResponseFilter{PromptIDs: opts.PromptIDs, Limit: 10000})
	if err != nil {
		return nil, err
	}

	llmStats := aggregateLLMStatsFromResponses(responses)
	promptStats := aggregatePromptStatsFromResponses(responses)
	topKeywords := aggregateTopKeywordsFromResponses(responses, opts.KeywordLimit)

	return &models.StatsResponse{
		TotalResponses: totalResponses,
		TotalPrompts:   int64(len(opts.PromptIDs)),
		TotalLLMs:      int64(len(llmStats)),
		TotalSchedules: 0,
		TopKeywords:    topKeywords,
		BrandTrends:    aggregateBrandTrendsFromResponses(responses, topKeywords, 30, opts.KeywordLimit),
		PromptStats:    promptStats,
		LLMStats:       llmStats,
		ResponseTrends: []models.TimeSeriesPoint{},
		LastUpdated:    time.Now(),
	}, nil
}

func emptyDashboardStats() *models.StatsResponse {
	return &models.StatsResponse{
		TopKeywords:    []models.KeywordCount{},
		BrandTrends:    []models.BrandTrendSeries{},
		PromptStats:    []*models.PromptStats{},
		LLMStats:       []*models.LLMStats{},
		ResponseTrends: []models.TimeSeriesPoint{},
		LastUpdated:    time.Now(),
	}
}

func aggregateTopKeywordsFromResponses(responses []*models.Response, limit int) []models.KeywordCount {
	wordCounts := make(map[string]int)
	for _, response := range responses {
		for _, word := range shared.ExtractCapitalizedWords(response.ResponseText) {
			wordCounts[word]++
		}
	}

	type kv struct {
		keyword string
		count   int
	}

	sorted := make([]kv, 0, len(wordCounts))
	for keyword, count := range wordCounts {
		sorted = append(sorted, kv{keyword: keyword, count: count})
	}

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].count == sorted[j].count {
			return sorted[i].keyword < sorted[j].keyword
		}
		return sorted[i].count > sorted[j].count
	})

	if limit <= 0 {
		limit = 10
	}
	if len(sorted) > limit {
		sorted = sorted[:limit]
	}

	results := make([]models.KeywordCount, 0, len(sorted))
	for _, item := range sorted {
		results = append(results, models.KeywordCount{
			Keyword: item.keyword,
			Count:   item.count,
		})
	}

	return results
}

const executionBatchGap = 30 * time.Minute

type executionEvent struct {
	runID     string
	timestamp time.Time
	responses []*models.Response
}

func groupResponsesByExecution(responses []*models.Response) []executionEvent {
	if len(responses) == 0 {
		return nil
	}

	sorted := append([]*models.Response(nil), responses...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
	})

	runIndex := make(map[string]int)
	events := make([]executionEvent, 0, len(sorted))

	for _, response := range sorted {
		if response.RunID != "" {
			if idx, ok := runIndex[response.RunID]; ok {
				events[idx].responses = append(events[idx].responses, response)
				if response.CreatedAt.Before(events[idx].timestamp) {
					events[idx].timestamp = response.CreatedAt
				}
				continue
			}

			runIndex[response.RunID] = len(events)
			events = append(events, executionEvent{
				runID:     response.RunID,
				timestamp: response.CreatedAt,
				responses: []*models.Response{response},
			})
			continue
		}

		if len(events) > 0 {
			last := &events[len(events)-1]
			if last.runID == "" && response.CreatedAt.Sub(latestResponseTime(last.responses)) <= executionBatchGap {
				last.responses = append(last.responses, response)
				continue
			}
		}

		events = append(events, executionEvent{
			timestamp: response.CreatedAt,
			responses: []*models.Response{response},
		})
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].timestamp.Before(events[j].timestamp)
	})

	return events
}

func latestResponseTime(responses []*models.Response) time.Time {
	latest := responses[0].CreatedAt
	for _, response := range responses[1:] {
		if response.CreatedAt.After(latest) {
			latest = response.CreatedAt
		}
	}
	return latest
}

func aggregateBrandTrendsFromResponses(
	responses []*models.Response,
	topKeywords []models.KeywordCount,
	maxExecutions int,
	seriesLimit int,
) []models.BrandTrendSeries {
	if maxExecutions <= 0 {
		maxExecutions = 30
	}
	if seriesLimit <= 0 {
		seriesLimit = 10
	}
	if len(topKeywords) == 0 {
		return []models.BrandTrendSeries{}
	}

	if seriesLimit > len(topKeywords) {
		seriesLimit = len(topKeywords)
	}

	tracked := make(map[string]bool, seriesLimit)
	keywords := topKeywords[:seriesLimit]
	for _, keyword := range keywords {
		tracked[keyword.Keyword] = true
	}

	events := groupResponsesByExecution(responses)
	if len(events) > maxExecutions {
		events = events[len(events)-maxExecutions:]
	}
	if len(events) == 0 {
		return []models.BrandTrendSeries{}
	}

	buckets := make(map[string][]int, seriesLimit)
	for _, keyword := range keywords {
		buckets[keyword.Keyword] = make([]int, len(events))
	}

	for eventIdx, event := range events {
		for _, response := range event.responses {
			for _, word := range shared.ExtractCapitalizedWords(response.ResponseText) {
				if tracked[word] {
					buckets[word][eventIdx]++
				}
			}
		}
	}

	result := make([]models.BrandTrendSeries, 0, seriesLimit)
	for _, keyword := range keywords {
		points := make([]models.TimeSeriesPoint, 0, len(events))
		for eventIdx, event := range events {
			points = append(points, models.TimeSeriesPoint{
				Timestamp: event.timestamp,
				Count:     buckets[keyword.Keyword][eventIdx],
			})
		}

		result = append(result, models.BrandTrendSeries{
			Keyword: keyword.Keyword,
			Points:  points,
		})
	}

	return result
}

func aggregateLLMStatsFromResponses(responses []*models.Response) []*models.LLMStats {
	type accumulator struct {
		totalResponses int
		totalTokens    int
		promptCounts   map[string]int
	}

	statsByLLM := make(map[string]*accumulator)
	for _, response := range responses {
		if statsByLLM[response.LLMID] == nil {
			statsByLLM[response.LLMID] = &accumulator{promptCounts: make(map[string]int)}
		}

		stats := statsByLLM[response.LLMID]
		stats.totalResponses++
		stats.totalTokens += response.TokensUsed
		stats.promptCounts[response.PromptID]++
	}

	results := make([]*models.LLMStats, 0, len(statsByLLM))
	now := time.Now()
	for llmID, stats := range statsByLLM {
		avgTokens := 0.0
		if stats.totalResponses > 0 {
			avgTokens = float64(stats.totalTokens) / float64(stats.totalResponses)
		}

		results = append(results, &models.LLMStats{
			LLMID:          llmID,
			TotalResponses: stats.totalResponses,
			UniquePrompts:  len(stats.promptCounts),
			PromptCounts:   stats.promptCounts,
			AvgTokens:      avgTokens,
			UpdatedAt:      now,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].TotalResponses == results[j].TotalResponses {
			return results[i].LLMID < results[j].LLMID
		}
		return results[i].TotalResponses > results[j].TotalResponses
	})

	return results
}

func aggregatePromptStatsFromResponses(responses []*models.Response) []*models.PromptStats {
	type accumulator struct {
		totalResponses int
		totalTokens    int
		llmCounts      map[string]int
	}

	statsByPrompt := make(map[string]*accumulator)
	for _, response := range responses {
		if statsByPrompt[response.PromptID] == nil {
			statsByPrompt[response.PromptID] = &accumulator{llmCounts: make(map[string]int)}
		}

		stats := statsByPrompt[response.PromptID]
		stats.totalResponses++
		stats.totalTokens += response.TokensUsed
		stats.llmCounts[response.LLMID]++
	}

	results := make([]*models.PromptStats, 0, len(statsByPrompt))
	now := time.Now()
	for promptID, stats := range statsByPrompt {
		avgTokens := 0.0
		if stats.totalResponses > 0 {
			avgTokens = float64(stats.totalTokens) / float64(stats.totalResponses)
		}

		results = append(results, &models.PromptStats{
			PromptID:       promptID,
			TotalResponses: stats.totalResponses,
			UniqueLLMs:     len(stats.llmCounts),
			LLMCounts:      stats.llmCounts,
			AvgTokens:      avgTokens,
			UpdatedAt:      now,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].TotalResponses == results[j].TotalResponses {
			return results[i].PromptID < results[j].PromptID
		}
		return results[i].TotalResponses > results[j].TotalResponses
	})

	return results
}
