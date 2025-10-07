package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
)

// MongoDB implements the Database interface for MongoDB
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	config   *db.Config
}

const (
	collLLMs      = "llms"
	collPrompts   = "prompts"
	collSchedules = "schedules"
	collResponses = "responses"
)

// New creates a new MongoDB database instance
func New(config *db.Config) (*MongoDB, error) {
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

	// Create indexes
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
	// Responses indexes - critical for performance
	responseIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "prompt_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "llm_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "schedule_id", Value: 1},
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

// CreateLLM creates a new LLM configuration
func (m *MongoDB) CreateLLM(ctx context.Context, llm *models.LLMConfig) error {
	llm.CreatedAt = time.Now()
	llm.UpdatedAt = time.Now()

	doc := bson.M{
		"_id":        llm.ID,
		"name":       llm.Name,
		"provider":   llm.Provider,
		"model":      llm.Model,
		"api_key":    llm.APIKey,
		"base_url":   llm.BaseURL,
		"enabled":    llm.Enabled,
		"created_at": llm.CreatedAt,
		"updated_at": llm.UpdatedAt,
	}

	if llm.Config != nil {
		doc["config"] = llm.Config
	}

	_, err := m.database.Collection(collLLMs).InsertOne(ctx, doc)
	return err
}

// GetLLM retrieves an LLM by ID
func (m *MongoDB) GetLLM(ctx context.Context, id string) (*models.LLMConfig, error) {
	var filter bson.M
	if objectID, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": id}
	}

	var doc bson.M
	err := m.database.Collection(collLLMs).FindOne(ctx, filter).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("LLM not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	var llmID string
	if docID, ok := doc["_id"].(string); ok {
		llmID = docID
	} else if objectID, ok := doc["_id"].(primitive.ObjectID); ok {
		llmID = objectID.Hex()
	} else {
		return nil, fmt.Errorf("invalid _id type in LLM document")
	}

	llm := &models.LLMConfig{
		ID:        llmID,
		Name:      getString(doc, "name"),
		Provider:  getString(doc, "provider"),
		Model:     getString(doc, "model"),
		APIKey:    getStringFromEither(doc, "api_key", "APIKey"),
		BaseURL:   getStringFromEither(doc, "base_url", "BaseURL"),
		Enabled:   getBool(doc, "enabled"),
		CreatedAt: getTime(doc, "created_at"),
		UpdatedAt: getTime(doc, "updated_at"),
	}

	return llm, nil
}

// ListLLMs lists all LLMs, optionally filtered by enabled status
func (m *MongoDB) ListLLMs(ctx context.Context, enabled *bool) ([]*models.LLMConfig, error) {
	filter := bson.M{}
	if enabled != nil {
		filter["enabled"] = *enabled
	}

	cursor, err := m.database.Collection(collLLMs).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var llms []*models.LLMConfig
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		var llmID string
		if id, ok := doc["_id"].(string); ok {
			llmID = id
		} else if objectID, ok := doc["_id"].(primitive.ObjectID); ok {
			llmID = objectID.Hex()
		} else {
			return nil, fmt.Errorf("invalid _id type in LLM document")
		}

		llm := &models.LLMConfig{
			ID:        llmID,
			Name:      getString(doc, "name"),
			Provider:  getString(doc, "provider"),
			Model:     getString(doc, "model"),
			APIKey:    getStringFromEither(doc, "api_key", "APIKey"),
			BaseURL:   getStringFromEither(doc, "base_url", "BaseURL"),
			Enabled:   getBool(doc, "enabled"),
			CreatedAt: getTime(doc, "created_at"),
			UpdatedAt: getTime(doc, "updated_at"),
		}

		llms = append(llms, llm)
	}

	return llms, nil
}

// UpdateLLM updates an existing LLM configuration
func (m *MongoDB) UpdateLLM(ctx context.Context, llm *models.LLMConfig) error {
	llm.UpdatedAt = time.Now()

	// Convert to BSON document with explicit _id field
	doc := bson.M{
		"name":       llm.Name,
		"provider":   llm.Provider,
		"model":      llm.Model,
		"api_key":    llm.APIKey,
		"base_url":   llm.BaseURL,
		"enabled":    llm.Enabled,
		"created_at": llm.CreatedAt,
		"updated_at": llm.UpdatedAt,
		// Clear old field names to prevent confusion
		"APIKey":  "",
		"BaseURL": "",
	}

	// Add config if it exists
	if llm.Config != nil {
		doc["config"] = llm.Config
	}

	// Try to parse as ObjectID first (for old documents), then as string
	var filter bson.M
	if objectID, err := primitive.ObjectIDFromHex(llm.ID); err == nil {
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": llm.ID}
	}

	result, err := m.database.Collection(collLLMs).ReplaceOne(ctx, filter, doc)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("LLM not found: %s", llm.ID)
	}

	return nil
}

