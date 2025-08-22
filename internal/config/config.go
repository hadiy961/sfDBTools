package config

import (
	"fmt"
	"os"
	"sync"

	"sfDBTools/internal/config/model"
	"sfDBTools/internal/config/validate"
)

var (
	cfg  *model.Config
	once sync.Once
)

func LoadConfig() (*model.Config, error) {
	var err error
	once.Do(func() {
		v, loadErr := loadViper()
		if loadErr != nil {
			err = loadErr
			return
		}

		var c model.Config
		if err = v.Unmarshal(&c); err != nil {
			err = fmt.Errorf("gagal parsing config: %w", err)
			return
		}

		// Merge with defaults to ensure all fields are set
		mergeWithDefaults(&c)

		if err = validate.All(&c); err != nil {
			err = fmt.Errorf("validasi config gagal: %w", err)
			return
		}

		cfg = &c
	})

	return cfg, err
}

// mergeWithDefaults ensures all required fields have default values
func mergeWithDefaults(c *model.Config) {
	// Get default config
	defaults := getDefaultConfig()
	
	// Merge ConfigDir if empty
	if c.ConfigDir.DatabaseConfig == "" {
		c.ConfigDir.DatabaseConfig = defaults.ConfigDir.DatabaseConfig
	}
	
	// Merge General config if empty
	if c.General.ClientCode == "" {
		c.General.ClientCode = defaults.General.ClientCode
	}
	if c.General.AppName == "" {
		c.General.AppName = defaults.General.AppName
	}
	if c.General.Version == "" {
		c.General.Version = defaults.General.Version
	}
	if c.General.Author == "" {
		c.General.Author = defaults.General.Author
	}
	
	// Merge Log config if empty
	if c.Log.Level == "" {
		c.Log.Level = defaults.Log.Level
	}
	if c.Log.Format == "" {
		c.Log.Format = defaults.Log.Format
	}
	if c.Log.Timezone == "" {
		c.Log.Timezone = defaults.Log.Timezone
	}
	if c.Log.File.Dir == "" {
		c.Log.File.Dir = defaults.Log.File.Dir
	}
	if c.Log.File.RetentionDays == 0 {
		c.Log.File.RetentionDays = defaults.Log.File.RetentionDays
	}
	
	// Merge MariaDB config if empty
	if c.MariaDB.DefaultVersion == "" {
		c.MariaDB.DefaultVersion = defaults.MariaDB.DefaultVersion
	}
	if c.MariaDB.Installation.Port == 0 {
		c.MariaDB.Installation.Port = defaults.MariaDB.Installation.Port
	}
	if c.MariaDB.Installation.BaseDir == "" {
		c.MariaDB.Installation.BaseDir = defaults.MariaDB.Installation.BaseDir
	}
	if c.MariaDB.Installation.DataDir == "" {
		c.MariaDB.Installation.DataDir = defaults.MariaDB.Installation.DataDir
	}
	if c.MariaDB.Installation.LogDir == "" {
		c.MariaDB.Installation.LogDir = defaults.MariaDB.Installation.LogDir
	}
	if c.MariaDB.Installation.BinlogDir == "" {
		c.MariaDB.Installation.BinlogDir = defaults.MariaDB.Installation.BinlogDir
	}
	if c.MariaDB.Installation.KeyFile == "" {
		c.MariaDB.Installation.KeyFile = defaults.MariaDB.Installation.KeyFile
	}
	
	// Merge backup config if empty
	if c.Backup.OutputDir == "" {
		c.Backup.OutputDir = defaults.Backup.OutputDir
	}
	if c.Backup.Compression == "" {
		c.Backup.Compression = defaults.Backup.Compression
	}
	if c.Backup.CompressionLevel == "" {
		c.Backup.CompressionLevel = defaults.Backup.CompressionLevel
	}
	if c.Backup.RetentionDays == 0 {
		c.Backup.RetentionDays = defaults.Backup.RetentionDays
	}
}

