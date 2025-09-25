package connection

import (
	"fmt"
	"time"

	"sfDBTools/internal/logger"
)

// ValidateConnection checks if the database connection is valid
func ValidateConnection(config Config) error {
	lg, err := getLogger()
	if err != nil {
		return err
	}

	dsn := buildDSN(config, false) // No need to specify database for connection validation
	lg.Debug("Validating database connection",
		logger.String("host", config.Host),
		logger.Int("port", config.Port))

	db, err := createConnection(dsn)
	if err != nil {
		lg.Error("Failed to open database connection", logger.Error(err))
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	// Override connection timeout for validation
	db.SetConnMaxLifetime(time.Second * 10)

	// Try to connect
	if err := db.Ping(); err != nil {
		lg.Error("Failed to connect to database", logger.Error(err))
		return fmt.Errorf("failed to connect to database server: %w", err)
	}

	lg.Debug("Database connection is valid",
		logger.String("host", config.Host),
		logger.Int("port", config.Port))
	return nil
}

// ValidateDatabase checks if the specified database exists
func ValidateDatabase(config Config) error {
	lg, err := getLogger()
	if err != nil {
		return err
	}

	dsn := buildDSN(config, false) // Connect without selecting a database
	lg.Debug("Validating database exists",
		logger.String("database", config.DBName))

	db, err := createConnection(dsn)
	if err != nil {
		lg.Error("Failed to open database connection", logger.Error(err))
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	// Check if database exists
	var exists bool
	query := "SELECT COUNT(*) > 0 FROM information_schema.schemata WHERE schema_name = ?"
	err = db.QueryRow(query, config.DBName).Scan(&exists)
	if err != nil {
		lg.Error("Failed to check if database exists",
			logger.Error(err),
			logger.String("database", config.DBName))
		return fmt.Errorf("failed to check if database '%s' exists: %w", config.DBName, err)
	}

	if !exists {
		lg.Error("Database does not exist",
			logger.String("database", config.DBName))
		return fmt.Errorf("database '%s' does not exist", config.DBName)
	}

	lg.Debug("Database exists",
		logger.String("database", config.DBName))
	return nil
}
