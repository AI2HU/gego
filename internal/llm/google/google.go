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
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
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
func (p *Provider) Generate(ctx context.Context, prompt string, config llm.Config) (*llm.Response, error) {
	startTime := time.Now()

	model := "gemini-1.5-flash"
	if config.Model != "" {
		model = config.Model
	}

	client := p.client
	if client == nil {
		var err error
		client, err = genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  p.apiKey,
			Backend: genai.BackendGeminiAPI,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Google client: %w", err)
		}
	}

	content := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: prompt},
			},
		},
	}

	generationConfig := &genai.GenerateContentConfig{
		Temperature: float32Ptr(float32(config.Temperature)),
		TopP:        float32Ptr(float32(config.TopP)),
		TopK:        float32Ptr(float32(config.TopK)),
	}

	result, err := client.Models.GenerateContent(ctx, model, content, generationConfig)
	if err != nil {
		return nil, fmt.Errorf("Google AI API error: %v", err)
	}

	var generatedText string
	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		if text := result.Candidates[0].Content.Parts[0].Text; text != "" {
			generatedText = text
		}
	}

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
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Google client: %w", err)
	}

	modelPage, err := client.Models.List(ctx, &genai.ListModelsConfig{})
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	var modelList []models.ModelInfo
	for _, model := range modelPage.Items {
		modelName := model.Name

		if strings.Contains(strings.ToLower(modelName), "embed") || strings.Contains(strings.ToLower(modelName), "embedding") {
			continue
		}

		if strings.Contains(strings.ToLower(modelName), "vision") || strings.Contains(strings.ToLower(modelName), "image") {
			continue
		}

		if strings.Contains(strings.ToLower(modelName), "gemini") {
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

func float32Ptr(f float32) *float32 {
	return &f
}
