package install

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"sfDBTools/internal/logger"
)

// ConfigFixer handles MariaDB configuration fixes
type ConfigFixer struct{}

// NewConfigFixer creates a new config fixer
func NewConfigFixer() *ConfigFixer {
	return &ConfigFixer{}
}

// FixPostInstallationIssues fixes common configuration issues after MariaDB installation
func (c *ConfigFixer) FixPostInstallationIssues(dataDir string) error {
	lg, _ := logger.Get()

	// Check if data directory exists and is properly initialized
	// Fix 1: Handle missing encryption key file
	if err := c.fixEncryptionKeyIssue(dataDir); err != nil {
		lg.Warn("Failed to fix encryption key issue", logger.Error(err))
	}

	// Fix 2: Ensure proper ownership of mysql directory
	if err := c.fixDataDirectoryOwnership(dataDir); err != nil {
		lg.Warn("Failed to fix data directory ownership", logger.Error(err))
	}

	// Fix 3: Check and fix configuration file issues
	if err := c.fixConfigurationIssues(); err != nil {
		lg.Warn("Failed to fix configuration issues", logger.Error(err))
	}

	lg.Info("Post-installation fixes completed")
	return nil
}

// fixEncryptionKeyIssue handles missing encryption key file
func (c *ConfigFixer) fixEncryptionKeyIssue(dataDir string) error {
	lg, _ := logger.Get()

	keyFile := dataDir + "/key_maria_nbc.txt"

	// Check if key file exists
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		lg.Info("Creating missing encryption key file", logger.String("key_file", keyFile))

		// Create a simple encryption key file
		keyContent := "1;F1E2D3C4B5A697887766554433221100"

		if err := os.WriteFile(keyFile, []byte(keyContent), 0600); err != nil {
			return fmt.Errorf("failed to create encryption key file: %w", err)
		}

		// Set proper ownership
		cmd := exec.Command("chown", "mysql:mysql", keyFile)
		if err := cmd.Run(); err != nil {
			lg.Warn("Failed to set ownership for key file", logger.Error(err))
		}

		lg.Info("Encryption key file created successfully", logger.String("key_file", keyFile))
	}

	return nil
}

// fixDataDirectoryOwnership ensures proper ownership of data directory
func (c *ConfigFixer) fixDataDirectoryOwnership(dataDir string) error {
	lg, _ := logger.Get()

	// Ensure mysql user exists
	cmd := exec.Command("id", "mysql")
	if err := cmd.Run(); err != nil {
		lg.Warn("MySQL user not found, creating...", logger.Error(err))

		// Create mysql user if it doesn't exist
		createUserCmd := exec.Command("useradd", "-r", "-s", "/bin/false", "mysql")
		if err := createUserCmd.Run(); err != nil {
			return fmt.Errorf("failed to create mysql user: %w", err)
		}
	}

	// Set proper ownership of data directory
	cmd = exec.Command("chown", "-R", "mysql:mysql", dataDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set data directory ownership: %w", err)
	}

	lg.Info("Data directory ownership fixed", logger.String("data_dir", dataDir))
	return nil
}

// fixConfigurationIssues handles MariaDB configuration problems
func (c *ConfigFixer) fixConfigurationIssues() error {
	lg, _ := logger.Get()

	configFiles := []string{
		"/etc/my.cnf",
		"/etc/mysql/mariadb.conf.d/50-server.cnf",
		"/etc/mysql/my.cnf",
	}

	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); err == nil {
			if err := c.fixConfigFile(configFile); err != nil {
				lg.Warn("Failed to fix config file",
					logger.String("file", configFile),
					logger.Error(err))
			}
		}
	}

	lg.Info("Configuration issues check completed")
	return nil
}

// fixConfigFile fixes specific configuration file issues
func (c *ConfigFixer) fixConfigFile(configFile string) error {
	lg, _ := logger.Get()

	// Read current config
	content, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	configContent := string(content)
	modified := false

	// Fix 1: Disable file_key_management if it's causing issues
	if strings.Contains(configContent, "file_key_management") &&
		!strings.Contains(configContent, "#file_key_management") {
		lg.Info("Disabling file_key_management plugin in config",
			logger.String("file", configFile))

		configContent = strings.ReplaceAll(configContent,
			"file_key_management", "#file_key_management")
		modified = true
	}

	// Fix 2: Add skip-networking temporarily if needed
	if !strings.Contains(configContent, "skip-networking") {
		lg.Info("Adding skip-networking to config for initial setup",
			logger.String("file", configFile))

		// Add skip-networking under [mysqld] section
		if strings.Contains(configContent, "[mysqld]") {
			configContent = strings.Replace(configContent,
				"[mysqld]", "[mysqld]\nskip-networking", 1)
			modified = true
		}
	}

	// Write back if modified
	if modified {
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
		lg.Info("Configuration file updated", logger.String("file", configFile))
	}

	return nil
}
