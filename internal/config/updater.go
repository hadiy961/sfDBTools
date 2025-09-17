package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// UpdateConfig safely updates the config.yaml file with new values
type ConfigUpdater struct {
	configFilePath string
	backupDir      string
}

// NewConfigUpdater creates a new config updater instance
func NewConfigUpdater() (*ConfigUpdater, error) {
	// Find the config file path using the same logic as loader.go
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to determine executable path: %w", err)
	}
	appDir := filepath.Dir(exePath)
	appConfigDir := filepath.Join(appDir, "config")
	systemConfigDir := "/etc/sfDBTools/config"

	cwd, _ := os.Getwd()
	cwdConfigDir := filepath.Join(cwd, "config")

	var configFilePath string
	if fileExists(filepath.Join(cwdConfigDir, "config.yaml")) {
		configFilePath = filepath.Join(cwdConfigDir, "config.yaml")
	} else if fileExists(filepath.Join(appConfigDir, "config.yaml")) {
		configFilePath = filepath.Join(appConfigDir, "config.yaml")
	} else if fileExists(filepath.Join(systemConfigDir, "config.yaml")) {
		configFilePath = filepath.Join(systemConfigDir, "config.yaml")
	} else {
		return nil, fmt.Errorf("config file not found in any of the expected locations")
	}

	return &ConfigUpdater{
		configFilePath: configFilePath,
		backupDir:      filepath.Join(filepath.Dir(configFilePath), "..", "backup"),
	}, nil
}

// backupConfigFile creates a backup of the current config file
func (cu *ConfigUpdater) backupConfigFile() (string, error) {
	// Ensure backup directory exists
	if err := os.MkdirAll(cu.backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupFileName := fmt.Sprintf("config-backup-%s.yaml", timestamp)
	backupPath := filepath.Join(cu.backupDir, backupFileName)

	// Read original file
	originalData, err := os.ReadFile(cu.configFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read original config file: %w", err)
	}

	// Write backup file
	if err := os.WriteFile(backupPath, originalData, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	return backupPath, nil
}

// UpdateMariaDBConfig updates the MariaDB section of the config file
func (cu *ConfigUpdater) UpdateMariaDBConfig(updates map[string]interface{}) error {
	// Create backup first
	backupPath, err := cu.backupConfigFile()
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Read current config file as raw YAML
	configData, err := os.ReadFile(cu.configFilePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML into generic map to preserve structure and comments
	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(configData, &yamlData); err != nil {
		return fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Update MariaDB section
	if yamlData["mariadb"] == nil {
		yamlData["mariadb"] = make(map[string]interface{})
	}

	mariadbSection, ok := yamlData["mariadb"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("mariadb section is not a valid map")
	}

	// Apply updates
	for key, value := range updates {
		if value != nil && value != "" && value != 0 { // Only update non-empty values
			mariadbSection[key] = value
		}
	}

	// Marshal back to YAML
	updatedData, err := yaml.Marshal(&yamlData)
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %w", err)
	}

	// Write updated config file
	if err := os.WriteFile(cu.configFilePath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write updated config file: %w", err)
	}

	_ = backupPath // Use backup path to avoid unused variable warning
	return nil
}

// GetConfigFilePath returns the path to the config file being used
func (cu *ConfigUpdater) GetConfigFilePath() string {
	return cu.configFilePath
}
