package database

import (
	"database/sql"

	"sfDBTools/utils/database/connection"
)

// Config is exported for backward compatibility
type Config = connection.Config

// ValidateConnection checks if the database connection is valid
func ValidateConnection(config Config) error {
	return connection.ValidateConnection(config)
}

// ValidateUser checks if the user has sufficient privileges
func ValidateUser(config Config) error {
	return connection.ValidateUser(config)
}

// ValidateDatabase checks if the specified database exists
func ValidateDatabase(config Config) error {
	return connection.ValidateDatabase(config)
}

// GetDatabaseConnection returns a connection to the specific database
func GetDatabaseConnection(config Config) (*sql.DB, error) {
	return connection.Get(config)
}

// GetWithoutDB returns a connection without selecting a specific database
func GetWithoutDB(config Config) (*sql.DB, error) {
	return connection.GetWithoutDB(config)
}

// EnsureDatabase checks that a database exists, creating it if necessary
func EnsureDatabase(config Config) error {
	return connection.EnsureDatabase(config)
}
