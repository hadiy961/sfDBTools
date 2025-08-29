package configure

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// ConfigFileManager handles MariaDB configuration file operations
type ConfigFileManager struct {
	settings     *MariaDBSettings
	templatePath string
	targetPath   string
	backupPath   string
}

// NewConfigFileManager creates a new config file manager
func NewConfigFileManager(settings *MariaDBSettings, templatePath, targetPath string) *ConfigFileManager {
	backupPath := fmt.Sprintf("%s.backup.%d", targetPath, os.Getpid())

	return &ConfigFileManager{
		settings:     settings,
		templatePath: templatePath,
		targetPath:   targetPath,
		backupPath:   backupPath,
	}
}

// ProcessConfigFile handles the complete config file processing
func (c *ConfigFileManager) ProcessConfigFile() error {
	lg, _ := logger.Get()

	// Check if target config file exists and backup
	if err := c.backupExistingConfig(); err != nil {
		return fmt.Errorf("failed to backup existing config: %w", err)
	}

	// Read template file
	templateContent, err := c.readTemplateFile()
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Process template with user settings
	processedContent, err := c.processTemplate(templateContent)
	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	// Write processed config to target location
	if err := c.writeConfigFile(processedContent); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	lg.Info("MariaDB configuration file processed successfully",
		logger.String("template", c.templatePath),
		logger.String("target", c.targetPath))

	terminal.PrintSuccess(fmt.Sprintf("MariaDB configuration written to %s", c.targetPath))
	return nil
}

// backupExistingConfig creates a backup of existing config file
func (c *ConfigFileManager) backupExistingConfig() error {
	lg, _ := logger.Get()

	// Check if target file exists
	if _, err := os.Stat(c.targetPath); os.IsNotExist(err) {
		lg.Info("No existing config file to backup", logger.String("path", c.targetPath))
		return nil
	}

	// Create backup
	if err := copyFile(c.targetPath, c.backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	lg.Info("Existing config file backed up",
		logger.String("original", c.targetPath),
		logger.String("backup", c.backupPath))

	terminal.PrintInfo(fmt.Sprintf("Existing config backed up to %s", c.backupPath))
	return nil
}

// readTemplateFile reads the template configuration file
func (c *ConfigFileManager) readTemplateFile() (string, error) {
	lg, _ := logger.Get()

	content, err := ioutil.ReadFile(c.templatePath)
	if err != nil {
		lg.Error("Failed to read template file",
			logger.String("path", c.templatePath),
			logger.Error(err))
		return "", err
	}

	lg.Info("Template file read successfully", logger.String("path", c.templatePath))
	return string(content), nil
}

// processTemplate processes the template with user settings
func (c *ConfigFileManager) processTemplate(template string) (string, error) {
	lg, _ := logger.Get()

	processed := template

	// Replace configuration values based on user settings
	// Handle encryption settings FIRST before general path replacements
	if c.settings.EncryptionEnabled && c.settings.FileKeyManagementFile != "" {
		processed = strings.ReplaceAll(processed, "/var/lib/mysql/key_maria_nbc.txt", c.settings.FileKeyManagementFile)
	}

	// General path replacements - order matters!
	replacements := map[string]string{
		"SERVER-1":                        c.settings.ServerID,
		"/var/lib/mysqlbinlogs/mysql-bin": fmt.Sprintf("%s/mysql-bin", c.settings.BinlogDir),
		"/var/lib/mysql/mysql_error.log":  fmt.Sprintf("%s/mysql_error.log", c.settings.LogDir),
		"/var/lib/mysql/mysql_slow.log":   fmt.Sprintf("%s/mysql_slow.log", c.settings.LogDir),
		"/var/lib/mysql/mysql.sock":       fmt.Sprintf("%s/mysql.sock", c.settings.DataDir),
		"3306":                            fmt.Sprintf("%d", c.settings.Port),
	}

	// Apply specific replacements first
	for old, new := range replacements {
		processed = strings.ReplaceAll(processed, old, new)
	}

	// Do datadir replacement LAST and more specifically to avoid affecting other paths
	specificDatadirReplacements := []struct {
		old string
		new string
	}{
		{"datadir                                         = /var/lib/mysql", fmt.Sprintf("datadir                                         = %s", c.settings.DataDir)},
		{"innodb_data_home_dir                            = /var/lib/mysql", fmt.Sprintf("innodb_data_home_dir                            = %s", c.settings.DataDir)},
		{"innodb_log_group_home_dir                       = /var/lib/mysql", fmt.Sprintf("innodb_log_group_home_dir                       = %s", c.settings.DataDir)},
	}

	for _, replacement := range specificDatadirReplacements {
		processed = strings.ReplaceAll(processed, replacement.old, replacement.new)
	}

	lg.Info("Template processed with user settings",
		logger.String("server_id", c.settings.ServerID),
		logger.String("data_dir", c.settings.DataDir),
		logger.String("port", fmt.Sprintf("%d", c.settings.Port)))

	return processed, nil
}

// writeConfigFile writes the processed configuration to target file
func (c *ConfigFileManager) writeConfigFile(content string) error {
	lg, _ := logger.Get()

	// Ensure target directory exists
	targetDir := filepath.Dir(c.targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
	}

	// Write file
	if err := ioutil.WriteFile(c.targetPath, []byte(content), 0644); err != nil {
		lg.Error("Failed to write config file",
			logger.String("path", c.targetPath),
			logger.Error(err))
		return err
	}

	lg.Info("Config file written successfully", logger.String("path", c.targetPath))
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dst, sourceFile, 0644)
}
