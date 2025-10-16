package mongodb

import (
	"context"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

// SearchKeyword searches for a keyword in all responses and calculates stats on-the-fly
func (m *MongoDB) SearchKeyword(ctx context.Context, keyword string, startTime, endTime *time.Time) (*models.KeywordStats, error) {
	pattern := regexp.QuoteMeta(keyword)
	regex := bson.M{"$regex": pattern, "$options": "i"}

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

	cursor, err := m.database.Collection(collResponses).Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

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

		responseText := getString(doc, "response_text")
		promptID := getString(doc, "prompt_id")
		llmID := getString(doc, "llm_id")
		llmProvider := getString(doc, "llm_provider")
		createdAt := getTime(doc, "created_at")

		count := shared.CountOccurrences(responseText, keyword)
		stats.TotalMentions += count

		stats.ByPrompt[promptID] += count
		promptsSeen[promptID] = true

		stats.ByLLM[llmID] += count
		llmsSeen[llmID] = true

		stats.ByProvider[llmProvider] += count

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
func (m *MongoDB) GetTopKeywords(ctx context.Context, limit int, startTime, endTime *time.Time) ([]models.KeywordCount, error) {
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

	cursor, err := m.database.Collection(collResponses).Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	wordCounts := make(map[string]int)
	for cursor.Next(ctx) {
		var response models.Response
		if err := cursor.Decode(&response); err != nil {
			continue
		}

		words := shared.ExtractCapitalizedWords(response.ResponseText)
		for _, word := range words {
			wordCounts[word]++
		}
	}

	type kv struct {
		keyword string
		count   int
	}

	var sorted []kv
	for k, v := range wordCounts {
		sorted = append(sorted, kv{k, v})
	}

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].count > sorted[i].count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	var results []models.KeywordCount
	for i := 0; i < len(sorted) && i < limit; i++ {
		results = append(results, models.KeywordCount{
			Keyword: sorted[i].keyword,
			Count:   sorted[i].count,
		})
	}

	return results, nil
}
