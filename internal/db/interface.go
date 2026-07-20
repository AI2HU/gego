package db

// Database defines the combined interface for both SQL and NoSQL database operations.
type Database interface {
	SQLDatabase
	ExclusionWordDatabase
	BrandDatabase
	SMTPSettingsDatabase
	PasswordInviteDatabase
	NoSQLDatabase
}
