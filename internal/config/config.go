package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	SQLDatabase   DatabaseConfig `yaml:"sql_database"`   // SQLite for LLMs and Schedules
	NoSQLDatabase DatabaseConfig `yaml:"nosql_database"` // MongoDB for Prompts and Responses
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

	// Create directory if it doesn't exist
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
