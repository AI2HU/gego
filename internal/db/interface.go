package db

// Database defines the combined interface for both SQL and NoSQL database operations
// This interface combines SQLDatabase and NoSQLDatabase for backward compatibility
type Database interface {
	SQLDatabase
	NoSQLDatabase
}
