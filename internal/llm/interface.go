package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AI2HU/gego/internal/models"
)

// Provider defines the interface for LLM providers
type Provider interface {
	// Name returns the provider name (e.g., "openai", "anthropic")
	Name() string

	// Generate sends a prompt to the LLM and returns the response
	Generate(ctx context.Context, prompt string, config map[string]interface{}) (*Response, error)

	// Validate validates the provider configuration
	Validate(config map[string]string) error

	// ListModels lists available text-to-text models from this provider
	// Returns models that can be used for text generation
	ListModels(ctx context.Context, apiKey, baseURL string) ([]models.ModelInfo, error)
}

// Response represents an LLM response
type Response struct {
	Text       string
	TokensUsed int
	LatencyMs  int64
	Model      string
	Provider   string
	Error      string
}

// Registry manages LLM providers
type Registry struct {
	providers map[string]Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register registers a provider
func (r *Registry) Register(provider Provider) {
	r.providers[provider.Name()] = provider
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (Provider, bool) {
	provider, ok := r.providers[name]
	return provider, ok
}

// List returns all registered provider names
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// GenerateRequest encapsulates a generation request
type GenerateRequest struct {
	Provider string
	Model    string
	Prompt   string
	Config   map[string]interface{}
	Timeout  time.Duration
}

// GenerateGEOPromptTemplate creates a pre-prompt for generating GEO statistics prompts
func GenerateGEOPromptTemplate(userInput string, existingPrompts []string, language string, count int) string {
	existingPromptsText := ""
	if len(existingPrompts) > 0 {
		existingPromptsText = fmt.Sprintf(`

IMPORTANT: Avoid repetition! The following prompts already exist in the system:
%s

Please generate NEW prompts that are different from the existing ones above.`, strings.Join(existingPrompts, "\n"))
	}

	languageInstruction := ""
	if language != "EN" {
		languageInstruction = fmt.Sprintf(`

IMPORTANT: All prompts must be written in %s language. Do not use English unless specifically requested.`, language)
	}

	return fmt.Sprintf(`You are a prompt generation assistant. The user wants to create prompts that people would naturally ask LLMs when looking for information.

User's request: %s%s%s

Please generate exactly %d different prompts that users would typically ask when searching for information. Each prompt should:
1. Be a natural, conversational question or request that people would ask an AI assistant
2. Be likely to generate responses that mention specific brands, products, services, or companies
3. Be varied in style and approach (questions, requests, scenarios, comparisons, recommendations)
4. Be suitable for different LLM models
5. Sound like something a real person would ask when looking for information
6. Cover different aspects and angles of the topic%s

Format your response as a simple list with each prompt on a new line. Do not include numbers, bullet points, dashes (-), or any additional text or explanations.`, userInput, existingPromptsText, languageInstruction, count, languageInstruction)
}
