package config

import (
	"fmt"
	"os"
	"strings"

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
	defaultCompress := false
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

	// Start with defaults
	host = defaultHost
	port = defaultPort
	user = defaultUser
	outputDir = defaultOutputDir
	compress = defaultCompress
	compression = defaultCompression
	compressionLevel = defaultCompressionLevel
	includeData = defaultIncludeData
	encrypt = defaultEncrypt
	verifyDisk = defaultVerifyDisk
	retentionDays = defaultRetentionDays
	calculateChecksum = defaultCalculateChecksum
	systemUser = defaultSystemUser

	// Try to load configuration and override defaults when available
	cfg, err := Get()
	if err != nil || cfg == nil {
		return
	}

	// Database defaults
	if cfg.Database.Host != "" {
		host = cfg.Database.Host
	}
	if cfg.Database.Port != 0 {
		port = cfg.Database.Port
	}
	if cfg.Database.User != "" {
		user = cfg.Database.User
	}

	// Output directory from backup storage base directory
	if cfg.Backup.Storage.BaseDirectory != "" {
		outputDir = cfg.Backup.Storage.BaseDirectory
	}

	// Compression settings
	if cfg.Backup.Compression.Algorithm != "" {
		compression = cfg.Backup.Compression.Algorithm
	}
	if cfg.Backup.Compression.Level != "" {
		compressionLevel = cfg.Backup.Compression.Level
	}
	// If config explicitly requires compression, use it; otherwise keep default
	compress = cfg.Backup.Compression.Required || compress

	// Determine includeData heuristically from mysqldump args (if --no-data present)
	if cfg.Mysqldump.Args != "" {
		argsLower := strings.ToLower(cfg.Mysqldump.Args)
		if strings.Contains(argsLower, "--no-data") {
			includeData = false
		} else {
			includeData = true
		}
	}

	// Security and verification
	encrypt = cfg.Backup.Security.EncryptionRequired || encrypt
	// consider either verify after write or disk space check as indicator to verify disk
	verifyDisk = cfg.Backup.Verification.VerifyAfterWrite || cfg.Backup.Verification.DiskSpaceCheck || verifyDisk
	calculateChecksum = cfg.Backup.Security.ChecksumVerification || cfg.Backup.Verification.CompareChecksums || calculateChecksum

	// Retention
	if cfg.Backup.Retention.Days != 0 {
		retentionDays = cfg.Backup.Retention.Days
	}

	// System user presence
	if len(cfg.SystemUsers.Users) > 0 {
		systemUser = true
	}

	return
}

// GetDatabaseCredentials returns database credentials, preferring encrypted config if available
func GetDatabaseCredentials() (host string, port int, user, password string, err error) {
	// Try encrypted config first, fallback to plain config if not available
	return GetDatabaseConfigWithEncryption()
}

// GetDatabaseConfigDirectory returns the directory path for database config files
func GetDatabaseConfigDirectory() (string, error) {
	cfg, err := Get()
	if err != nil {
		return "", fmt.Errorf("gagal membaca config: %w", err)
	}

	if cfg == nil {
		return "", fmt.Errorf("config tidak tersedia")
	}

	// Wajib menggunakan configured directory dari config file
	if cfg.ConfigDir.DatabaseConfig == "" {
		return "", fmt.Errorf("config_dir.database_config tidak diset di config.yaml")
	}

	return cfg.ConfigDir.DatabaseConfig, nil
}
