package generate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/crypto"

	"github.com/spf13/cobra"
)

// GenerateEncryptedConfig generates encrypted database configuration
func GenerateEncryptedConfig(cmd *cobra.Command, configName, dbHost string, dbPort int, dbUser string, autoMode bool) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	// Load current config to get general settings
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	lg.Info("Starting encrypted database configuration generation")

	// Check if we should use automated mode (based on --auto flag and availability of required parameters)
	useAutoMode := autoMode && configName != "" && dbHost != "" && dbPort != 0 && dbUser != ""

	var finalEncryptionPassword string
	var finalConfigName string
	var dbConfig *EncryptedDatabaseConfig

	if useAutoMode {
		// Automated mode - use flags + environment variables or prompts for passwords
		fmt.Println("🤖 Automated Mode - Using provided parameters")

		// Use provided config name
		finalConfigName = configName

		// Validate filename (remove invalid characters)
		finalConfigName = strings.ReplaceAll(finalConfigName, " ", "_")
		finalConfigName = strings.ReplaceAll(finalConfigName, "/", "_")
		finalConfigName = strings.ReplaceAll(finalConfigName, "\\", "_")

		// Validate parameters
		if dbPort < 1 || dbPort > 65535 {
			return fmt.Errorf("port number must be between 1 and 65535")
		}

		// Get encryption password from environment or prompt
		finalEncryptionPassword, err = crypto.GetEncryptionPassword("Enter encryption password: ")
		if err != nil {
			return fmt.Errorf("failed to get encryption password: %w", err)
		}

		// Get database password from environment or prompt
		dbPassword, err := crypto.GetDatabasePassword("Enter database password: ")
		if err != nil {
			return fmt.Errorf("failed to get database password: %w", err)
		}

		// Create database config from flags and prompted passwords
		dbConfig = &EncryptedDatabaseConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
		}

		fmt.Printf("📋 Configuration: %s.cnf.enc\n", finalConfigName)
		fmt.Printf("   Host: %s:%d\n", dbConfig.Host, dbConfig.Port)
		fmt.Printf("   User: %s\n", dbConfig.User)
		fmt.Printf("   Password: %s\n", strings.Repeat("*", len(dbConfig.Password)))

	} else {
		// Interactive mode
		fmt.Println("🔐 Encryption Setup")
		fmt.Println("The database configuration will be encrypted using:")
		fmt.Println("   1. Application configuration (app_name, client_code, version, author)")
		fmt.Println("   2. Additional encryption password (from environment or user input)")
		fmt.Printf("   Environment variable: %s\n\n", crypto.ENV_ENCRYPTION_PASSWORD)

		finalEncryptionPassword, err = crypto.ConfirmEncryptionPassword("Enter encryption password: ")
		if err != nil {
			return fmt.Errorf("failed to get encryption password: %w", err)
		}

		// Prompt for configuration name
		finalConfigName, err = PromptConfigName()
		if err != nil {
			return fmt.Errorf("failed to get configuration name: %w", err)
		}

		// Prompt for database configuration
		dbConfig, err = PromptDatabaseConfig()
		if err != nil {
			return fmt.Errorf("failed to get database configuration: %w", err)
		}
	}

	// Generate encryption key from app config and user password
	key, err := crypto.DeriveKeyWithPassword(
		cfg.General.AppName,
		cfg.General.ClientCode,
		cfg.General.Version,
		cfg.General.Author,
		finalEncryptionPassword,
	)
	if err != nil {
		return fmt.Errorf("failed to derive encryption key: %w", err)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal database config to JSON: %w", err)
	}

	// Encrypt the configuration
	encryptedData, err := crypto.EncryptData(jsonData, key, crypto.AES_GCM)
	if err != nil {
		return fmt.Errorf("failed to encrypt database configuration: %w", err)
	}

	// Save encrypted configuration
	configDir := config.GetDatabaseConfigDirectory()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	encryptedConfigPath := filepath.Join(configDir, finalConfigName+".cnf.enc")
	if err := os.WriteFile(encryptedConfigPath, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write encrypted config file: %w", err)
	}

	lg.Info("Encrypted database configuration generated successfully",
		logger.String("file", encryptedConfigPath))

	fmt.Printf("✅ Encrypted database configuration saved to: %s\n", encryptedConfigPath)
	fmt.Println("🔐 Configuration encrypted using application settings + encryption password")

	return nil
}
