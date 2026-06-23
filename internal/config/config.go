package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LLMProvider represents an LLM provider type
type LLMProvider int

const (
	ProviderGemini LLMProvider = iota
	ProviderChatGPT
	ProviderClaude
)

// String returns the string representation of the provider
func (p LLMProvider) String() string {
	switch p {
	case ProviderGemini:
		return "gemini"
	case ProviderChatGPT:
		return "chatgpt"
	case ProviderClaude:
		return "claude"
	default:
		return "unknown"
	}
}

// AuthConfig represents JWT authentication settings.
type AuthConfig struct {
	Issuer          string `yaml:"issuer"`
	Audience        string `yaml:"audience"`
	AccessTokenTTL  string `yaml:"access_token_ttl"`
	RefreshTokenTTL string `yaml:"refresh_token_ttl"`
}

// Config represents the application configuration
type Config struct {
	SQLDatabase              DatabaseConfig `yaml:"sql_database"`                         // SQLite for LLMs and Schedules
	NoSQLDatabase            DatabaseConfig `yaml:"nosql_database"`                       // MongoDB for Prompts and Responses
	CORSOrigin               string         `yaml:"cors_origin,omitempty"`                // CORS origin for API server
	Auth                     AuthConfig     `yaml:"auth,omitempty"`                       // JWT authentication settings
	KeywordsExclusionPath    string         `yaml:"keywords_exclusion_path,omitempty"`    // Path to keywords exclusion file
	GeminiSystemInstruction  string         `yaml:"gemini_system_instruction,omitempty"`  // System instruction for Gemini models
	ChatGPTSystemInstruction string         `yaml:"chatgpt_system_instruction,omitempty"` // System instruction for ChatGPT models
	ClaudeSystemInstruction  string         `yaml:"claude_system_instruction,omitempty"`  // System instruction for Claude models
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Provider string            `yaml:"provider"` // sqlite, mongodb, cassandra
	URI      string            `yaml:"uri"`
	Database string            `yaml:"database"`
	Options  map[string]string `yaml:"options,omitempty"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	configPath := GetConfigPath()
	configDir := filepath.Dir(configPath)
	return &Config{
		SQLDatabase: DatabaseConfig{
			Provider: "sqlite",
			URI:      "gego.db",
			Database: "gego",
		},
		NoSQLDatabase: DatabaseConfig{
			Provider: "mongodb",
			URI:      "mongodb://localhost:27017",
			Database: "gego",
		},
		CORSOrigin: "*",
		Auth: AuthConfig{
			Issuer:          "gego-api",
			Audience:        "gego-api",
			AccessTokenTTL:  "15m",
			RefreshTokenTTL: "168h",
		},
		KeywordsExclusionPath: filepath.Join(configDir, "keywords_exclusion"),
		GeminiSystemInstruction:  "You are Gemini, a helpful, concise, and friendly assistant. Respond conversationally and safely, the way the Gemini chat experience behaves. Focus on clear, direct answers that follow user intent.",
		ChatGPTSystemInstruction: "You are ChatGPT, a large language model trained by OpenAI. Follow the user's instructions carefully. Respond using markdown when appropriate.",
		ClaudeSystemInstruction:  "You are a helpful, harmless, and honest assistant.",
	}
}

// Load loads configuration from file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save saves configuration to file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the default config file path
func GetConfigPath() string {
	if envPath := os.Getenv("GEGO_CONFIG_PATH"); envPath != "" {
		return envPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".gego/config.yaml"
	}
	return filepath.Join(home, ".gego", "config.yaml")
}

// Exists checks if config file exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetSystemInstruction returns the system instruction for the given provider.
// If cfg is nil or the instruction is empty, returns the default instruction.
func GetSystemInstruction(cfg *Config, provider LLMProvider) string {
	var instruction string
	var defaultInstruction string

	switch provider {
	case ProviderGemini:
		if cfg != nil && cfg.GeminiSystemInstruction != "" {
			instruction = cfg.GeminiSystemInstruction
		}
		defaultCfg := DefaultConfig()
		defaultInstruction = defaultCfg.GeminiSystemInstruction
	case ProviderChatGPT:
		if cfg != nil && cfg.ChatGPTSystemInstruction != "" {
			instruction = cfg.ChatGPTSystemInstruction
		}
		defaultCfg := DefaultConfig()
		defaultInstruction = defaultCfg.ChatGPTSystemInstruction
	case ProviderClaude:
		if cfg != nil && cfg.ClaudeSystemInstruction != "" {
			instruction = cfg.ClaudeSystemInstruction
		}
		defaultCfg := DefaultConfig()
		defaultInstruction = defaultCfg.ClaudeSystemInstruction
	default:
		return ""
	}

	if instruction != "" {
		return instruction
	}
	return defaultInstruction
}
