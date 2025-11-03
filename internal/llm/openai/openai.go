package openai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"

	"github.com/AI2HU/gego/internal/llm"
	"github.com/AI2HU/gego/internal/models"
)

// Provider implements the LLM Provider interface for OpenAI
type Provider struct {
	apiKey  string
	baseURL string
	client  openai.Client
}

// New creates a new OpenAI provider
func New(apiKey, baseURL string) *Provider {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	if baseURL != "" && baseURL != "https://api.openai.com/v1" {
		client = openai.NewClient(
			option.WithAPIKey(apiKey),
			option.WithBaseURL(baseURL),
		)
	}

	return &Provider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  client,
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "openai"
}

// Validate validates the provider configuration
func (p *Provider) Validate(config map[string]string) error {
	if config["api_key"] == "" {
		return fmt.Errorf("api_key is required")
	}
	return nil
}

// Generate sends a prompt to OpenAI and returns the response
func (p *Provider) Generate(ctx context.Context, prompt string, config llm.Config) (*llm.Response, error) {
	startTime := time.Now()

	model := shared.ChatModelGPT3_5Turbo
	if config.Model != "" {
		model = shared.ChatModel(config.Model)
	}

	temperature := config.Temperature
	maxTokens := config.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1000
	}

	chatCompletion, err := p.client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model: model,
			Messages: []openai.ChatCompletionMessageParamUnion{
				{
					OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.ChatCompletionUserMessageParamContentUnion{
							OfString: openai.String(prompt),
						},
					},
				},
			},
			Temperature: openai.Float(temperature),
			MaxTokens:   openai.Int(int64(maxTokens)),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	var generatedText string
	if len(chatCompletion.Choices) > 0 && chatCompletion.Choices[0].Message.Content != "" {
		generatedText = chatCompletion.Choices[0].Message.Content
	}

	tokensUsed := 0
	if chatCompletion.Usage.TotalTokens != 0 {
		tokensUsed = int(chatCompletion.Usage.TotalTokens)
	}

	return &llm.Response{
		Text:       generatedText,
		TokensUsed: tokensUsed,
		LatencyMs:  time.Since(startTime).Milliseconds(),
		Model:      string(model),
		Provider:   "openai",
	}, nil
}

// ListModels lists available text-to-text models from OpenAI
func (p *Provider) ListModels(ctx context.Context, apiKey, baseURL string) ([]models.ModelInfo, error) {
	client := p.client
	if apiKey != "" && apiKey != p.apiKey {
		client = openai.NewClient(
			option.WithAPIKey(apiKey),
		)
		if baseURL != "" && baseURL != "https://api.openai.com/v1" {
			client = openai.NewClient(
				option.WithAPIKey(apiKey),
				option.WithBaseURL(baseURL),
			)
		}
	}

	modelList, err := client.Models.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	var textModels []models.ModelInfo
	for _, model := range modelList.Data {
		modelID := string(model.ID)

		if strings.HasPrefix(strings.ToLower(modelID), "gpt") {
			if strings.Contains(modelID, ":") {
				continue
			}

			if strings.Contains(strings.ToLower(modelID), "embed") || strings.Contains(strings.ToLower(modelID), "embedding") {
				continue
			}

			if strings.Contains(strings.ToLower(modelID), "vision") || strings.Contains(strings.ToLower(modelID), "image") {
				continue
			}

			if strings.Contains(strings.ToLower(modelID), "whisper") || strings.Contains(strings.ToLower(modelID), "audio") {
				continue
			}

			textModels = append(textModels, models.ModelInfo{
				ID:          modelID,
				Name:        modelID,
				Description: fmt.Sprintf("OpenAI %s", modelID),
			})
		}
	}

	return textModels, nil
}
