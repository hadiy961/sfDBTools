package generate

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/crypto"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"
)

// EncryptedDatabaseConfig represents the encrypted database configuration
type EncryptedDatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// ProcessGenerate handles the core generate operation logic
func ProcessGenerate(cfg *dbconfig.Config) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting encrypted database configuration generation")

	// Check if we should use automated mode
	useAutoMode := cfg.AutoMode && cfg.ConfigName != "" && cfg.Host != "" && cfg.Port != 0 && cfg.User != ""

	if useAutoMode {
		return processAutoMode(cfg)
	}

	return processInteractiveMode()
}

// processAutoMode handles automated generation using provided parameters
func processAutoMode(cfg *dbconfig.Config) error {
	terminal.PrintInfo("🤖 Automated Mode - Using provided parameters")

	// Use provided config name
	finalConfigName := cfg.ConfigName

	// Validate filename (remove invalid characters)
	finalConfigName = strings.ReplaceAll(finalConfigName, " ", "_")
	finalConfigName = strings.ReplaceAll(finalConfigName, "/", "_")
	finalConfigName = strings.ReplaceAll(finalConfigName, "\\", "_")

	// Validate parameters
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("port number must be between 1 and 65535")
	}

	// Get encryption password from environment or prompt
	finalEncryptionPassword, err := crypto.GetEncryptionPassword("Enter encryption password: ")
	if err != nil {
		return fmt.Errorf("failed to get encryption password: %w", err)
	}

	// Get database password from environment or prompt
	dbPassword, err := crypto.GetDatabasePassword("Enter database password: ")
	if err != nil {
		return fmt.Errorf("failed to get database password: %w", err)
	}

	// Create database config from flags and prompted passwords
	dbConfig := &EncryptedDatabaseConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: dbPassword,
	}

	// Display summary
	dbconfig.DisplayConfigSummary(finalConfigName, &config.EncryptedDatabaseConfig{
		Host:     dbConfig.Host,
		Port:     dbConfig.Port,
		User:     dbConfig.User,
		Password: dbConfig.Password,
	})

	return saveConfiguration(finalConfigName, dbConfig, finalEncryptionPassword)
}

// processInteractiveMode handles interactive generation
func processInteractiveMode() error {
	terminal.PrintSubHeader("🔐 Encryption Setup")
	terminal.PrintInfo("The database configuration will be encrypted using:")
	terminal.PrintInfo("   1. encryption password (from environment or user input)")
	terminal.PrintInfo(fmt.Sprintf("   Environment variable: %s", crypto.ENV_ENCRYPTION_PASSWORD))
	fmt.Println()

	finalEncryptionPassword, err := crypto.ConfirmEncryptionPassword("Enter encryption password: ")
	if err != nil {
		return fmt.Errorf("failed to get encryption password: %w", err)
	}

	// Prompt for configuration name
	finalConfigName, err := promptConfigName()
	if err != nil {
		return fmt.Errorf("failed to get configuration name: %w", err)
	}

	// Prompt for database configuration
	dbConfig, err := promptDatabaseConfig()
	if err != nil {
		return fmt.Errorf("failed to get database configuration: %w", err)
	}

	return saveConfiguration(finalConfigName, dbConfig, finalEncryptionPassword)
}

// promptConfigName prompts the user for configuration name
func promptConfigName() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	terminal.PrintSubHeader("📁 Configuration File Name")
	fmt.Print("Enter configuration name (without extension) [database]: ")

	name, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read configuration name: %w", err)
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = "database"
	}

	// Validate filename (remove invalid characters)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")

	terminal.PrintInfo(fmt.Sprintf("Configuration will be saved as: %s.cnf.enc", name))

	return name, nil
}

// promptDatabaseConfig prompts the user for database configuration
func promptDatabaseConfig() (*EncryptedDatabaseConfig, error) {
	reader := bufio.NewReader(os.Stdin)
	dbConfig := &EncryptedDatabaseConfig{}

	terminal.PrintSubHeader("📋 Database Configuration")

	// Prompt for host
	fmt.Print("Enter database host [localhost]: ")
	host, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read host: %w", err)
	}
	host = strings.TrimSpace(host)
	if host == "" {
		host = "localhost"
	}
	dbConfig.Host = host

	// Prompt for port
	fmt.Print("Enter database port [3306]: ")
	portStr, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read port: %w", err)
	}
	portStr = strings.TrimSpace(portStr)
	if portStr == "" {
		dbConfig.Port = 3306
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port number: %w", err)
		}
		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("port number must be between 1 and 65535")
		}
		dbConfig.Port = port
	}

	// Prompt for username
	fmt.Print("Enter database username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read username: %w", err)
	}
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	dbConfig.User = username

	// Get database password from environment variable or prompt
	password, err := crypto.GetDatabasePassword("Enter database password: ")
	if err != nil {
		return nil, fmt.Errorf("failed to get database password: %w", err)
	}
	dbConfig.Password = password

	// Show summary and confirm
	dbconfig.DisplayConfigSummary("config", &config.EncryptedDatabaseConfig{
		Host:     dbConfig.Host,
		Port:     dbConfig.Port,
		User:     dbConfig.User,
		Password: dbConfig.Password,
	})

	fmt.Print("\nSave this configuration? [Y/n]: ")
	confirm, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read confirmation: %w", err)
	}
	confirm = strings.ToLower(strings.TrimSpace(confirm))

	if confirm != "" && confirm != "y" && confirm != "yes" {
		return nil, fmt.Errorf("configuration generation cancelled")
	}

	return dbConfig, nil
}

// saveConfiguration saves the configuration to encrypted file
func saveConfiguration(configName string, dbConfig *EncryptedDatabaseConfig, encryptionPassword string) error {
	lg, _ := logger.Get()

	// Generate encryption key from user password only
	key, err := crypto.DeriveKeyWithPassword(encryptionPassword)
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
	configDir, err := config.GetDatabaseConfigDirectory()
	if err != nil {
		return fmt.Errorf("failed to get database config directory: %w", err)
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	encryptedConfigPath := filepath.Join(configDir, configName+".cnf.enc")
	if err := os.WriteFile(encryptedConfigPath, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write encrypted config file: %w", err)
	}

	lg.Info("Encrypted database configuration generated successfully",
		logger.String("file", encryptedConfigPath))

	terminal.PrintSuccess(fmt.Sprintf("✅ Encrypted database configuration saved to: %s", encryptedConfigPath))
	terminal.PrintSuccess("🔐 Configuration encrypted using encryption password")

	return nil
}
