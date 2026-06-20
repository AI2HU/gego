package services

import (
	"fmt"

	"github.com/AI2HU/gego/internal/config"
	"github.com/AI2HU/gego/internal/llm"
	"github.com/AI2HU/gego/internal/llm/anthropic"
	"github.com/AI2HU/gego/internal/llm/google"
	"github.com/AI2HU/gego/internal/llm/ollama"
	"github.com/AI2HU/gego/internal/llm/openai"
	"github.com/AI2HU/gego/internal/llm/perplexity"
	"github.com/AI2HU/gego/internal/models"
)

func NewLLMRegistryForConfig(llmConfig *models.LLMConfig, appCfg *config.Config) (*llm.Registry, error) {
	registry := llm.NewRegistry()

	provider, err := providerForLLMConfig(llmConfig, appCfg)
	if err != nil {
		return nil, err
	}

	registry.Register(provider)
	return registry, nil
}

func providerForLLMConfig(llmConfig *models.LLMConfig, appCfg *config.Config) (llm.Provider, error) {
	geminiSystemInstruction := config.GetSystemInstruction(appCfg, config.ProviderGemini)
	chatGPTSystemInstruction := config.GetSystemInstruction(appCfg, config.ProviderChatGPT)

	switch llmConfig.Provider {
	case "openai":
		return openai.New(llmConfig.APIKey, llmConfig.BaseURL, chatGPTSystemInstruction), nil
	case "anthropic":
		return anthropic.New(llmConfig.APIKey, llmConfig.BaseURL), nil
	case "ollama":
		return ollama.New(llmConfig.BaseURL), nil
	case "google":
		return google.New(llmConfig.APIKey, llmConfig.BaseURL, geminiSystemInstruction), nil
	case "perplexity":
		return perplexity.New(llmConfig.APIKey, llmConfig.BaseURL), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", llmConfig.Provider)
	}
}
