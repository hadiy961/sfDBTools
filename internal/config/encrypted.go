package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"sfDBTools/internal/config/model"
	"sfDBTools/utils/crypto"
)

// EncryptedDatabaseConfig represents the encrypted database configuration
type EncryptedDatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// LoadEncryptedDatabaseConfig loads and decrypts the database configuration
func LoadEncryptedDatabaseConfig(cfg *model.Config, encryptionPassword string) (*EncryptedDatabaseConfig, error) {
	// Path to encrypted config file
	configPath := filepath.Join("./config", "database.encrypted")

	// Check if encrypted config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("encrypted database configuration not found at %s", configPath)
	}

	// Read encrypted data
	encryptedData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read encrypted config file: %w", err)
	}

	// Generate encryption key from app config and user password
	key, err := crypto.DeriveKeyWithPassword(
		cfg.General.AppName,
		cfg.General.ClientCode,
		cfg.General.Version,
		cfg.General.Author,
		encryptionPassword,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to derive decryption key: %w", err)
	}

	// Decrypt the data
	decryptedData, err := crypto.DecryptData(encryptedData, key, crypto.AES_GCM)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt database configuration: %w", err)
	}

	// Parse JSON
	var dbConfig EncryptedDatabaseConfig
	if err := json.Unmarshal(decryptedData, &dbConfig); err != nil {
		return nil, fmt.Errorf("failed to parse decrypted database configuration: %w", err)
	}

	return &dbConfig, nil
} // GetDatabaseConfigWithEncryption returns database configuration, preferring encrypted config if available
func GetDatabaseConfigWithEncryption() (host string, port int, user, password string, err error) {
	// Load main config
	cfg, err := Get()
	if err != nil {
		return "", 0, "", "", fmt.Errorf("failed to load main configuration: %w", err)
	}

	// Check if encrypted config exists
	encryptedConfigPath := filepath.Join("./config", "database.encrypted")
	if _, statErr := os.Stat(encryptedConfigPath); os.IsNotExist(statErr) {
		// If encrypted config is not available, fallback to plain config
		return cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, nil
	}

	// Try to load encrypted database config
	encryptionPassword, err := crypto.GetEncryptionPassword("Enter encryption password to decrypt database config: ")
	if err != nil {
		return "", 0, "", "", fmt.Errorf("failed to get encryption password: %w", err)
	}

	encryptedDB, err := LoadEncryptedDatabaseConfig(cfg, encryptionPassword)
	if err != nil {
		return "", 0, "", "", fmt.Errorf("failed to load encrypted database config: %w", err)
	}

	// Return encrypted configuration
	return encryptedDB.Host, encryptedDB.Port, encryptedDB.User, encryptedDB.Password, nil
}

// GetDatabaseConfigWithPassword returns database configuration using provided encryption password
func GetDatabaseConfigWithPassword(encryptionPassword string) (host string, port int, user, password string, err error) {
	// Load main config
	cfg, err := Get()
	if err != nil {
		return "", 0, "", "", fmt.Errorf("failed to load main configuration: %w", err)
	}

	// Check if encrypted config exists
	encryptedConfigPath := filepath.Join("./config", "database.encrypted")
	if _, statErr := os.Stat(encryptedConfigPath); os.IsNotExist(statErr) {
		// If encrypted config is not available, fallback to plain config
		return cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, nil
	}

	// Load encrypted database config with provided password
	encryptedDB, err := LoadEncryptedDatabaseConfig(cfg, encryptionPassword)
	if err != nil {
		return "", 0, "", "", fmt.Errorf("failed to load encrypted database config: %w", err)
	}

	// Return encrypted configuration
	return encryptedDB.Host, encryptedDB.Port, encryptedDB.User, encryptedDB.Password, nil
}

// ValidateEncryptedDatabaseConfig validates that the encrypted database configuration can be decrypted
func ValidateEncryptedDatabaseConfig(cfg *model.Config, encryptionPassword string) error {
	_, err := LoadEncryptedDatabaseConfig(cfg, encryptionPassword)
	return err
}

// LoadEncryptedDatabaseConfigFromFile loads and decrypts the database configuration from specific file
func LoadEncryptedDatabaseConfigFromFile(configPath string, cfg *model.Config, encryptionPassword string) (*EncryptedDatabaseConfig, error) {
	// Check if encrypted config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("encrypted database configuration not found at %s", configPath)
	}

	// Read encrypted data
	encryptedData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read encrypted config file: %w", err)
	}

	// Generate encryption key from app config and user password
	key, err := crypto.DeriveKeyWithPassword(
		cfg.General.AppName,
		cfg.General.ClientCode,
		cfg.General.Version,
		cfg.General.Author,
		encryptionPassword,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to derive decryption key: %w", err)
	}

	// Decrypt the data
	decryptedData, err := crypto.DecryptData(encryptedData, key, crypto.AES_GCM)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt database configuration: %w", err)
	}

	// Parse JSON
	var dbConfig EncryptedDatabaseConfig
	if err := json.Unmarshal(decryptedData, &dbConfig); err != nil {
		return nil, fmt.Errorf("failed to parse decrypted database configuration: %w", err)
	}

	return &dbConfig, nil
}
