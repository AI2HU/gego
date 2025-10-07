package ollama

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

// Provider implements the LLM Provider interface for Ollama
type Provider struct {
	baseURL string
	client  *http.Client
}

// New creates a new Ollama provider
func New(baseURL string) *Provider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	return &Provider{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "ollama"
}

// Validate validates the provider configuration
func (p *Provider) Validate(config map[string]string) error {
	// Ollama doesn't require API key, just a reachable endpoint
	return nil
}

// Generate sends a prompt to Ollama and returns the response
func (p *Provider) Generate(ctx context.Context, prompt string, config map[string]interface{}) (*llm.Response, error) {
	startTime := time.Now()

	model := "llama2"
	if m, ok := config["model"].(string); ok && m != "" {
		model = m
	}

	temperature := 0.7
	if t, ok := config["temperature"].(float64); ok {
		temperature = t
	}

	requestBody := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": temperature,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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

	var ollamaResp struct {
		Model         string `json:"model"`
		Response      string `json:"response"`
		Done          bool   `json:"done"`
		Context       []int  `json:"context"`
		TotalDuration int64  `json:"total_duration"`
	}

	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Ollama doesn't provide token count in the same way, estimate from context length
	tokensUsed := len(ollamaResp.Context)

	return &llm.Response{
		Text:       ollamaResp.Response,
		TokensUsed: tokensUsed,
		LatencyMs:  time.Since(startTime).Milliseconds(),
		Model:      ollamaResp.Model,
		Provider:   "ollama",
	}, nil
}

// ListModels lists available models from Ollama
func (p *Provider) ListModels(ctx context.Context, apiKey, baseURL string) ([]models.ModelInfo, error) {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

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
		Models []struct {
			Name       string `json:"name"`
			ModifiedAt string `json:"modified_at"`
			Size       int64  `json:"size"`
		} `json:"models"`
	}

	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var textModels []models.ModelInfo
	for _, model := range listResp.Models {
		// Filter for text-to-text models only
		// Skip embedding models, image models, and other non-text models
		modelName := strings.ToLower(model.Name)

		// Skip embedding models
		if strings.Contains(modelName, "embed") || strings.Contains(modelName, "embedding") {
			continue
		}

		// Skip image models
		if strings.Contains(modelName, "vision") || strings.Contains(modelName, "image") || strings.Contains(modelName, "clip") {
			continue
		}

		// Skip code-specific models that aren't general text models
		if strings.Contains(modelName, "code") && !strings.Contains(modelName, "llama") && !strings.Contains(modelName, "mistral") {
			continue
		}

		// Skip multimodal models that aren't primarily text
		if strings.Contains(modelName, "multimodal") && !strings.Contains(modelName, "llama") {
			continue
		}

		textModels = append(textModels, models.ModelInfo{
			ID:          model.Name,
			Name:        model.Name,
			Description: fmt.Sprintf("Ollama %s (%.2f GB)", model.Name, float64(model.Size)/(1024*1024*1024)),
		})
	}

	return textModels, nil
}
