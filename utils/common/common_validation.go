package common

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
)

// isRemoteConnection checks if the connection is remote (not localhost)
func IsRemoteConnection(host string) bool {
	return host != "localhost" && host != "127.0.0.1" && host != "::1" && host != ""
}

// calculateChecksum calculates SHA256 checksum of the backup file
func CalculateChecksum(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// validateConnection validates the database connection and user privileges
func validateConnection(config database.Config) error {
	lg, _ := logger.Get()

	// Validate basic connection
	if err := database.ValidateConnection(config); err != nil {
		lg.Error("Connection validation failed", logger.Error(err))
		return fmt.Errorf("connection validation failed: %w", err)
	}

	// Validate user privileges
	if err := database.ValidateUser(config); err != nil {
		lg.Error("User validation failed", logger.Error(err))
		return fmt.Errorf("user validation failed: %w", err)
	}

	// Validate database exists
	if err := database.ValidateDatabase(config); err != nil {
		lg.Error("Database validation failed", logger.Error(err))
		return fmt.Errorf("database validation failed: %w", err)
	}

	return nil
}
