package database

import (
	"fmt"
	"sfDBTools/internal/logger"
)

// getMySQLVersion gets the MySQL server version
func GetMySQLVersion(config Config) (string, error) {
	db, err := GetDatabaseConnection(config)
	if err != nil {
		return "", err
	}
	defer db.Close()

	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	return version, err
}

// validateConnection validates the database connection and user privileges
func ValidateBeforeAction(config Config) error {
	lg, _ := logger.Get()

	// Validate basic connection
	if err := ValidateConnection(config); err != nil {
		lg.Error("Connection validation failed", logger.Error(err))
		return fmt.Errorf("connection validation failed: %w", err)
	}

	// Validate user privileges
	if err := ValidateUser(config); err != nil {
		lg.Error("User validation failed", logger.Error(err))
		return fmt.Errorf("user validation failed: %w", err)
	}

	// Validate database exists
	if err := ValidateDatabase(config); err != nil {
		lg.Error("Database validation failed", logger.Error(err))
		return fmt.Errorf("database validation failed: %w", err)
	}

	return nil
}