// getDefaultConfig returns a config with all default values
func getDefaultConfig() *model.Config {
	return &model.Config{
		General: model.GeneralConfig{
			ClientCode: "DEFAULT",
			AppName:    "sfDBTools",
			Version:    "1.0.0",
			Author:     "Hadiyatna Muflihun",
		},
		Log: model.LogConfig{
			Level:    "info",
			Format:   "text",
			Timezone: "UTC",
			Output: model.LogOutput{
				Console: true,
				File:    true,
				Syslog:  false,
			},
			File: model.LogFileSetting{
				Dir:           "./logs",
				RotateDaily:   true,
				RetentionDays: 7,
			},
		},
		Database: model.DatabaseConfig{
			Host: "localhost",
			Port: 3306,
			User: "root",
		},
		Backup: model.BackupConfig{
			OutputDir:         "./backup",
			Compress:          true,
			Compression:       "pgzip",
			CompressionLevel:  "fast",
			IncludeData:       true,
			Encrypt:           false,
			VerifyDisk:        true,
			RetentionDays:     30,
			CalculateChecksum: true,
			SystemUser:        false,
		},
		ConfigDir: model.ConfigDirConfig{
			DatabaseConfig: "./config/db_config",
		},
		SystemUsers: model.SystemUsers{
			Users: []string{"root"},
		},
		MariaDB: model.MariaDBConfig{
			DefaultVersion: "10.6.23",
			Installation: model.MariaDBInstallConfig{
				BaseDir:             "/var/lib/mysql",
				DataDir:             "/var/lib/mysql",
				LogDir:              "/var/lib/mysql",
				BinlogDir:           "/var/lib/mysqlbinlogs",
				Port:                3306,
				KeyFile:             "./config/key_maria_nbc.txt",
				SeparateDirectories: true,
			},
		},
	}
}

// Get returns the loaded configuration
func Get() (*model.Config, error) {
	if cfg == nil {
		return LoadConfig()
	}
	return cfg, nil
}

// GetOrDefault returns the loaded configuration or default values if config loading fails
func GetOrDefault() *model.Config {
	config, err := Get()
	if err != nil {
		// Return default configuration if loading fails
		return &model.Config{
			General: model.GeneralConfig{
				ClientCode: "DEFAULT",
				AppName:    "sfDBTools",
				Version:    "1.0.0",
				Author:     "Hadiyatna Muflihun",
			},
			Log: model.LogConfig{
				Level:    "info",
				Format:   "text",
				Timezone: "UTC",
				Output: model.LogOutput{
					Console: true,
					File:    true,
					Syslog:  false,
				},
				File: model.LogFileSetting{
					Dir:           "./logs",
					RotateDaily:   true,
					RetentionDays: 7,
				},
			},
			Database: model.DatabaseConfig{
				Host: "localhost",
				Port: 3306,
				User: "root",
			},
			Backup: model.BackupConfig{
				OutputDir:         "./backup",
				Compress:          true,
				Compression:       "pgzip",
				CompressionLevel:  "fast",
				IncludeData:       true,
				Encrypt:           false,
				VerifyDisk:        true,
				RetentionDays:     30,
				CalculateChecksum: true,
				SystemUser:        false,
			},
			ConfigDir: model.ConfigDirConfig{
				DatabaseConfig: "./config/db_config",
			},
			SystemUsers: model.SystemUsers{
				Users: []string{"root"},
			},
			MariaDB: model.MariaDBConfig{
				DefaultVersion: "10.6.23",
				Installation: model.MariaDBInstallConfig{
					BaseDir:             "/var/lib/mysql",
					DataDir:             "/var/lib/mysql",
					LogDir:              "/var/lib/mysql",
					BinlogDir:           "/var/lib/mysqlbinlogs",
					Port:                3306,
					KeyFile:             "./config/key_maria_nbc.txt",
					SeparateDirectories: true,
				},
			},
		}
	}
	return config
}

