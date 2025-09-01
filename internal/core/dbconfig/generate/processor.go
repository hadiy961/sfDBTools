package generate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sfDBTools/internal/config"
	coredbconfig "sfDBTools/internal/core/dbconfig"
	"sfDBTools/utils/crypto"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"
)

// Processor handles generate operations for database configurations
type Processor struct {
	*coredbconfig.BaseProcessor
	configHelper *coredbconfig.ConfigHelper
}

// NewProcessor creates a new generate processor
func NewProcessor() (*Processor, error) {
	base, err := coredbconfig.NewBaseProcessor()
	if err != nil {
		return nil, err
	}

	configHelper, err := coredbconfig.NewConfigHelper()
	if err != nil {
		return nil, err
	}

	return &Processor{
		BaseProcessor: base,
		configHelper:  configHelper,
	}, nil
}

// ProcessGenerate handles the core generate operation logic
func ProcessGenerate(cfg *dbconfig.Config) error {
	processor, err := NewProcessor()
	if err != nil {
		return err
	}

	processor.LogOperation("database configuration generation", "")

	// Check if we should use automated mode
	useAutoMode := cfg.AutoMode && cfg.ConfigName != "" && cfg.Host != "" && cfg.Port != 0 && cfg.User != ""

	if useAutoMode {
		return processor.processAutoMode(cfg)
	}

	return processor.processInteractiveMode()
}

// processAutoMode handles automated generation using provided parameters
func (p *Processor) processAutoMode(cfg *dbconfig.Config) error {
	terminal.PrintInfo("🤖 Automated Mode - Using provided parameters")

	// Create configuration from provided parameters
	dbConfig := &config.EncryptedDatabaseConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
	}

	// Generate final config name
	finalConfigName := cfg.ConfigName

	// Display summary
	configInfo := &dbconfig.ConfigInfo{
		Name:         finalConfigName,
		Host:         dbConfig.Host,
		Port:         dbConfig.Port,
		User:         dbConfig.User,
		HasPassword:  dbConfig.Password != "",
		FileSize:     "New file",
		LastModified: time.Now(),
		IsValid:      true,
	}

	dbconfig.DisplayConfigSummary([]*dbconfig.ConfigInfo{configInfo})

	// Get encryption password
	encryptionPassword, err := p.GetEncryptionPassword("encrypt the configuration")
	if err != nil {
		return err
	}

	// Save configuration
	return p.saveEncryptedConfig(finalConfigName, dbConfig, encryptionPassword)
}

// processInteractiveMode handles interactive generation
func (p *Processor) processInteractiveMode() error {
	terminal.PrintSubHeader("🛠️ Interactive Configuration Generator")
	terminal.PrintInfo("Please provide database connection details:")

	// Get configuration details from user
	inputConfig, err := dbconfig.PromptDatabaseConfig()
	if err != nil {
		return fmt.Errorf("error getting database configuration: %v", err)
	}

	// Get configuration name
	configName, err := dbconfig.PromptConfigName("")
	if err != nil {
		return fmt.Errorf("error getting configuration name: %v", err)
	}

	// Get password handling option
	passwordOption, err := dbconfig.DisplayPasswordOption()
	if err != nil {
		return fmt.Errorf("error getting password option: %v", err)
	}

	// Handle password based on option
	var password string
	switch passwordOption {
	case "manual":
		password = terminal.AskString("Enter database password", "")
		if password == "" {
			return fmt.Errorf("password cannot be empty - please provide a valid password")
		}
	case "env":
		envVar := terminal.AskString("Environment variable name", "DB_PASSWORD")
		if envVar == "" {
			return fmt.Errorf("environment variable name cannot be empty")
		}
		password = fmt.Sprintf("${%s}", envVar)
	default:
		return fmt.Errorf("invalid password option: %s", passwordOption)
	}

	// Create final configuration
	dbConfig := &config.EncryptedDatabaseConfig{
		Host:     inputConfig.Host,
		Port:     inputConfig.Port,
		User:     inputConfig.User,
		Password: password,
	}

	// Display summary
	configInfo := &dbconfig.ConfigInfo{
		Name:         configName,
		Host:         dbConfig.Host,
		Port:         dbConfig.Port,
		User:         dbConfig.User,
		HasPassword:  dbConfig.Password != "",
		FileSize:     "New file",
		LastModified: time.Now(),
		IsValid:      true,
	}

	dbconfig.DisplayConfigSummary([]*dbconfig.ConfigInfo{configInfo})

	// Confirm save
	if !terminal.AskYesNo("Save this configuration?", true) {
		terminal.PrintWarning("❌ Configuration not saved.")
		return nil
	}

	// Get encryption password
	encryptionPassword, err := p.GetEncryptionPassword("encrypt the configuration")
	if err != nil {
		return err
	}

	// Save configuration
	return p.saveEncryptedConfig(configName, dbConfig, encryptionPassword)
}

// saveEncryptedConfig saves the configuration to an encrypted file
func (p *Processor) saveEncryptedConfig(configName string, dbConfig *config.EncryptedDatabaseConfig, encryptionPassword string) error {
	// Get config directory
	configDir, err := config.GetDatabaseConfigDirectory()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %v", err)
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Get full file path
	filePath := filepath.Join(configDir, configName+".cnf.enc")

	// Check if file exists
	if _, err := os.Stat(filePath); err == nil {
		if !dbconfig.ConfirmOverwrite(configName) {
			terminal.PrintWarning("❌ Configuration not saved.")
			return nil
		}

		// Create backup of existing file
		if _, err := p.configHelper.BackupConfigFile(filePath); err != nil {
			terminal.PrintWarning(fmt.Sprintf("⚠️ Could not create backup: %v", err))
		}
	}

	// Convert configuration to JSON
	configJSON, err := json.Marshal(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %v", err)
	}

	// Encrypt and save
	spinner := terminal.NewProgressSpinner("Encrypting and saving configuration...")
	spinner.Start()

	// Generate encryption key from user password only
	key, err := crypto.DeriveKeyWithPassword(encryptionPassword)
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to derive encryption key: %v", err)
	}

	// Encrypt the configuration
	encryptedData, err := crypto.EncryptData(configJSON, key, crypto.AES_GCM)
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to encrypt configuration: %v", err)
	}

	// Save encrypted data to file
	err = os.WriteFile(filePath, encryptedData, 0600)
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("failed to save configuration file: %v", err)
	}

	terminal.PrintSuccess(fmt.Sprintf("✅ Configuration '%s' saved successfully!", configName))
	terminal.PrintInfo(fmt.Sprintf("📁 Saved to: %s", filePath))

	return nil
}
