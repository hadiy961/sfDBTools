package configure

import "sfDBTools/internal/config/model"

// ConfigureConfig holds configuration for MariaDB configuration
type ConfigureConfig struct {
	AutoConfirm   bool
	SkipUserSetup bool
	SkipDBSetup   bool
}

// MariaDBSettings represents the MariaDB configuration settings from user input
type MariaDBSettings struct {
	ServerID              string
	FileKeyManagementFile string
	BinlogDir             string
	DataDir               string
	LogDir                string
	Port                  int
	EncryptionEnabled     bool
}

// DefaultMariaDBSettings creates default settings from config
func DefaultMariaDBSettings(config *model.Config) *MariaDBSettings {
	mariadbConfig := config.MariaDB.Installation

	// Ensure data directory is not /var/lib/mysql
	dataDir := mariadbConfig.DataDir
	if dataDir == "/var/lib/mysql" {
		dataDir = "/data/mysql" // Use alternative default
	}

	// Ensure binlog directory matches the data directory path structure
	binlogDir := mariadbConfig.BinlogDir
	if dataDir != "/var/lib/mysql" && binlogDir == "/var/lib/mysqlbinlogs" {
		binlogDir = "/data/mysqlbinlogs" // Use consistent path
	}

	// Ensure log directory matches the data directory path structure
	logDir := mariadbConfig.LogDir
	if dataDir != "/var/lib/mysql" && logDir == "/var/lib/mysql" {
		logDir = dataDir // Use data directory for logs
	}

	return &MariaDBSettings{
		ServerID:              "SERVER-1",
		FileKeyManagementFile: config.ConfigDir.MariaDBKey,
		BinlogDir:             binlogDir,
		DataDir:               dataDir,
		LogDir:                logDir,
		Port:                  mariadbConfig.Port,
		EncryptionEnabled:     true,
	}
}

// DefaultConfigureConfig returns default configuration
func DefaultConfigureConfig() *ConfigureConfig {
	return &ConfigureConfig{
		AutoConfirm:   false,
		SkipUserSetup: false,
		SkipDBSetup:   false,
	}
}
