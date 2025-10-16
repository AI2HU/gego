package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

// MongoDB implements the Database interface for MongoDB
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	config   *models.Config
}

const (
	collPrompts   = "prompts"
	collResponses = "responses"
)

// New creates a new MongoDB database instance
func New(config *models.Config) (*MongoDB, error) {
	return &MongoDB{
		config: config,
	}, nil
}

// Connect establishes connection to MongoDB
func (m *MongoDB) Connect(ctx context.Context) error {
	clientOptions := options.Client().ApplyURI(m.config.URI)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.client = client
	m.database = client.Database(m.config.Database)

	if err := m.createIndexes(ctx); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// Disconnect closes the MongoDB connection
func (m *MongoDB) Disconnect(ctx context.Context) error {
	if m.client != nil {
		return m.client.Disconnect(ctx)
	}
	return nil
}

// Ping checks the database connection
func (m *MongoDB) Ping(ctx context.Context) error {
	if m.client == nil {
		return fmt.Errorf("not connected to database")
	}
	return m.client.Ping(ctx, nil)
}

// createIndexes creates necessary indexes for optimal query performance
func (m *MongoDB) createIndexes(ctx context.Context) error {
	responseIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "prompt_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "created_at", Value: -1},
			},
		},
	}

	_, err := m.database.Collection(collResponses).Indexes().CreateMany(ctx, responseIndexes)
	if err != nil {
		return fmt.Errorf("failed to create response indexes: %w", err)
	}

	return nil
}

// CreatePrompt creates a new prompt
func (m *MongoDB) CreatePrompt(ctx context.Context, prompt *models.Prompt) error {
	prompt.CreatedAt = time.Now()
	prompt.UpdatedAt = time.Now()

	doc := bson.M{
		"_id":        prompt.ID,
		"template":   prompt.Template,
		"tags":       prompt.Tags,
		"enabled":    prompt.Enabled,
		"created_at": prompt.CreatedAt,
		"updated_at": prompt.UpdatedAt,
	}

	_, err := m.database.Collection(collPrompts).InsertOne(ctx, doc)
	return err
}

// GetPrompt retrieves a prompt by ID
func (m *MongoDB) GetPrompt(ctx context.Context, id string) (*models.Prompt, error) {
	var doc bson.M
	err := m.database.Collection(collPrompts).FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("prompt not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	var promptID string
	if id, ok := doc["_id"].(string); ok {
		promptID = id
	} else if objectID, ok := doc["_id"].(primitive.ObjectID); ok {
		promptID = objectID.Hex()
	} else {
		return nil, fmt.Errorf("invalid _id type in document")
	}

	prompt := &models.Prompt{
		ID:        promptID,
		Template:  getString(doc, "template"),
		Enabled:   getBool(doc, "enabled"),
		CreatedAt: getTime(doc, "created_at"),
		UpdatedAt: getTime(doc, "updated_at"),
	}

	if tags, ok := doc["tags"].([]interface{}); ok {
		for _, t := range tags {
			if str, ok := t.(string); ok {
				prompt.Tags = append(prompt.Tags, str)
			}
		}
	}

	return prompt, nil
}

// ListPrompts lists all prompts, optionally filtered by enabled status
func (m *MongoDB) ListPrompts(ctx context.Context, enabled *bool) ([]*models.Prompt, error) {
	filter := bson.M{}
	if enabled != nil {
		filter["enabled"] = *enabled
	}

	cursor, err := m.database.Collection(collPrompts).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var prompts []*models.Prompt
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		// Convert BSON document to Prompt struct
		var promptID string
		if id, ok := doc["_id"].(string); ok {
			promptID = id
		} else if objectID, ok := doc["_id"].(primitive.ObjectID); ok {
			promptID = objectID.Hex()
		} else {
			return nil, fmt.Errorf("invalid _id type in document")
		}

		prompt := &models.Prompt{
			ID:        promptID,
			Template:  getString(doc, "template"),
			Enabled:   getBool(doc, "enabled"),
			CreatedAt: getTime(doc, "created_at"),
			UpdatedAt: getTime(doc, "updated_at"),
		}

		// Handle optional fields
		if tags, ok := doc["tags"].([]interface{}); ok {
			for _, t := range tags {
				if str, ok := t.(string); ok {
					prompt.Tags = append(prompt.Tags, str)
				}
			}
		}

		prompts = append(prompts, prompt)
	}

	return prompts, nil
}

// UpdatePrompt updates an existing prompt
func (m *MongoDB) UpdatePrompt(ctx context.Context, prompt *models.Prompt) error {
	prompt.UpdatedAt = time.Now()

	// Convert to BSON document with explicit _id field
	doc := bson.M{
		"_id":        prompt.ID,
		"template":   prompt.Template,
		"tags":       prompt.Tags,
		"enabled":    prompt.Enabled,
		"created_at": prompt.CreatedAt,
		"updated_at": prompt.UpdatedAt,
	}

	result, err := m.database.Collection(collPrompts).ReplaceOne(
		ctx,
		bson.M{"_id": prompt.ID},
		doc,
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("prompt not found: %s", prompt.ID)
	}

	return nil
}

