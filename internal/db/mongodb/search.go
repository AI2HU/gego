package mongodb

import (
	"context"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
)

// SearchKeyword searches for a keyword in all responses and calculates stats on-the-fly
func (m *MongoDB) SearchKeyword(ctx context.Context, keyword string, startTime, endTime *time.Time) (*models.KeywordStats, error) {
	// Build regex pattern for case-insensitive search
	pattern := regexp.QuoteMeta(keyword)
	regex := bson.M{"$regex": pattern, "$options": "i"}

	// Build query
	query := bson.M{
		"response_text": regex,
	}

	if startTime != nil || endTime != nil {
		timeQuery := bson.M{}
		if startTime != nil {
			timeQuery["$gte"] = *startTime
		}
		if endTime != nil {
			timeQuery["$lte"] = *endTime
		}
		query["created_at"] = timeQuery
	}

	// Find all matching responses
	cursor, err := m.database.Collection(collResponses).Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Calculate stats
	stats := &models.KeywordStats{
		Keyword:    keyword,
		ByPrompt:   make(map[string]int),
		ByLLM:      make(map[string]int),
		ByProvider: make(map[string]int),
	}

	promptsSeen := make(map[string]bool)
	llmsSeen := make(map[string]bool)

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		// Extract fields manually
		responseText := getString(doc, "response_text")
		promptID := getString(doc, "prompt_id")
		llmID := getString(doc, "llm_id")
		llmProvider := getString(doc, "llm_provider")
		createdAt := getTime(doc, "created_at")

		// Count occurrences in this response (case-insensitive)
		count := countOccurrences(responseText, keyword)
		stats.TotalMentions += count

		// Track by prompt
		stats.ByPrompt[promptID] += count
		promptsSeen[promptID] = true

		// Track by LLM
		stats.ByLLM[llmID] += count
		llmsSeen[llmID] = true

		// Track by provider
		stats.ByProvider[llmProvider] += count

		// Track first/last seen
		if stats.FirstSeen.IsZero() || createdAt.Before(stats.FirstSeen) {
			stats.FirstSeen = createdAt
		}
		if stats.LastSeen.IsZero() || createdAt.After(stats.LastSeen) {
			stats.LastSeen = createdAt
		}
	}

	stats.UniquePrompts = len(promptsSeen)
	stats.UniqueLLMs = len(llmsSeen)

	return stats, nil
}

// GetTopKeywords returns the most common keywords across all responses
func (m *MongoDB) GetTopKeywords(ctx context.Context, limit int, startTime, endTime *time.Time) ([]db.KeywordCount, error) {
	// Build query
	query := bson.M{}
	if startTime != nil || endTime != nil {
		timeQuery := bson.M{}
		if startTime != nil {
			timeQuery["$gte"] = *startTime
		}
		if endTime != nil {
			timeQuery["$lte"] = *endTime
		}
		query["created_at"] = timeQuery
	}

	// Find all responses
	cursor, err := m.database.Collection(collResponses).Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Extract all words and count them
	wordCounts := make(map[string]int)
	for cursor.Next(ctx) {
		var response models.Response
		if err := cursor.Decode(&response); err != nil {
			continue
		}

		// Extract capitalized words (likely to be keywords)
		words := extractCapitalizedWords(response.ResponseText)
		for _, word := range words {
			wordCounts[word]++
		}
	}

	// Sort by count and return top N
	type kv struct {
		keyword string
		count   int
	}

	var sorted []kv
	for k, v := range wordCounts {
		sorted = append(sorted, kv{k, v})
	}

	// Sort descending by count
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].count > sorted[i].count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Build result
	var results []db.KeywordCount
	for i := 0; i < len(sorted) && i < limit; i++ {
		results = append(results, db.KeywordCount{
			Keyword: sorted[i].keyword,
			Count:   sorted[i].count,
		})
	}

	return results, nil
}

// countOccurrences counts how many times a keyword appears in text (case-insensitive)
func countOccurrences(text, keyword string) int {
	lower := strings.ToLower(text)
	lowerKeyword := strings.ToLower(keyword)
	count := 0
	index := 0

	for {
		i := strings.Index(lower[index:], lowerKeyword)
		if i == -1 {
			break
		}
		count++
		index += i + len(lowerKeyword)
	}

	return count
}

// extractCapitalizedWords extracts words that start with a capital letter
func extractCapitalizedWords(text string) []string {
	re := regexp.MustCompile(`\b[A-Z][a-zA-Z]+(?:\s+[A-Z][a-zA-Z]+)*\b`)
	matches := re.FindAllString(text, -1)

	// Filter common words
	var filtered []string
	commonWords := map[string]bool{
		"The": true, "A": true, "An": true, "And": true, "Or": true,
		"But": true, "In": true, "On": true, "At": true, "To": true,
		"For": true, "Of": true, "With": true, "By": true, "From": true,
		"This": true, "That": true, "These": true, "Those": true,
		"I": true, "You": true, "He": true, "She": true, "It": true,
		"We": true, "They": true, "My": true, "Your": true, "His": true,
		"Her": true, "Its": true, "Our": true, "Their": true,
	}

	for _, word := range matches {
		if !commonWords[word] {
			filtered = append(filtered, word)
		}
	}

	return filtered
}
