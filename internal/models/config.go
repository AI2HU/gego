package models

// Configuration models

// Config holds database configuration
type Config struct {
	Provider string            // sqlite, mongodb, cassandra
	URI      string            // Connection URI
	Database string            // Database name
	Options  map[string]string // Provider-specific options
}