// DeletePrompt deletes a prompt by ID
func (m *MongoDB) DeletePrompt(ctx context.Context, id string) error {
	var filter bson.M
	if objectID, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": id}
	}

	result, err := m.database.Collection(collPrompts).DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("prompt not found: %s", id)
	}

	return nil
}

// DeleteAllPrompts deletes all prompts
func (m *MongoDB) DeleteAllPrompts(ctx context.Context) (int, error) {
	result, err := m.database.Collection(collPrompts).DeleteMany(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	return int(result.DeletedCount), nil
}

// CreateResponse creates a new response
func (m *MongoDB) CreateResponse(ctx context.Context, response *models.Response) error {
	response.CreatedAt = time.Now()

	doc := bson.M{
		"_id":           response.ID,
		"prompt_id":     response.PromptID,
		"prompt_text":   response.PromptText,
		"llm_id":        response.LLMID,
		"llm_name":      response.LLMName,
		"llm_provider":  response.LLMProvider,
		"llm_model":     response.LLMModel,
		"response_text": response.ResponseText,
		"schedule_id":   response.ScheduleID,
		"tokens_used":   response.TokensUsed,
		"temperature":   response.Temperature,
		"created_at":    response.CreatedAt,
	}

	if response.Metadata != nil {
		doc["metadata"] = response.Metadata
	}

	_, err := m.database.Collection(collResponses).InsertOne(ctx, doc)
	return err
}

// GetResponse retrieves a response by ID
func (m *MongoDB) GetResponse(ctx context.Context, id string) (*models.Response, error) {
	var response models.Response
	err := m.database.Collection(collResponses).FindOne(ctx, bson.M{"_id": id}).Decode(&response)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("response not found: %s", id)
	}
	return &response, err
}

// ListResponses lists responses with filtering
func (m *MongoDB) ListResponses(ctx context.Context, filter shared.ResponseFilter) ([]*models.Response, error) {
	query := bson.M{}

	if filter.PromptID != "" {
		query["prompt_id"] = filter.PromptID
	}
	if filter.LLMID != "" {
		query["llm_id"] = filter.LLMID
	}
	if filter.ScheduleID != "" {
		query["schedule_id"] = filter.ScheduleID
	}
	if filter.Keyword != "" {
		query["response_text"] = bson.M{
			"$regex":   filter.Keyword,
			"$options": "i",
		}
	}
	if filter.StartTime != nil || filter.EndTime != nil {
		timeQuery := bson.M{}
		if filter.StartTime != nil {
			timeQuery["$gte"] = *filter.StartTime
		}
		if filter.EndTime != nil {
			timeQuery["$lte"] = *filter.EndTime
		}
		query["created_at"] = timeQuery
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	if filter.Limit > 0 {
		opts.SetLimit(int64(filter.Limit))
	}
	if filter.Offset > 0 {
		opts.SetSkip(int64(filter.Offset))
	}

	cursor, err := m.database.Collection(collResponses).Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var responses []*models.Response
	if err := cursor.All(ctx, &responses); err != nil {
		return nil, err
	}

	return responses, nil
}

// GetDatabase returns the underlying MongoDB database instance
func (m *MongoDB) GetDatabase() *mongo.Database {
	return m.database
}

// Helper functions for safe field extraction
func getString(doc bson.M, key string) string {
	if val, ok := doc[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBool(doc bson.M, key string) bool {
	if val, ok := doc[key]; ok && val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func getTime(doc bson.M, key string) time.Time {
	if val, ok := doc[key]; ok && val != nil {
		if t, ok := val.(time.Time); ok {
			return t
		}
		if dt, ok := val.(primitive.DateTime); ok {
			return dt.Time()
		}
		if ts, ok := val.(int64); ok {
			return time.Unix(ts, 0)
		}
		if ts, ok := val.(float64); ok {
			return time.Unix(int64(ts), 0)
		}
	}
	return time.Time{}
}

// DeleteAllResponses deletes all responses from the database
func (m *MongoDB) DeleteAllResponses(ctx context.Context) (int, error) {
	result, err := m.database.Collection(collResponses).DeleteMany(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	return int(result.DeletedCount), nil
}

// GetPromptStats calculates prompt statistics on-demand from responses
func (m *MongoDB) GetPromptStats(ctx context.Context, promptID string) (*models.PromptStats, error) {
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

	cursor, err := m.database.Collection(collResponses).Aggregate(ctx, pipeline)
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

	llmCounts, err := m.getLLMCountsForPrompt(ctx, promptID)
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
func (m *MongoDB) GetLLMStats(ctx context.Context, llmID string) (*models.LLMStats, error) {
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

	cursor, err := m.database.Collection(collResponses).Aggregate(ctx, pipeline)
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

	promptCounts, err := m.getPromptCountsForLLM(ctx, llmID)
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
func (m *MongoDB) getLLMCountsForPrompt(ctx context.Context, promptID string) (map[string]int, error) {
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

	cursor, err := m.database.Collection(collResponses).Aggregate(ctx, pipeline)
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
func (m *MongoDB) getPromptCountsForLLM(ctx context.Context, llmID string) (map[string]int, error) {
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

	cursor, err := m.database.Collection(collResponses).Aggregate(ctx, pipeline)
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
