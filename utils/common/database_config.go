package common

import (
	"fmt"

	"sfDBTools/internal/config"
	"sfDBTools/utils/crypto"
	"sfDBTools/utils/database"
)

// GetDatabaseConfigFromDefault gets database configuration from the main config.yaml
func GetDatabaseConfigFromDefault() (*database.Config, error) {
	// Load the main application configuration
	cfg, err := config.Get()
	if err != nil {
		return nil, err
	}

	// Create database config from main config
	dbConfig := &database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   "", // Don't specify database for health check
	}

	return dbConfig, nil
}

// LoadDatabaseConfig loads database configuration from encrypted file
func LoadDatabaseConfig(configFilePath string) (*config.EncryptedDatabaseConfig, error) {
	// Get encryption password
	encryptionPassword, err := crypto.GetEncryptionPassword("Enter encryption password: ")
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption password: %w", err)
	}

	// Load and decrypt the configuration
	return LoadEncryptedConfigFromFile(configFilePath, encryptionPassword)
}
