package edit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sfDBTools/internal/config"
	coredbconfig "sfDBTools/internal/core/dbconfig"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/crypto"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"
)

// Processor handles edit operations for database configurations
type Processor struct {
	*coredbconfig.BaseProcessor
	configHelper *coredbconfig.ConfigHelper
}

// NewProcessor creates a new edit processor
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

// ProcessEdit handles the core edit operation logic
func ProcessEdit(cfg *dbconfig.Config, Lg *logger.Logger) error {
	processor, err := NewProcessor()
	if err != nil {
		return err
	}

	processor.LogOperation("database configuration editing", cfg.FilePath)
	return processor.editSpecificConfig(cfg.FilePath)
}

// editSpecificConfig edits a specific configuration file
func (p *Processor) editSpecificConfig(filePath string) error {
	// Validate config file exists
	if err := p.configHelper.ValidateConfigExists(filePath); err != nil {
		return err
	}

	// Get encryption password
	encryptionPassword, err := p.GetEncryptionPassword("decrypt the configuration")
	if err != nil {
		return err
	}

	// Load and decrypt the configuration
	dbConfig, err := p.configHelper.LoadDecryptedConfig(filePath, encryptionPassword)
	if err != nil {
		return err
	}

	// Display current configuration
	p.displayCurrentConfig(filePath, dbConfig)

	// Get current name from file path
	currentName := p.extractConfigName(filePath)

	// Prompt for new values
	updatedConfig, newName, hasChanges, err := p.promptForUpdates(dbConfig, currentName)
	if err != nil {
		return err
	}
	// Check if any changes were made
	if !hasChanges {
		terminal.PrintInfo("No changes made.")
		terminal.PrintSuccess("Configuration editing completed (no changes).")
		return nil
	}

	// Confirm changes
	if !terminal.AskYesNo("Save these changes?", true) {
		terminal.PrintWarning("Changes cancelled.")
		return nil
	}

	// Save the updated configuration
	return p.saveUpdatedConfig(filePath, currentName, newName, updatedConfig, encryptionPassword)
}

// displayCurrentConfig shows the current configuration
func (p *Processor) displayCurrentConfig(filePath string, dbConfig *config.EncryptedDatabaseConfig) {
	terminal.PrintSubHeader("Current Configuration:")
	terminal.PrintInfo(fmt.Sprintf("📁 Source: %s", filePath))
	terminal.PrintInfo(fmt.Sprintf("   Host: %s", dbConfig.Host))
	terminal.PrintInfo(fmt.Sprintf("   Port: %d", dbConfig.Port))
	terminal.PrintInfo(fmt.Sprintf("   User: %s", dbConfig.User))
	terminal.PrintInfo(fmt.Sprintf("   Password: %s", dbConfig.Password))
}

// extractConfigName extracts configuration name from file path
func (p *Processor) extractConfigName(filePath string) string {
	currentName := strings.TrimSuffix(strings.TrimSuffix(filePath, ".cnf.enc"), "config/")
	if strings.Contains(currentName, "/") {
		parts := strings.Split(currentName, "/")
		currentName = parts[len(parts)-1]
	}
	return currentName
}

// promptForUpdates prompts user for configuration updates
func (p *Processor) promptForUpdates(dbConfig *config.EncryptedDatabaseConfig, currentName string) (*config.EncryptedDatabaseConfig, string, bool, error) {
	terminal.PrintSubHeader("Edit Configuration:")
	terminal.PrintInfo("Press Enter to keep current value, or type new value:")

	// Edit configuration name
	newName := terminal.AskString("Configuration name", currentName)

	// Edit host
	newHost := terminal.AskString("Host", dbConfig.Host)

	// Edit port
	newPort := p.promptForPort(dbConfig.Port)

	// Edit user
	newUser := terminal.AskString("User", dbConfig.User)

	// Edit password
	newPassword := terminal.AskString("Password", dbConfig.Password)

	// Create updated configuration
	updatedConfig := &config.EncryptedDatabaseConfig{
		Host:     newHost,
		Port:     newPort,
		User:     newUser,
		Password: newPassword,
	}

	terminal.Headers("Edit Database Configuration")
	// Display changes summary and check for changes
	hasChanges := p.displayChangesSummary(dbConfig, updatedConfig, currentName, newName)

	return updatedConfig, newName, hasChanges, nil
}

