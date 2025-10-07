package google

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

// Provider implements the LLM Provider interface for Google AI
type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// New creates a new Google provider
func New(apiKey, baseURL string) *Provider {
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
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
	start := time.Now()

	// Get model from config, default to gemini-pro
	model := "gemini-pro"
	if m, ok := config["model"].(string); ok && m != "" {
		model = m
	}

	// Prepare the request payload
	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": prompt,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature": 0.7,
			"topP":        0.8,
			"topK":        40,
		},
	}

	// Override generation config if provided
	if temp, ok := config["temperature"].(float64); ok {
		payload["generationConfig"].(map[string]interface{})["temperature"] = temp
	}
	if topP, ok := config["top_p"].(float64); ok {
		payload["generationConfig"].(map[string]interface{})["topP"] = topP
	}
	if topK, ok := config["top_k"].(float64); ok {
		payload["generationConfig"].(map[string]interface{})["topK"] = topK
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the request
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.baseURL, model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return &llm.Response{
				Error: fmt.Sprintf("Google AI API error (%d): %s", errorResp.Error.Code, errorResp.Error.Message),
			}, nil
		}
		return &llm.Response{
			Error: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	// Parse the response
	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CompletionTokenCount int `json:"completionTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract the generated text
	var generatedText string
	if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
		generatedText = response.Candidates[0].Content.Parts[0].Text
	}

	latency := time.Since(start).Milliseconds()

	return &llm.Response{
		Text:       generatedText,
		TokensUsed: response.UsageMetadata.TotalTokenCount,
		LatencyMs:  latency,
		Model:      model,
		Provider:   "google",
	}, nil
}

// ListModels lists available Google AI models
func (p *Provider) ListModels(ctx context.Context, apiKey, baseURL string) ([]models.ModelInfo, error) {
	if apiKey == "" {
		apiKey = p.apiKey
	}
	if baseURL == "" {
		baseURL = p.baseURL
	}

	// Create a temporary client for this request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	url := fmt.Sprintf("%s/models?key=%s", baseURL, apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

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
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("Google AI API error (%d): %s", errorResp.Error.Code, errorResp.Error.Message)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Models []struct {
			Name                       string   `json:"name"`
			DisplayName                string   `json:"displayName"`
			Description                string   `json:"description"`
			SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
		} `json:"models"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var modelList []models.ModelInfo
	for _, model := range response.Models {
		// Only include models that support text generation
		supportsGeneration := false
		for _, method := range model.SupportedGenerationMethods {
			if method == "generateContent" {
				supportsGeneration = true
				break
			}
		}

		if supportsGeneration {
			// Additional filtering for text-to-text models only
			modelName := strings.ToLower(model.Name)

			// Skip embedding models
			if strings.Contains(modelName, "embed") || strings.Contains(modelName, "embedding") {
				continue
			}

			// Skip image models
			if strings.Contains(modelName, "vision") || strings.Contains(modelName, "image") {
				continue
			}

			// Skip multimodal models that aren't primarily text
			if strings.Contains(modelName, "multimodal") && !strings.Contains(modelName, "gemini") {
				continue
			}

			// Extract model name from full path (e.g., "models/gemini-pro" -> "gemini-pro")
			name := strings.TrimPrefix(model.Name, "models/")

			modelList = append(modelList, models.ModelInfo{
				ID:          model.Name,
				Name:        name,
				Description: model.Description,
			})
		}
	}

	return modelList, nil
}
