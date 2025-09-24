package connection

import (
	"database/sql"
	"fmt"
	"sfDBTools/internal/logger"
)

// Get returns a connection to the specific database
func Get(config Config) (*sql.DB, error) {
	lg, err := getLogger()
	if err != nil {
		return nil, err
	}

	dsn := buildDSN(config, true) // Connect with database selected

	db, err := createConnection(dsn)
	if err != nil {
		lg.Error("Failed to open database connection", logger.Error(err))
		return nil, fmt.Errorf("failed to open connection to database '%s': %w", config.DBName, err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		lg.Error("Failed to connect to database",
			logger.Error(err),
			logger.String("database", config.DBName))
		return nil, fmt.Errorf("failed to connect to database '%s': %w", config.DBName, err)
	}

	// success - connection established (no verbose debug log to avoid noisy output)
	return db, nil
}

// GetWithoutDB returns a connection without selecting a specific database
func GetWithoutDB(config Config) (*sql.DB, error) {
	lg, err := getLogger()
	if err != nil {
		return nil, err
	}

	dsn := buildDSN(config, false) // Connect without database selected

	db, err := createConnection(dsn)
	if err != nil {
		lg.Error("Failed to open database connection", logger.Error(err))
		return nil, fmt.Errorf("failed to open connection to database server: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		lg.Error("Failed to connect to database server", logger.Error(err))
		return nil, fmt.Errorf("failed to connect to database server: %w", err)
	}

	// success - connection to server established (no verbose debug log to avoid noisy output)
	return db, nil
}
