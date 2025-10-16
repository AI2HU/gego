package services

import (
	"context"
	"fmt"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
)

// LLMService provides business logic for LLM management
type LLMService struct {
	db db.Database
}

// NewLLMService creates a new LLM service
func NewLLMService(database db.Database) *LLMService {
	return &LLMService{db: database}
}

// Provider represents available LLM providers
type Provider int

const (
	OpenAI Provider = iota + 1
	Anthropic
	Ollama
	Google
	Perplexity
)

// String returns the string representation of the provider
func (p Provider) String() string {
	switch p {
	case OpenAI:
		return "openai"
	case Anthropic:
		return "anthropic"
	case Ollama:
		return "ollama"
	case Google:
		return "google"
	case Perplexity:
		return "perplexity"
	default:
		return "unknown"
	}
}

// FromString converts a string to a Provider
func FromString(s string) Provider {
	switch s {
	case "openai":
		return OpenAI
	case "anthropic":
		return Anthropic
	case "ollama":
		return Ollama
	case "google":
		return Google
	case "perplexity":
		return Perplexity
	default:
		return 0 // Unknown provider
	}
}

// DisplayName returns the display name for the provider with model info
func (p Provider) DisplayName() string {
	switch p {
	case OpenAI:
		return "OpenAI (ChatGPT)"
	case Anthropic:
		return "Anthropic (Claude)"
	case Ollama:
		return "Ollama (local models)"
	case Google:
		return "Google (Gemini)"
	case Perplexity:
		return "Perplexity (Sonar)"
	default:
		return "Unknown"
	}
}

// AllProviders returns a slice of all available providers
func AllProviders() []Provider {
	return []Provider{OpenAI, Anthropic, Ollama, Google, Perplexity}
}

// GetConsoleURL returns the console URL where API keys can be generated for the provider
func (p Provider) GetConsoleURL() string {
	switch p {
	case OpenAI:
		return "https://platform.openai.com/api-keys"
	case Anthropic:
		return "https://console.anthropic.com/"
	case Google:
		return "https://makersuite.google.com/app/apikey"
	case Perplexity:
		return "https://www.perplexity.ai/settings/api"
	case Ollama:
		return "https://ollama.ai/" // Ollama doesn't need API keys, but provides setup info
	default:
		return ""
	}
}

// MaskAPIKey masks the API key for display (shows first 4 and last 4 characters)
func MaskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "(not set)"
	}
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}

// GetExistingAPIKeysForProvider returns existing API keys for a given provider
func (s *LLMService) GetExistingAPIKeysForProvider(ctx context.Context, provider string) ([]string, error) {
	llms, err := s.db.ListLLMs(ctx, nil)
	if err != nil {
		return nil, err
	}

	var apiKeys []string
	seenKeys := make(map[string]bool)

	for _, llm := range llms {
		if llm.Provider == provider && llm.APIKey != "" {
			if !seenKeys[llm.APIKey] {
				apiKeys = append(apiKeys, llm.APIKey)
				seenKeys[llm.APIKey] = true
			}
		}
	}

	return apiKeys, nil
}

// ValidateLLMConfig validates LLM configuration
func (s *LLMService) ValidateLLMConfig(config *models.LLMConfig) error {
	if config.Name == "" {
		return fmt.Errorf("LLM name is required")
	}
	if config.Provider == "" {
		return fmt.Errorf("LLM provider is required")
	}
	if config.Model == "" {
		return fmt.Errorf("LLM model is required")
	}

	provider := FromString(config.Provider)
	if provider == 0 {
		return fmt.Errorf("unknown provider: %s", config.Provider)
	}

	if provider != Ollama && config.APIKey == "" {
		return fmt.Errorf("API key is required for %s", provider.DisplayName())
	}

	return nil
}

// CreateLLM creates a new LLM configuration
func (s *LLMService) CreateLLM(ctx context.Context, config *models.LLMConfig) error {
	if err := s.ValidateLLMConfig(config); err != nil {
		return err
	}
	return s.db.CreateLLM(ctx, config)
}

// UpdateLLM updates an existing LLM configuration
func (s *LLMService) UpdateLLM(ctx context.Context, config *models.LLMConfig) error {
	if err := s.ValidateLLMConfig(config); err != nil {
		return err
	}
	return s.db.UpdateLLM(ctx, config)
}

// GetLLM retrieves an LLM configuration by ID
func (s *LLMService) GetLLM(ctx context.Context, id string) (*models.LLMConfig, error) {
	return s.db.GetLLM(ctx, id)
}

// ListLLMs lists LLM configurations with optional filtering
func (s *LLMService) ListLLMs(ctx context.Context, enabled *bool) ([]*models.LLMConfig, error) {
	return s.db.ListLLMs(ctx, enabled)
}

// DeleteLLM deletes an LLM configuration
func (s *LLMService) DeleteLLM(ctx context.Context, id string) error {
	return s.db.DeleteLLM(ctx, id)
}

// EnableLLM enables an LLM configuration
func (s *LLMService) EnableLLM(ctx context.Context, id string) error {
	llm, err := s.db.GetLLM(ctx, id)
	if err != nil {
		return err
	}
	llm.Enabled = true
	return s.db.UpdateLLM(ctx, llm)
}

// DisableLLM disables an LLM configuration
func (s *LLMService) DisableLLM(ctx context.Context, id string) error {
	llm, err := s.db.GetLLM(ctx, id)
	if err != nil {
		return err
	}
	llm.Enabled = false
	return s.db.UpdateLLM(ctx, llm)
}

// GetEnabledLLMs returns only enabled LLM configurations
func (s *LLMService) GetEnabledLLMs(ctx context.Context) ([]*models.LLMConfig, error) {
	enabled := true
	return s.db.ListLLMs(ctx, &enabled)
}

// ValidateProviderModel validates if a model is available for a provider
func (s *LLMService) ValidateProviderModel(provider string, model string) error {
	if provider == "" {
		return fmt.Errorf("provider is required")
	}
	if model == "" {
		return fmt.Errorf("model is required")
	}

	if FromString(provider) == 0 {
		return fmt.Errorf("unknown provider: %s", provider)
	}

	return nil
}
