package configure

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sfDBTools/internal/config/model"
	"sfDBTools/internal/logger"
)

// ServerConfigParser handles parsing MariaDB server configuration files
type ServerConfigParser struct {
	configPaths []string
}

// NewServerConfigParser creates a new server config parser
func NewServerConfigParser() *ServerConfigParser {
	return &ServerConfigParser{
		configPaths: []string{
			"/etc/my.cnf.d/server.cnf",
			"/etc/mysql/my.cnf",
			"/etc/my.cnf",
			"/usr/local/mysql/my.cnf",
		},
	}
}

// ParsedServerConfig represents parsed MariaDB configuration
type ParsedServerConfig struct {
	DataDir               string
	BinlogDir             string
	LogDir                string
	Port                  int
	ServerID              string
	EncryptionEnabled     bool
	FileKeyManagementFile string
	Found                 bool
}

// ParseExistingConfig attempts to parse existing MariaDB configuration
func (p *ServerConfigParser) ParseExistingConfig() (*ParsedServerConfig, error) {
	lg, _ := logger.Get()

	config := &ParsedServerConfig{
		Port:              3306,
		ServerID:          "1",
		EncryptionEnabled: false,
		Found:             false,
	}

	// Try each config path
	for _, configPath := range p.configPaths {
		if _, err := os.Stat(configPath); err == nil {
			lg.Info("Found MariaDB config file", logger.String("path", configPath))

			if err := p.parseConfigFile(configPath, config); err != nil {
				lg.Warn("Failed to parse config file",
					logger.String("path", configPath),
					logger.Error(err))
				continue
			}

			config.Found = true
			lg.Info("Successfully parsed MariaDB configuration",
				logger.String("config_file", configPath),
				logger.String("data_dir", config.DataDir),
				logger.String("binlog_dir", config.BinlogDir),
				logger.Int("port", config.Port))
			break
		}
	}

	if !config.Found {
		lg.Info("No existing MariaDB configuration found, will use defaults")
	}

	return config, nil
}

// parseConfigFile parses a single configuration file
func (p *ServerConfigParser) parseConfigFile(configPath string, config *ParsedServerConfig) error {
	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle sections
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			continue
		}

		// Only parse mysqld and server sections
		if currentSection != "mysqld" && currentSection != "server" {
			continue
		}

		// Parse key-value pairs
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Remove quotes if present
			if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
				value = strings.Trim(value, `"`)
			}

			p.parseConfigValue(key, value, config)
		}
	}

	return scanner.Err()
}

// parseConfigValue parses individual configuration values
func (p *ServerConfigParser) parseConfigValue(key, value string, config *ParsedServerConfig) {
	switch key {
	case "datadir":
		config.DataDir = value
	case "log_bin":
		// Extract directory from log_bin path
		if value != "" {
			config.BinlogDir = filepath.Dir(value)
		}
	case "log_error", "general_log_file", "slow_query_log_file":
		// Extract directory from log file paths
		if value != "" && config.LogDir == "" {
			config.LogDir = filepath.Dir(value)
		}
	case "port":
		if port, err := strconv.Atoi(value); err == nil {
			config.Port = port
		}
	case "server_id", "server-id":
		config.ServerID = value
	case "innodb-encrypt-tables", "innodb_encrypt_tables":
		config.EncryptionEnabled = strings.ToLower(value) == "on" ||
			strings.ToLower(value) == "1" ||
			strings.ToLower(value) == "true"
	case "file_key_management_filename", "file-key-management-filename":
		config.FileKeyManagementFile = value
	}
}

// CreateDynamicDefaults creates MariaDB settings using existing config as defaults
func CreateDynamicDefaults(appConfig *model.Config) (*MariaDBSettings, error) {
	lg, _ := logger.Get()

	// First try to parse existing config
	parser := NewServerConfigParser()
	existingConfig, err := parser.ParseExistingConfig()
	if err != nil {
		lg.Warn("Failed to parse existing config, using app defaults", logger.Error(err))
		return DefaultMariaDBSettings(appConfig), nil
	}

	// Start with app config defaults
	defaults := DefaultMariaDBSettings(appConfig)

	// Override with existing configuration if found
	if existingConfig.Found {
		lg.Info("Using existing MariaDB configuration as defaults")

		if existingConfig.DataDir != "" {
			defaults.DataDir = existingConfig.DataDir
		}

		if existingConfig.BinlogDir != "" {
			defaults.BinlogDir = existingConfig.BinlogDir
		}

		if existingConfig.LogDir != "" {
			defaults.LogDir = existingConfig.LogDir
		}

		if existingConfig.Port > 0 {
			defaults.Port = existingConfig.Port
		}

		if existingConfig.ServerID != "" {
			defaults.ServerID = existingConfig.ServerID
		}

		defaults.EncryptionEnabled = existingConfig.EncryptionEnabled

		if existingConfig.FileKeyManagementFile != "" {
			defaults.FileKeyManagementFile = existingConfig.FileKeyManagementFile
		}

		lg.Info("Dynamic defaults created from existing configuration",
			logger.String("data_dir", defaults.DataDir),
			logger.String("binlog_dir", defaults.BinlogDir),
			logger.String("log_dir", defaults.LogDir),
			logger.Int("port", defaults.Port),
			logger.String("server_id", defaults.ServerID))
	} else {
		lg.Info("No existing configuration found, using app config defaults")
	}

	return defaults, nil
}
