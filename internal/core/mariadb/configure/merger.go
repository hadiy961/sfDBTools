package configure

import (
	"sfDBTools/internal/config/model"
	"sfDBTools/internal/logger"
)

// ConfigMerger handles merging existing configuration with app defaults
type ConfigMerger struct{}

// NewConfigMerger creates a new config merger
func NewConfigMerger() *ConfigMerger {
	return &ConfigMerger{}
}

// CreateDynamicDefaults creates MariaDB settings using existing config as defaults
func (cm *ConfigMerger) CreateDynamicDefaults(appConfig *model.Config) (*MariaDBSettings, error) {
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

	lg.Info("App config defaults created",
		logger.String("mariadb_key_from_config", appConfig.ConfigDir.MariaDBKey),
		logger.String("file_key_management_file", defaults.FileKeyManagementFile))

	// Override with existing configuration if found
	if existingConfig.Found {
		lg.Info("Using existing MariaDB configuration as defaults")
		defaults = cm.mergeWithExistingConfig(defaults, existingConfig)

		lg.Info("Dynamic defaults created from existing configuration",
			logger.String("data_dir", defaults.DataDir),
			logger.String("binlog_dir", defaults.BinlogDir),
			logger.String("log_dir", defaults.LogDir),
			logger.Int("port", defaults.Port),
			logger.String("server_id", defaults.ServerID),
			logger.String("file_key_management_file", defaults.FileKeyManagementFile))
	} else {
		lg.Info("No existing configuration found, using app config defaults",
			logger.String("file_key_management_file", defaults.FileKeyManagementFile))
	}

	return defaults, nil
}

// mergeWithExistingConfig merges existing configuration with defaults
func (cm *ConfigMerger) mergeWithExistingConfig(defaults *MariaDBSettings, existing *ParsedServerConfig) *MariaDBSettings {
	lg, _ := logger.Get()

	// Override directory settings if found in existing config
	if existing.DataDir != "" {
		defaults.DataDir = existing.DataDir
	}

	if existing.BinlogDir != "" {
		defaults.BinlogDir = existing.BinlogDir
	}

	if existing.LogDir != "" {
		defaults.LogDir = existing.LogDir
	}

	// Override port if found
	if existing.Port > 0 {
		defaults.Port = existing.Port
	}

	// Override server ID if found
	if existing.ServerID != "" {
		defaults.ServerID = existing.ServerID
	}

	// Set encryption status
	defaults.EncryptionEnabled = existing.EncryptionEnabled

	// Handle file key management file
	if existing.FileKeyManagementFile != "" {
		defaults.FileKeyManagementFile = existing.FileKeyManagementFile
		lg.Info("Overriding FileKeyManagementFile from existing config",
			logger.String("existing_file_key", existing.FileKeyManagementFile))
	} else {
		lg.Info("No FileKeyManagementFile found in existing config, keeping app config default",
			logger.String("app_config_default", defaults.FileKeyManagementFile))
	}

	return defaults
}

// CreateDynamicDefaults is a convenience function for backward compatibility
func CreateDynamicDefaults(appConfig *model.Config) (*MariaDBSettings, error) {
	merger := NewConfigMerger()
	return merger.CreateDynamicDefaults(appConfig)
}