// promptForPort prompts for port with validation
func (p *Processor) promptForPort(currentPort int) int {
	portStr := terminal.AskString("Port", fmt.Sprintf("%d", currentPort))

	if portStr == fmt.Sprintf("%d", currentPort) {
		return currentPort
	}

	if port, err := strconv.Atoi(portStr); err == nil {
		if port >= 1 && port <= 65535 {
			return port
		} else {
			terminal.PrintWarning(fmt.Sprintf("❌ Invalid port number: %d. Using current value: %d", port, currentPort))
		}
	} else {
		terminal.PrintWarning(fmt.Sprintf("❌ Invalid port format. Using current value: %d", currentPort))
	}

	return currentPort
}

// displayChangesSummary shows what will be changed
func (p *Processor) displayChangesSummary(oldConfig, newConfig *config.EncryptedDatabaseConfig, oldName, newName string) bool {
	fmt.Println("\n📋 Changes Summary:")
	fmt.Println("===================")

	hasChanges := false

	if newName != oldName {
		fmt.Printf("   Name: %s → %s\n", oldName, newName)
		hasChanges = true
	}
	if newConfig.Host != oldConfig.Host {
		fmt.Printf("   Host: %s → %s\n", oldConfig.Host, newConfig.Host)
		hasChanges = true
	}
	if newConfig.Port != oldConfig.Port {
		fmt.Printf("   Port: %d → %d\n", oldConfig.Port, newConfig.Port)
		hasChanges = true
	}
	if newConfig.User != oldConfig.User {
		fmt.Printf("   User: %s → %s\n", oldConfig.User, newConfig.User)
		hasChanges = true
	}
	if newConfig.Password != oldConfig.Password {
		fmt.Printf("   Password: %s → %s\n", p.maskPassword(oldConfig.Password), p.maskPassword(newConfig.Password))
		hasChanges = true
	}

	return hasChanges
}

// maskPassword creates a masked version of password for display
func (p *Processor) maskPassword(password string) string {
	if password == "" {
		return "[empty]"
	}
	return strings.Repeat("*", len(password))
}

// saveUpdatedConfig saves the updated configuration to file
func (p *Processor) saveUpdatedConfig(originalPath, currentName, newName string, dbConfig *config.EncryptedDatabaseConfig, encryptionPassword string) error {
	// Generate encryption key from user password only
	key, err := crypto.DeriveKeyWithPassword(encryptionPassword)
	if err != nil {
		return fmt.Errorf("failed to derive encryption key: %v", err)
	}

	// Convert to JSON
	jsonData, err := p.marshalConfig(dbConfig)
	if err != nil {
		return err
	}

	// Encrypt the configuration
	encryptedData, err := crypto.EncryptData(jsonData, key, crypto.AES_GCM)
	if err != nil {
		return fmt.Errorf("failed to encrypt database configuration: %v", err)
	}

	// Determine new file path
	configDir, err := config.GetDatabaseConfigDirectory()
	if err != nil {
		return fmt.Errorf("failed to get database config directory: %v", err)
	}

	newFileName := newName + ".cnf.enc"
	newFilePath := filepath.Join(configDir, newFileName)

	// Handle file rename/save
	if newName != currentName {
		return p.saveWithRename(originalPath, newFilePath, encryptedData, newName)
	} else {
		return p.saveInPlace(originalPath, encryptedData)
	}
}

// marshalConfig converts config to JSON
func (p *Processor) marshalConfig(dbConfig *config.EncryptedDatabaseConfig) ([]byte, error) {
	jsonData, err := json.Marshal(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal database config to JSON: %v", err)
	}
	return jsonData, nil
}

// saveWithRename saves to new file and removes old one
func (p *Processor) saveWithRename(originalPath, newFilePath string, encryptedData []byte, newName string) error {
	// Check if new file already exists
	if _, err := os.Stat(newFilePath); err == nil {
		return fmt.Errorf("configuration file with name '%s' already exists", newName)
	}

	// Write to new file
	if err := os.WriteFile(newFilePath, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write new encrypted config file: %v", err)
	}

	// Remove old file
	if err := os.Remove(originalPath); err != nil {
		// If we can't remove old file, warn but don't fail
		terminal.PrintWarning(fmt.Sprintf("⚠️  Warning: Could not remove old file %s: %v", originalPath, err))
	} else {
		terminal.PrintInfo(fmt.Sprintf("🗑️  Old file removed: %s", originalPath))
	}

	terminal.PrintSuccess(fmt.Sprintf("✅ Configuration saved to: %s", newFilePath))
	return nil
}

// saveInPlace overwrites the existing file
func (p *Processor) saveInPlace(originalPath string, encryptedData []byte) error {
	if err := os.WriteFile(originalPath, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write encrypted config file: %v", err)
	}

	terminal.PrintSuccess(fmt.Sprintf("Configuration updated: %s", originalPath))
	return nil
}
