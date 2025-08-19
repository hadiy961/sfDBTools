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

// GetBackupDefaults returns default values for backup command flags
func GetBackupDefaults() (host string, port int, user string, outputDir string,
	compress bool, compression string, compressionLevel string, includeData bool,
	encrypt bool, verifyDisk bool, retentionDays int, calculateChecksum bool, systemUser bool) {

	// Load config with fallback values
	cfg, err := Get()
	if err != nil {
		// Return hardcoded defaults if config fails
		return "localhost", 3306, "root", "./backup",
			true, "pgzip", "fast", true,
			false, true, 30, true, false
	}

	return cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Backup.OutputDir,
		cfg.Backup.Compress, cfg.Backup.Compression, cfg.Backup.CompressionLevel, cfg.Backup.IncludeData,
		cfg.Backup.Encrypt, cfg.Backup.VerifyDisk, cfg.Backup.RetentionDays, cfg.Backup.CalculateChecksum, cfg.Backup.SystemUser
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
