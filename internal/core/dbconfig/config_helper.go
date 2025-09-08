package dbconfig

import (
	"fmt"
	"path/filepath"
	"strings"

	"sfDBTools/internal/config"
	"sfDBTools/utils/common"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"
)

// ConfigHelper provides common configuration file operations
type ConfigHelper struct {
	*BaseProcessor
	fileManager *dbconfig.FileManager
}

// NewConfigHelper creates a new config helper
func NewConfigHelper() (*ConfigHelper, error) {
	base, err := NewBaseProcessor()
	if err != nil {
		return nil, err
	}

	return &ConfigHelper{
		BaseProcessor: base,
		fileManager:   dbconfig.NewFileManager(),
	}, nil
}

// LoadDecryptedConfig loads and decrypts a configuration file
func (ch *ConfigHelper) LoadDecryptedConfig(filePath, encryptionPassword string) (*config.EncryptedDatabaseConfig, error) {
	// Load and decrypt configuration
	spinner := terminal.NewProgressSpinner("Decrypting configuration...")
	spinner.Start()

	dbConfig, err := common.LoadEncryptedConfigFromFile(filePath, encryptionPassword)
	spinner.Stop()

	if err != nil {
		return nil, common.HandleDecryptionError(err, filePath)
	}

	return dbConfig, nil
}

// GetConfigNameFromPath extracts config name from file path
func (ch *ConfigHelper) GetConfigNameFromPath(filePath string) string {
	filename := filepath.Base(filePath)
	return strings.TrimSuffix(filename, ".cnf.enc")
}

// DisplayConfigDetails shows configuration details in a formatted way
func (ch *ConfigHelper) DisplayConfigDetails(configName string, filePath string) error {
	return dbconfig.DisplayConfigDetails(configName, filePath)
}

// SelectConfigFile prompts user to select a configuration file
func (ch *ConfigHelper) SelectConfigFile(operation dbconfig.OperationType) (string, error) {
	files, err := ch.fileManager.ListConfigFiles()
	if err != nil {
		return "", fmt.Errorf("error listing config files: %v", err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no configuration files found")
	}

	configName, err := dbconfig.PromptForConfigName(files, operation)
	if err != nil {
		return "", err
	}

	// Find the file by name
	for _, file := range files {
		if file.Name == configName {
			return file.Path, nil
		}
	}

	return "", fmt.Errorf("configuration file not found: %s", configName)
}

// ValidateConfigExists checks if a config file exists and is valid
func (ch *ConfigHelper) ValidateConfigExists(filePath string) error {
	if err := common.ValidateConfigFile(filePath); err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}
	return nil
}

// BackupConfigFile creates a backup of a configuration file
func (ch *ConfigHelper) BackupConfigFile(filePath string) (string, error) {
	backupPath, err := ch.fileManager.BackupConfigFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup: %v", err)
	}

	terminal.PrintInfo(fmt.Sprintf("Backup created: %s", filepath.Base(backupPath)))
	return backupPath, nil
}

// DeleteConfigFile deletes a configuration file safely
func (ch *ConfigHelper) DeleteConfigFile(filePath string) error {
	return ch.fileManager.DeleteConfigFile(filePath)
}

// GetFileManager returns the file manager instance
func (ch *ConfigHelper) GetFileManager() *dbconfig.FileManager {
	return ch.fileManager
}
