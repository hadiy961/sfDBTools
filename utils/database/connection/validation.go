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
