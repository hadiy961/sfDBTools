package generate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"sfDBTools/internal/config"
	"sfDBTools/utils/crypto"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"
)

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
			terminal.PrintWarning("Configuration not saved.")
			return nil
		}

		// Create backup of existing file
		if _, err := p.configHelper.BackupConfigFile(filePath); err != nil {
			terminal.PrintWarning(fmt.Sprintf("Could not create backup: %v", err))
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

	terminal.PrintSuccess(fmt.Sprintf("Configuration '%s' saved successfully!", configName))
	terminal.PrintInfo(fmt.Sprintf("Saved to: %s", filePath))

	return nil
}
