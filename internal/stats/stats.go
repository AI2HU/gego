package stats

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/AI2HU/gego/internal/models"
)

// Service provides statistics calculations by consuming data directly from MongoDB
type Service struct {
	database *mongo.Database
}

// New creates a new stats service
func New(database *mongo.Database) *Service {
	return &Service{
		database: database,
	}
}

// GetPromptStats calculates prompt statistics on-demand from responses
func (s *Service) GetPromptStats(ctx context.Context, promptID string) (*models.PromptStats, error) {
	// Aggregate responses for this prompt
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"prompt_id": promptID,
			},
		},
		{
			"$group": bson.M{
				"_id":             nil,
				"total_responses": bson.M{"$sum": 1},
				"avg_tokens": bson.M{
					"$avg": "$tokens_used",
				},
				"unique_llms": bson.M{"$addToSet": "$llm_id"},
			},
		},
		{
			"$project": bson.M{
				"total_responses": 1,
				"avg_tokens":      1,
				"unique_llms":     bson.M{"$size": "$unique_llms"},
			},
		},
	}

	cursor, err := s.database.Collection("responses").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate prompt stats: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct {
		TotalResponses int     `bson:"total_responses"`
		AvgTokens      float64 `bson:"avg_tokens"`
		UniqueLLMs     int     `bson:"unique_llms"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode prompt stats: %w", err)
		}
	}

	// Get LLM counts for this prompt
	llmCounts, err := s.getLLMCountsForPrompt(ctx, promptID)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM counts: %w", err)
	}

	return &models.PromptStats{
		PromptID:       promptID,
		TotalResponses: result.TotalResponses,
		UniqueLLMs:     result.UniqueLLMs,
		LLMCounts:      llmCounts,
		AvgTokens:      result.AvgTokens,
		UpdatedAt:      time.Now(),
	}, nil
}

// GetLLMStats calculates LLM statistics on-demand from responses
func (s *Service) GetLLMStats(ctx context.Context, llmID string) (*models.LLMStats, error) {
	// Aggregate responses for this LLM
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"llm_id": llmID,
			},
		},
		{
			"$group": bson.M{
				"_id":             nil,
				"total_responses": bson.M{"$sum": 1},
				"avg_tokens": bson.M{
					"$avg": "$tokens_used",
				},
				"unique_prompts": bson.M{"$addToSet": "$prompt_id"},
			},
		},
		{
			"$project": bson.M{
				"total_responses": 1,
				"avg_tokens":      1,
				"unique_prompts":  bson.M{"$size": "$unique_prompts"},
			},
		},
	}

	cursor, err := s.database.Collection("responses").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate LLM stats: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct {
		TotalResponses int     `bson:"total_responses"`
		AvgTokens      float64 `bson:"avg_tokens"`
		UniquePrompts  int     `bson:"unique_prompts"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode LLM stats: %w", err)
		}
	}

	// Get prompt counts for this LLM
	promptCounts, err := s.getPromptCountsForLLM(ctx, llmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt counts: %w", err)
	}

	return &models.LLMStats{
		LLMID:          llmID,
		TotalResponses: result.TotalResponses,
		UniquePrompts:  result.UniquePrompts,
		PromptCounts:   promptCounts,
		AvgTokens:      result.AvgTokens,
		UpdatedAt:      time.Now(),
	}, nil
}

// getLLMCountsForPrompt gets the count of responses by LLM for a specific prompt
func (s *Service) getLLMCountsForPrompt(ctx context.Context, promptID string) (map[string]int, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"prompt_id": promptID,
			},
		},
		{
			"$group": bson.M{
				"_id":   "$llm_id",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := s.database.Collection("responses").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	counts := make(map[string]int)
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		counts[result.ID] = result.Count
	}

	return counts, nil
}

// getPromptCountsForLLM gets the count of responses by prompt for a specific LLM
func (s *Service) getPromptCountsForLLM(ctx context.Context, llmID string) (map[string]int, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"llm_id": llmID,
			},
		},
		{
			"$group": bson.M{
				"_id":   "$prompt_id",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := s.database.Collection("responses").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	counts := make(map[string]int)
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		counts[result.ID] = result.Count
	}

	return counts, nil
}
