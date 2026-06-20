package anthropic

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
	model := "claude-3-7-sonnet-20250219"
	if config.Model != "" {
		model = config.Model
	}

	temperature := config.Temperature
	maxTokens := config.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1000
	}

	tools := []map[string]interface{}{
		{
			"type":     "web_search_20250305",
			"name":     "web_search",
			"max_uses": 5,
		},
	}

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": maxTokens,
		"tools":      tools,
	}

	if modelSupportsSamplingParameters(model) && temperature > 0 {
		requestBody["temperature"] = temperature
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
		Content []json.RawMessage `json:"content"`
		Usage   struct {
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

	var generatedText string
	var searchURLs []llm.SearchURL
	var searchQuery string

	for _, contentItem := range anthropicResp.Content {
		var contentBlock struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(contentItem, &contentBlock); err != nil {
			continue
		}

		switch contentBlock.Type {
		case "text":
			var textBlock struct {
				Type      string `json:"type"`
				Text      string `json:"text"`
				Citations []struct {
					Type           string `json:"type"`
					URL            string `json:"url"`
					Title          string `json:"title"`
					EncryptedIndex string `json:"encrypted_index"`
					CitedText      string `json:"cited_text"`
				} `json:"citations,omitempty"`
			}
			if err := json.Unmarshal(contentItem, &textBlock); err == nil {
				if textBlock.Text != "" {
					generatedText += textBlock.Text
				}
				for index, citation := range textBlock.Citations {
					if citation.Type == "web_search_result_location" && citation.URL != "" {
						resolvedURL := resolveRedirectURL(ctx, citation.URL)
						searchURLs = append(searchURLs, llm.SearchURL{
							SearchQuery:   searchQuery,
							URL:           resolvedURL,
							Title:         citation.Title,
							CitationIndex: index,
						})
					}
				}
			}

		case "server_tool_use":
			var toolUseBlock struct {
				Type  string `json:"type"`
				ID    string `json:"id"`
				Name  string `json:"name"`
				Input struct {
					Query string `json:"query"`
				} `json:"input"`
			}
			if err := json.Unmarshal(contentItem, &toolUseBlock); err == nil {
				if toolUseBlock.Name == "web_search" && toolUseBlock.Input.Query != "" {
					searchQuery = toolUseBlock.Input.Query
				}
			}

		case "web_search_tool_result":
			var searchResultBlock struct {
				Type      string          `json:"type"`
				ToolUseID string          `json:"tool_use_id"`
				Content   json.RawMessage `json:"content"`
			}
			if err := json.Unmarshal(contentItem, &searchResultBlock); err == nil {
				var contentArray []json.RawMessage
				if err := json.Unmarshal(searchResultBlock.Content, &contentArray); err == nil {
					for _, resultItem := range contentArray {
						var resultContent struct {
							Type             string `json:"type"`
							URL              string `json:"url"`
							Title            string `json:"title"`
							EncryptedContent string `json:"encrypted_content"`
							PageAge          string `json:"page_age"`
						}
						if err := json.Unmarshal(resultItem, &resultContent); err == nil {
							if resultContent.Type == "web_search_result" && resultContent.URL != "" {
								resolvedURL := resolveRedirectURL(ctx, resultContent.URL)
								searchURLs = append(searchURLs, llm.SearchURL{
									SearchQuery:   searchQuery,
									URL:           resolvedURL,
									Title:         resultContent.Title,
									CitationIndex: len(searchURLs),
								})
							}
						}
					}
				} else {
					var errorContent struct {
						Type      string `json:"type"`
						ErrorCode string `json:"error_code"`
					}
					if err := json.Unmarshal(searchResultBlock.Content, &errorContent); err == nil {
						if errorContent.Type == "web_search_tool_result_error" {
							continue
						}
					}
				}
			}
		}
	}

	if generatedText == "" {
		return nil, fmt.Errorf("no text content returned from API")
	}

	totalTokens := anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens

	return &llm.Response{
		Text:       generatedText,
		TokensUsed: totalTokens,
		Model:      anthropicResp.Model,
		Provider:   "anthropic",
		SearchURLs: searchURLs,
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
	var popularModels []models.ModelInfo
	popularModelIDs := map[string]bool{
		"claude-3-7-sonnet-20250219": true,
		"claude-3-5-sonnet-20241022": true,
		"claude-3-opus-20240229":     true,
		"claude-3-5-haiku-20241022":  true,
		"claude-3-haiku-20240307":    true,
	}

	for _, model := range response.Data {
		if model.Type == "model" {
			info := models.ModelInfo{
				ID:          model.ID,
				Name:        model.DisplayName,
				Description: fmt.Sprintf("Anthropic %s (created: %s)", model.DisplayName, model.CreatedAt),
				UsedInChat:  popularModelIDs[model.ID],
			}

			if popularModelIDs[model.ID] {
				popularModels = append(popularModels, info)
			} else {
				modelList = append(modelList, info)
			}
		}
	}

	return append(popularModels, modelList...), nil
}

func modelSupportsSamplingParameters(model string) bool {
	lower := strings.ToLower(model)
	if strings.Contains(lower, "opus-4-7") || strings.Contains(lower, "opus-4-8") {
		return false
	}
	return true
}

func resolveRedirectURL(ctx context.Context, url string) string {
	if url == "" {
		return url
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return url
	}

	resp, err := client.Do(req)
	if err != nil {
		return url
	}
	defer resp.Body.Close()

	if resp.Request != nil && resp.Request.URL != nil {
		return resp.Request.URL.String()
	}

	return url
}