// DeleteLLM deletes an LLM by ID
func (m *MongoDB) DeleteLLM(ctx context.Context, id string) error {
	result, err := m.database.Collection(collLLMs).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("LLM not found: %s", id)
	}

	return nil
}

// DeleteAllLLMs deletes all LLMs
func (m *MongoDB) DeleteAllLLMs(ctx context.Context) (int, error) {
	result, err := m.database.Collection(collLLMs).DeleteMany(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	return int(result.DeletedCount), nil
}

// CreatePrompt creates a new prompt
func (m *MongoDB) CreatePrompt(ctx context.Context, prompt *models.Prompt) error {
	prompt.CreatedAt = time.Now()
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
	// Try to parse as ObjectID first (for old documents), then as string
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

// CreateSchedule creates a new schedule
func (m *MongoDB) CreateSchedule(ctx context.Context, schedule *models.Schedule) error {
	schedule.CreatedAt = time.Now()
	schedule.UpdatedAt = time.Now()

	// Convert to BSON document with explicit _id field
	doc := bson.M{
		"_id":        schedule.ID,
		"name":       schedule.Name,
		"cron_expr":  schedule.CronExpr,
		"enabled":    schedule.Enabled,
		"prompt_ids": schedule.PromptIDs,
		"llm_ids":    schedule.LLMIDs,
		"created_at": schedule.CreatedAt,
		"updated_at": schedule.UpdatedAt,
	}

	// Add optional fields if they exist
	if schedule.LastRun != nil {
		doc["last_run"] = *schedule.LastRun
	}
	if schedule.NextRun != nil {
		doc["next_run"] = *schedule.NextRun
	}

	_, err := m.database.Collection(collSchedules).InsertOne(ctx, doc)
	return err
}

// GetSchedule retrieves a schedule by ID
func (m *MongoDB) GetSchedule(ctx context.Context, id string) (*models.Schedule, error) {
	// Try to parse as ObjectID first (for old documents), then as string
	var filter bson.M
	if objectID, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": id}
	}

	var doc bson.M
	err := m.database.Collection(collSchedules).FindOne(ctx, filter).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("schedule not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	// Convert BSON document to Schedule struct
	var scheduleID string
	if docID, ok := doc["_id"].(string); ok {
		scheduleID = docID
	} else if objectID, ok := doc["_id"].(primitive.ObjectID); ok {
		scheduleID = objectID.Hex()
	} else {
		return nil, fmt.Errorf("invalid _id type in schedule document")
	}

	schedule := &models.Schedule{
		ID:        scheduleID,
		Name:      getString(doc, "name"),
		CronExpr:  getString(doc, "cron_expr"),
		Enabled:   getBool(doc, "enabled"),
		CreatedAt: getTime(doc, "created_at"),
		UpdatedAt: getTime(doc, "updated_at"),
	}

	// Handle optional fields
	if lastRun, ok := doc["last_run"].(time.Time); ok {
		schedule.LastRun = &lastRun
	}
	if nextRun, ok := doc["next_run"].(time.Time); ok {
		schedule.NextRun = &nextRun
	}

	// Handle arrays
	if promptIDs, ok := doc["prompt_ids"].([]interface{}); ok {
		for _, id := range promptIDs {
			if str, ok := id.(string); ok {
				schedule.PromptIDs = append(schedule.PromptIDs, str)
			}
		}
	} else if promptIDs, ok := doc["prompt_ids"].(primitive.A); ok {
		for _, id := range promptIDs {
			if str, ok := id.(string); ok {
				schedule.PromptIDs = append(schedule.PromptIDs, str)
			}
		}
	}

	if llmIDs, ok := doc["llm_ids"].([]interface{}); ok {
		for _, id := range llmIDs {
			if str, ok := id.(string); ok {
				schedule.LLMIDs = append(schedule.LLMIDs, str)
			}
		}
	} else if llmIDs, ok := doc["llm_ids"].(primitive.A); ok {
		for _, id := range llmIDs {
			if str, ok := id.(string); ok {
				schedule.LLMIDs = append(schedule.LLMIDs, str)
			}
		}
	}

	return schedule, nil
}

// ListSchedules lists all schedules, optionally filtered by enabled status
func (m *MongoDB) ListSchedules(ctx context.Context, enabled *bool) ([]*models.Schedule, error) {
	filter := bson.M{}
	if enabled != nil {
		filter["enabled"] = *enabled
	}

	cursor, err := m.database.Collection(collSchedules).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var schedules []*models.Schedule
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		// Convert BSON document to Schedule struct
		var scheduleID string
		if id, ok := doc["_id"].(string); ok {
			scheduleID = id
		} else if objectID, ok := doc["_id"].(primitive.ObjectID); ok {
			scheduleID = objectID.Hex()
		} else {
			return nil, fmt.Errorf("invalid _id type in schedule document")
		}

		schedule := &models.Schedule{
			ID:        scheduleID,
			Name:      getString(doc, "name"),
			CronExpr:  getString(doc, "cron_expr"),
			Enabled:   getBool(doc, "enabled"),
			CreatedAt: getTime(doc, "created_at"),
			UpdatedAt: getTime(doc, "updated_at"),
		}

		// Handle optional fields
		if lastRun, ok := doc["last_run"].(time.Time); ok {
			schedule.LastRun = &lastRun
		}
		if nextRun, ok := doc["next_run"].(time.Time); ok {
			schedule.NextRun = &nextRun
		}

		// Handle arrays
		if promptIDs, ok := doc["prompt_ids"].([]interface{}); ok {
			for _, id := range promptIDs {
				if str, ok := id.(string); ok {
					schedule.PromptIDs = append(schedule.PromptIDs, str)
				}
			}
		} else if promptIDs, ok := doc["prompt_ids"].(primitive.A); ok {
			for _, id := range promptIDs {
				if str, ok := id.(string); ok {
					schedule.PromptIDs = append(schedule.PromptIDs, str)
				}
			}
		}

		if llmIDs, ok := doc["llm_ids"].([]interface{}); ok {
			for _, id := range llmIDs {
				if str, ok := id.(string); ok {
					schedule.LLMIDs = append(schedule.LLMIDs, str)
				}
			}
		} else if llmIDs, ok := doc["llm_ids"].(primitive.A); ok {
			for _, id := range llmIDs {
				if str, ok := id.(string); ok {
					schedule.LLMIDs = append(schedule.LLMIDs, str)
				}
			}
		}

		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// UpdateSchedule updates an existing schedule
func (m *MongoDB) UpdateSchedule(ctx context.Context, schedule *models.Schedule) error {
	schedule.UpdatedAt = time.Now()

	// Convert to BSON document with explicit _id field
	doc := bson.M{
		"_id":        schedule.ID,
		"name":       schedule.Name,
		"cron_expr":  schedule.CronExpr,
		"enabled":    schedule.Enabled,
		"prompt_ids": schedule.PromptIDs,
		"llm_ids":    schedule.LLMIDs,
		"created_at": schedule.CreatedAt,
		"updated_at": schedule.UpdatedAt,
	}

	// Add optional fields if they exist
	if schedule.LastRun != nil {
		doc["last_run"] = *schedule.LastRun
	}
	if schedule.NextRun != nil {
		doc["next_run"] = *schedule.NextRun
	}

	// Try to parse as ObjectID first (for old documents), then as string
	var filter bson.M
	if objectID, err := primitive.ObjectIDFromHex(schedule.ID); err == nil {
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": schedule.ID}
	}

	result, err := m.database.Collection(collSchedules).ReplaceOne(ctx, filter, doc)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("schedule not found: %s", schedule.ID)
	}

	return nil
}

// DeleteSchedule deletes a schedule by ID
func (m *MongoDB) DeleteSchedule(ctx context.Context, id string) error {
	// Try to parse as ObjectID first (for old documents), then as string
	var filter bson.M
	if objectID, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": id}
	}

	result, err := m.database.Collection(collSchedules).DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("schedule not found: %s", id)
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

// DeleteAllSchedules deletes all schedules
func (m *MongoDB) DeleteAllSchedules(ctx context.Context) (int, error) {
	result, err := m.database.Collection(collSchedules).DeleteMany(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	return int(result.DeletedCount), nil
}

// CreateResponse creates a new response
func (m *MongoDB) CreateResponse(ctx context.Context, response *models.Response) error {
	response.CreatedAt = time.Now()

	// Convert to BSON document with explicit field names
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
		"created_at":    response.CreatedAt,
	}

	// Add metadata if it exists
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
func (m *MongoDB) ListResponses(ctx context.Context, filter db.ResponseFilter) ([]*models.Response, error) {
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
		query["keywords.keyword"] = filter.Keyword
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
		// Handle time.Time directly
		if t, ok := val.(time.Time); ok {
			return t
		}
		// Handle primitive.DateTime
		if dt, ok := val.(primitive.DateTime); ok {
			return dt.Time()
		}
		// Handle int64 (Unix timestamp)
		if ts, ok := val.(int64); ok {
			return time.Unix(ts, 0)
		}
		// Handle float64 (Unix timestamp)
		if ts, ok := val.(float64); ok {
			return time.Unix(int64(ts), 0)
		}
	}
	return time.Time{}
}

// getStringFromEither gets a string value from either of two possible field names
func getStringFromEither(doc bson.M, field1, field2 string) string {
	// Always prefer the correct field name (lowercase with underscores)
	if val, ok := doc[field1]; ok && val != nil {
		if str, ok := val.(string); ok && str != "" {
			return str
		}
	}
	// Only try the alternative if the first field doesn't exist or is empty
	if val, ok := doc[field2]; ok && val != nil {
		if str, ok := val.(string); ok && str != "" {
			return str
		}
	}
	return ""
}
