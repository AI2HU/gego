package db

import (
	"database/sql"
)

// SQLBackend extends SQLDatabase with migration support.
type SQLBackend interface {
	SQLDatabase
	GetDB() *sql.DB
	DriverName() string
}
