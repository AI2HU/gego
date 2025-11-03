package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AI2HU/gego/internal/llm"
	"github.com/AI2HU/gego/internal/models"
)

// Provider implements the LLM Provider interface for Anthropic
type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// New creates a new Anthropic provider
func New(apiKey, baseURL string) *Provider {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
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
	return "anthropic"
}

// Validate validates the provider configuration
func (p *Provider) Validate(config map[string]string) error {
	if config["api_key"] == "" {
		return fmt.Errorf("api_key is required")
	}
	return nil
}

// Generate sends a prompt to Anthropic and returns the response
func (p *Provider) Generate(ctx context.Context, prompt string, config llm.Config) (*llm.Response, error) {
	startTime := time.Now()

	model := "claude-3-7-sonnet-20250219"
	if config.Model != "" {
		model = config.Model
	}

	temperature := config.Temperature
	maxTokens := config.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1000
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

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var anthropicResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}

	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(anthropicResp.Content) == 0 {
		return nil, fmt.Errorf("no content returned from API")
	}

	totalTokens := anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens

	return &llm.Response{
		Text:       anthropicResp.Content[0].Text,
		TokensUsed: totalTokens,
		LatencyMs:  time.Since(startTime).Milliseconds(),
		Model:      anthropicResp.Model,
		Provider:   "anthropic",
	}, nil
}

// ListModels lists available text-to-text models from Anthropic
func (p *Provider) ListModels(ctx context.Context, apiKey, baseURL string) ([]models.ModelInfo, error) {
	if apiKey == "" {
		apiKey = p.apiKey
	}
	if baseURL == "" {
		baseURL = p.baseURL
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("Anthropic API error (%s): %s", errorResp.Error.Type, errorResp.Error.Message)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data []struct {
			ID          string `json:"id"`
			DisplayName string `json:"display_name"`
			CreatedAt   string `json:"created_at"`
			Type        string `json:"type"`
		} `json:"data"`
		FirstID string `json:"first_id"`
		HasMore bool   `json:"has_more"`
		LastID  string `json:"last_id"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var modelList []models.ModelInfo
	for _, model := range response.Data {
		if model.Type == "model" {
			modelList = append(modelList, models.ModelInfo{
				ID:          model.ID,
				Name:        model.DisplayName,
				Description: fmt.Sprintf("Anthropic %s (created: %s)", model.DisplayName, model.CreatedAt),
			})
		}
	}

	return modelList, nil
}
