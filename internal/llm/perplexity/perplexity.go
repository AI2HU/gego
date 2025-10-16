package perplexity

import (
	"context"
	"fmt"
	"time"

	pplx "github.com/sgaunet/perplexity-go/v2"

	"github.com/AI2HU/gego/internal/llm"
	"github.com/AI2HU/gego/internal/models"
)

// Provider implements the LLM Provider interface for Perplexity
type Provider struct {
	apiKey  string
	baseURL string
	client  *pplx.Client
}

// New creates a new Perplexity provider
func New(apiKey, baseURL string) *Provider {
	if baseURL == "" {
		baseURL = "https://api.perplexity.ai"
	}

	client := pplx.NewClient(apiKey)

	return &Provider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  client,
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "perplexity"
}

// Validate validates the provider configuration
func (p *Provider) Validate(config map[string]string) error {
	if config["api_key"] == "" {
		return fmt.Errorf("api_key is required")
	}
	return nil
}

// Generate sends a prompt to Perplexity and returns the response
func (p *Provider) Generate(ctx context.Context, prompt string, config llm.Config) (*llm.Response, error) {
	startTime := time.Now()

	model := "sonar"
	if config.Model != "" {
		model = config.Model
	}

	temperature := config.Temperature
	maxTokens := config.MaxTokens

	messages := []pplx.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	req := pplx.NewCompletionRequest(
		pplx.WithMessages(messages),
		pplx.WithModel(model),
		pplx.WithTemperature(temperature),
		pplx.WithMaxTokens(maxTokens),
		pplx.WithReturnImages(false),
	)

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	resp, err := p.client.SendCompletionRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	content := resp.GetLastContent()
	if content == "" {
		return nil, fmt.Errorf("no content returned from API")
	}

	tokensUsed := resp.Usage.TotalTokens

	return &llm.Response{
		Text:       content,
		TokensUsed: tokensUsed,
		LatencyMs:  time.Since(startTime).Milliseconds(),
		Model:      model,
		Provider:   "perplexity",
	}, nil
}

// ListModels lists available text-to-text models from Perplexity
// Since Perplexity doesn't have a public models API, we return a curated list
func (p *Provider) ListModels(ctx context.Context, apiKey, baseURL string) ([]models.ModelInfo, error) {
	return []models.ModelInfo{
		{
			ID:          "sonar",
			Name:        "Sonar",
			Description: "Default model optimized for general use cases",
		},
		{
			ID:          "sonar-pro",
			Name:        "Sonar Pro",
			Description: "Advanced model for complex tasks and longer outputs",
		},
		{
			ID:          "sonar-medium",
			Name:        "Sonar Medium",
			Description: "Balanced model for moderate complexity tasks",
		},
		{
			ID:          "sonar-small",
			Name:        "Sonar Small",
			Description: "Fast and efficient model for simple tasks",
		},
		{
			ID:          "sonar-large",
			Name:        "Sonar Large",
			Description: "Most capable model for highly complex tasks",
		},
	}, nil
}
