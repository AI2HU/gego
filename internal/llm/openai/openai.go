package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AI2HU/gego/internal/llm"
	"github.com/AI2HU/gego/internal/models"
)

// Provider implements the LLM Provider interface for OpenAI
type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// New creates a new OpenAI provider
func New(apiKey, baseURL string) *Provider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &Provider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
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
func (p *Provider) Generate(ctx context.Context, prompt string, config map[string]interface{}) (*llm.Response, error) {
	startTime := time.Now()

	model := "gpt-3.5-turbo"
	if m, ok := config["model"].(string); ok && m != "" {
		model = m
	}

	temperature := 0.7
	if t, ok := config["temperature"].(float64); ok {
		temperature = t
	}

	maxTokens := 1000
	if mt, ok := config["max_tokens"].(int); ok {
		maxTokens = mt
	}

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": temperature,
		"max_tokens":  maxTokens,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return &llm.Response{
			Error:     err.Error(),
			LatencyMs: time.Since(startTime).Milliseconds(),
		}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &llm.Response{
			Error:     fmt.Sprintf("API error: %s", string(body)),
			LatencyMs: time.Since(startTime).Milliseconds(),
		}, nil
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}

	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return &llm.Response{
			Error:     "no choices returned from API",
			LatencyMs: time.Since(startTime).Milliseconds(),
		}, nil
	}

	return &llm.Response{
		Text:       openAIResp.Choices[0].Message.Content,
		TokensUsed: openAIResp.Usage.TotalTokens,
		LatencyMs:  time.Since(startTime).Milliseconds(),
		Model:      openAIResp.Model,
		Provider:   "openai",
	}, nil
}

// ListModels lists available text-to-text models from OpenAI
func (p *Provider) ListModels(ctx context.Context, apiKey, baseURL string) ([]models.ModelInfo, error) {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var listResp struct {
		Data []struct {
			ID         string `json:"id"`
			OwnedBy    string `json:"owned_by"`
			Permission []struct {
				AllowCreateEngine bool `json:"allow_create_engine"`
			} `json:"permission"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Filter for text-to-text models only (GPT chat models)
	var textModels []models.ModelInfo
	seen := make(map[string]bool)

	for _, model := range listResp.Data {
		modelID := strings.ToLower(model.ID)

		// Only include GPT models that support chat completions
		if strings.HasPrefix(modelID, "gpt-") && !seen[model.ID] {
			// Skip fine-tuned models (contains colons)
			if strings.Contains(model.ID, ":") {
				continue
			}

			// Skip embedding models
			if strings.Contains(modelID, "embed") || strings.Contains(modelID, "embedding") {
				continue
			}

			// Skip image models
			if strings.Contains(modelID, "vision") || strings.Contains(modelID, "image") {
				continue
			}

			// Skip audio models
			if strings.Contains(modelID, "whisper") || strings.Contains(modelID, "audio") {
				continue
			}

			textModels = append(textModels, models.ModelInfo{
				ID:          model.ID,
				Name:        model.ID,
				Description: fmt.Sprintf("OpenAI %s", model.ID),
			})
			seen[model.ID] = true
		}
	}

	return textModels, nil
}
