package google

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/genai"

	"github.com/AI2HU/gego/internal/llm"
	"github.com/AI2HU/gego/internal/models"
)

// Provider implements the LLM Provider interface for Google AI
type Provider struct {
	apiKey  string
	baseURL string
	client  *genai.Client
}

// New creates a new Google provider
func New(apiKey, baseURL string) *Provider {
	// Create client using the official SDK
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		// If client creation fails, we'll handle it in Generate method
		client = nil
	}

	return &Provider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  client,
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "google"
}

// Validate validates the provider configuration
func (p *Provider) Validate(config map[string]string) error {
	if config["api_key"] == "" {
		return fmt.Errorf("api_key is required")
	}
	return nil
}

// Generate sends a prompt to Google AI and returns the response
func (p *Provider) Generate(ctx context.Context, prompt string, config map[string]interface{}) (*llm.Response, error) {
	startTime := time.Now()

	// Get model from config, default to gemini-1.5-flash
	model := "gemini-1.5-flash"
	if m, ok := config["model"].(string); ok && m != "" {
		model = m
	}

	// Create client if not already created
	client := p.client
	if client == nil {
		var err error
		client, err = genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  p.apiKey,
			Backend: genai.BackendGeminiAPI,
		})
		if err != nil {
			return &llm.Response{
				Error:     fmt.Sprintf("failed to create Google client: %v", err),
				LatencyMs: time.Since(startTime).Milliseconds(),
			}, nil
		}
	}

	// Create content with the prompt
	content := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: prompt},
			},
		},
	}

	// Set generation config
	generationConfig := &genai.GenerateContentConfig{
		Temperature: float32Ptr(0.7),
		TopP:        float32Ptr(0.8),
		TopK:        float32Ptr(40),
	}

	// Override generation config if provided
	if temp, ok := config["temperature"].(float64); ok {
		generationConfig.Temperature = float32Ptr(float32(temp))
	}
	if topP, ok := config["top_p"].(float64); ok {
		generationConfig.TopP = float32Ptr(float32(topP))
	}
	if topK, ok := config["top_k"].(float64); ok {
		generationConfig.TopK = float32Ptr(float32(topK))
	}

	// Generate content
	result, err := client.Models.GenerateContent(ctx, model, content, generationConfig)
	if err != nil {
		return &llm.Response{
			Error:     fmt.Sprintf("Google AI API error: %v", err),
			LatencyMs: time.Since(startTime).Milliseconds(),
		}, nil
	}

	// Extract the generated text
	var generatedText string
	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		if text := result.Candidates[0].Content.Parts[0].Text; text != "" {
			generatedText = text
		}
	}

	// Get token usage
	tokensUsed := 0
	if result.UsageMetadata != nil {
		tokensUsed = int(result.UsageMetadata.TotalTokenCount)
	}

	return &llm.Response{
		Text:       generatedText,
		TokensUsed: tokensUsed,
		LatencyMs:  time.Since(startTime).Milliseconds(),
		Model:      model,
		Provider:   "google",
	}, nil
}

// ListModels lists available Google AI models
func (p *Provider) ListModels(ctx context.Context, apiKey, baseURL string) ([]models.ModelInfo, error) {
	// Create client for this request
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Google client: %w", err)
	}

	// List models using the SDK
	modelPage, err := client.Models.List(ctx, &genai.ListModelsConfig{})
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	var modelList []models.ModelInfo
	for _, model := range modelPage.Items {
		// Filter for text-to-text models only
		modelName := model.Name

		// Skip embedding models
		if strings.Contains(strings.ToLower(modelName), "embed") || strings.Contains(strings.ToLower(modelName), "embedding") {
			continue
		}

		// Skip image models
		if strings.Contains(strings.ToLower(modelName), "vision") || strings.Contains(strings.ToLower(modelName), "image") {
			continue
		}

		// Only include Gemini models
		if strings.Contains(strings.ToLower(modelName), "gemini") {
			// Extract model name from full path (e.g., "models/gemini-pro" -> "gemini-pro")
			name := modelName
			if len(name) > 7 && name[:7] == "models/" {
				name = name[7:]
			}

			modelList = append(modelList, models.ModelInfo{
				ID:          model.Name,
				Name:        name,
				Description: model.Description,
			})
		}
	}

	return modelList, nil
}

// Helper function to create float32 pointer
func float32Ptr(f float32) *float32 {
	return &f
}