// ValidateConfigFile checks if config file exists and is readable
func ValidateConfigFile() error {
	possiblePaths := []string{
		"./config/config.yaml",
		"./config/config.yml",
		"./config.yaml",
		"./config.yml",
		"/etc/sfdbtools/config.yaml",
		"/etc/sfdbtools/config.yml",
	}

	// Add user config path if HOME is set
	if homeDir := os.Getenv("HOME"); homeDir != "" {
		possiblePaths = append(possiblePaths,
			homeDir+"/.config/sfdbtools/config.yaml",
			homeDir+"/.config/sfdbtools/config.yml",
		)
	}

	for _, path := range possiblePaths {
		if fileExists(path) {
			return nil // Found valid config file
		}
	}

	return fmt.Errorf("file konfigurasi tidak ditemukan. Jalankan 'sfdbtools config generate' untuk membuat konfigurasi default")
}

// GetBackupDefaults returns default values for backup command flags
func GetBackupDefaults() (host string, port int, user string, outputDir string,
	compress bool, compression string, compressionLevel string, includeData bool,
	encrypt bool, verifyDisk bool, retentionDays int, calculateChecksum bool, systemUser bool) {

	// Hardcoded defaults - safer approach to prevent any segfault
	defaultHost := "localhost"
	defaultPort := 3306
	defaultUser := "root"
	defaultOutputDir := "./backup"
	defaultCompress := true
	defaultCompression := "pgzip"
	defaultCompressionLevel := "fast"
	defaultIncludeData := true
	defaultEncrypt := false
	defaultVerifyDisk := true
	defaultRetentionDays := 30
	defaultCalculateChecksum := true
	defaultSystemUser := false

	// Try to get config safely
	defer func() {
		if r := recover(); r != nil {
			// If any panic occurs, just use defaults
		}
	}()

	cfg := GetOrDefault()
	if cfg != nil && cfg.Database.Host != "" {
		return cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Backup.OutputDir,
			cfg.Backup.Compress, cfg.Backup.Compression, cfg.Backup.CompressionLevel, cfg.Backup.IncludeData,
			cfg.Backup.Encrypt, cfg.Backup.VerifyDisk, cfg.Backup.RetentionDays, cfg.Backup.CalculateChecksum, cfg.Backup.SystemUser
	}

	// Return hardcoded defaults if config is not available
	return defaultHost, defaultPort, defaultUser, defaultOutputDir,
		defaultCompress, defaultCompression, defaultCompressionLevel, defaultIncludeData,
		defaultEncrypt, defaultVerifyDisk, defaultRetentionDays, defaultCalculateChecksum, defaultSystemUser
}

// GetDatabaseCredentials returns database credentials, preferring encrypted config if available
func GetDatabaseCredentials() (host string, port int, user, password string, err error) {
	// Try encrypted config first, fallback to plain config if not available
	return GetDatabaseConfigWithEncryption()
}

// GetDatabaseConfigDirectory returns the directory path for database config files
func GetDatabaseConfigDirectory() string {
	cfg, err := Get()
	if err != nil || cfg == nil {
		// Return hardcoded default if config fails to load
		return "./config/db_config"
	}

	// Use configured directory or fallback to default
	if cfg.ConfigDir.DatabaseConfig != "" {
		return cfg.ConfigDir.DatabaseConfig
	}

	// Try some common default paths based on where the config might be located
	possiblePaths := []string{
		"./config/db_config",
		"/etc/sfdbtools/db_config",
	}

	// If we're running as a system service, prefer /etc path
	homeDir := os.Getenv("HOME")
	if homeDir != "" && homeDir != "/" {
		// Running as user
		userConfigPath := homeDir + "/.config/sfdbtools/db_config"
		possiblePaths = append([]string{userConfigPath}, possiblePaths...)
	}

	// Return first path that exists, or default if none exist
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "./config/db_config"
}
