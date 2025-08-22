package config

import (
	"fmt"
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

		if err = validate.All(&c); err != nil {
			err = fmt.Errorf("validasi config gagal: %w", err)
			return
		}

		cfg = &c
	})

	return cfg, err
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
		}
	}
	return config
}

// ValidateConfigFile checks if config file exists and is readable
func ValidateConfigFile() error {
	possiblePaths := []string{
		"./config/config.yaml",
		"./config/config.yml",
		"config.yaml",
		"config.yml",
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
	if err != nil {
		// Return hardcoded default if config fails to load
		return "./config"
	}

	// Use configured directory or fallback to default
	if cfg.ConfigDir.DatabaseConfig != "" {
		return cfg.ConfigDir.DatabaseConfig
	}

	return "./config"
}
