package config

import (
	"fmt"
	"os"

	"sfDBTools/internal/config/model"
	"sfDBTools/internal/config/validate"
)

var (
	cfg *model.Config
)

func LoadConfig() (*model.Config, error) {
	v, loadErr := loadViper()
	if loadErr != nil {
		return nil, loadErr
	}

	var c model.Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("gagal parsing config: %w", err)
	}

	if err := validate.All(&c); err != nil {
		return nil, fmt.Errorf("validasi config gagal: %w", err)
	}

	cfg = &c
	return cfg, nil
}

// Get returns the loaded configuration
func Get() (*model.Config, error) {
	if cfg == nil {
		return LoadConfig()
	}
	return cfg, nil
}

// ValidateConfigFile checks if config file exists and is readable
func ValidateConfigFile() error {
	requiredPath := "/etc/sfDBTools/config/config.yaml"

	if _, err := os.Stat(requiredPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file konfigurasi tidak ditemukan di %s. Jalankan 'sfdbtools config generate' untuk membuat konfigurasi default", requiredPath)
		}
		return fmt.Errorf("tidak dapat mengakses file konfigurasi di %s: %w", requiredPath, err)
	}

	return nil
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
		// Return default path consistent with main config location
		return "/etc/sfDBTools/config/db_config"
	}

	// Use configured directory or fallback to default
	if cfg.ConfigDir.DatabaseConfig != "" {
		return cfg.ConfigDir.DatabaseConfig
	}

	// Return default path consistent with main config location
	return "/etc/sfDBTools/config/db_config"
}
