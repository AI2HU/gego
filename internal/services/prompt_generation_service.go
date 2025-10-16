package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/AI2HU/gego/internal/llm"
	"github.com/AI2HU/gego/internal/models"
)

// PromptGenerationService provides business logic for LLM-based prompt generation
type PromptGenerationService struct {
	llmRegistry *llm.Registry
}

// NewPromptGenerationService creates a new prompt generation service
func NewPromptGenerationService(registry *llm.Registry) *PromptGenerationService {
	return &PromptGenerationService{
		llmRegistry: registry,
	}
}

// GenerationConfig represents configuration for prompt generation
type GenerationConfig struct {
	LanguageCode    string   `json:"language_code"`
	UserInput       string   `json:"user_input"`
	PromptCount     int      `json:"prompt_count"`
	ExistingPrompts []string `json:"existing_prompts"`
}

// ValidateGenerationConfig validates generation configuration
func (s *PromptGenerationService) ValidateGenerationConfig(config *GenerationConfig) error {
	if config.LanguageCode == "" {
		return fmt.Errorf("language code is required")
	}
	if len(config.LanguageCode) < 2 || len(config.LanguageCode) > 3 {
		return fmt.Errorf("language code should be 2-3 characters (e.g., FR, EN, IT)")
	}
	if config.UserInput == "" {
		return fmt.Errorf("user input is required")
	}
	if config.PromptCount < 1 {
		return fmt.Errorf("prompt count must be at least 1")
	}
	if config.PromptCount > 100 {
		return fmt.Errorf("prompt count cannot exceed 100")
	}
	return nil
}

// GeneratePrompts generates prompts using an LLM
func (s *PromptGenerationService) GeneratePrompts(ctx context.Context, llmConfig *models.LLMConfig, config *GenerationConfig) ([]string, error) {
	if err := s.ValidateGenerationConfig(config); err != nil {
		return nil, err
	}

	provider, ok := s.llmRegistry.Get(llmConfig.Provider)
	if !ok {
		return nil, fmt.Errorf("LLM provider %s not found", llmConfig.Provider)
	}

	prePrompt := llm.GenerateGEOPromptTemplate(
		config.UserInput,
		config.ExistingPrompts,
		config.LanguageCode,
		config.PromptCount,
	)

	response, err := provider.Generate(ctx, prePrompt, llm.Config{
		Model: llmConfig.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate prompts: %w", err)
	}

	if response.Error != "" {
		return nil, fmt.Errorf("LLM error: %s", response.Error)
	}

	promptLines := strings.Split(strings.TrimSpace(response.Text), "\n")
	var generatedPrompts []string

	for _, line := range promptLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if len(line) > 2 && (line[1] == '.' || line[1] == ')') && (line[0] >= '0' && line[0] <= '9') {
			line = strings.TrimSpace(line[2:])
		}

		if line != "" {
			generatedPrompts = append(generatedPrompts, line)
		}
	}

	if len(generatedPrompts) == 0 {
		return nil, fmt.Errorf("no valid prompts were generated")
	}

	return generatedPrompts, nil
}

// GetLanguageName returns the display name for a language code
func GetLanguageName(languageCode string) string {
	languageNames := map[string]string{
		"EN": "English", "FR": "Français", "IT": "Italiano", "ES": "Español", "DE": "Deutsch",
		"PT": "Português", "RU": "Русский", "JA": "日本語", "KO": "한국어", "ZH": "中文",
		"AR": "العربية", "NL": "Nederlands", "SV": "Svenska", "NO": "Norsk", "DA": "Dansk",
		"FI": "Suomi", "PL": "Polski", "CS": "Čeština", "HU": "Magyar", "RO": "Română",
		"BG": "Български", "HR": "Hrvatski", "SK": "Slovenčina", "SL": "Slovenščina",
		"ET": "Eesti", "LV": "Latviešu", "LT": "Lietuvių", "EL": "Ελληνικά", "TR": "Türkçe",
		"HE": "עברית", "HI": "हिन्दी", "TH": "ไทย", "VI": "Tiếng Việt", "ID": "Bahasa Indonesia",
		"MS": "Bahasa Melayu", "TL": "Filipino",
	}

	languageName := languageNames[strings.ToUpper(languageCode)]
	if languageName == "" {
		return languageCode
	}
	return languageName
}

// ValidateLanguageCode validates a language code
func ValidateLanguageCode(languageCode string) error {
	languageCode = strings.ToUpper(strings.TrimSpace(languageCode))
	if languageCode == "" {
		return fmt.Errorf("language code is required")
	}
	if len(languageCode) < 2 || len(languageCode) > 3 {
		return fmt.Errorf("language code should be 2-3 characters (e.g., FR, EN, IT)")
	}
	return nil
}

// GetSupportedLanguages returns a list of supported language codes
func GetSupportedLanguages() []string {
	return []string{
		"EN", "FR", "IT", "ES", "DE", "PT", "RU", "JA", "KO", "ZH",
		"AR", "NL", "SV", "NO", "DA", "FI", "PL", "CS", "HU", "RO",
		"BG", "HR", "SK", "SL", "ET", "LV", "LT", "EL", "TR", "HE",
		"HI", "TH", "VI", "ID", "MS", "TL",
	}
}

// CreatePromptFromGenerated creates a prompt model from generated text
func CreatePromptFromGenerated(template string, languageCode string) *models.Prompt {
	return &models.Prompt{
		Template: template,
		Tags:     []string{"generated", "llm-created", fmt.Sprintf("lang-%s", languageCode)},
		Enabled:  true,
	}
}

// ValidateGeneratedPrompt validates a generated prompt
func ValidateGeneratedPrompt(template string) error {
	if template == "" {
		return fmt.Errorf("prompt template cannot be empty")
	}
	if len(strings.TrimSpace(template)) == 0 {
		return fmt.Errorf("prompt template cannot be empty")
	}
	if len(template) > 10000 {
		return fmt.Errorf("prompt template too long (max 10000 characters)")
	}
	return nil
}
