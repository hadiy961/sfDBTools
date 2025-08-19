package backup_utils

import (
	"fmt"
	"sfDBTools/utils/database"
	"sfDBTools/utils/disk"
)

// ValidateBackupOptions validates the backup options before proceeding
func ValidateBackupOptions(options BackupOptions) error {
	if options.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	if options.Port <= 0 || options.Port > 65535 {
		return fmt.Errorf("invalid port: %d", options.Port)
	}
	if options.User == "" {
		return fmt.Errorf("user cannot be empty")
	}
	if options.DBName == "" {
		return fmt.Errorf("database name cannot be empty")
	}
	if options.OutputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}
	return nil
}

// validateAndPrepareBackup performs initial validation and preparation
func ValidateAndPrepareBackup(options BackupOptions) error {
	// Validate database connection
	config := database.Config{
		Host:     options.Host,
		Port:     options.Port,
		User:     options.User,
		Password: options.Password,
		DBName:   options.DBName,
	}

	if err := database.ValidateBeforeAction(config); err != nil {
		return err
	}

	// Check disk space if required (using default 1GB minimum)
	if options.VerifyDisk {
		if err := disk.CheckDiskSpace(options.OutputDir, 1024); err != nil { // 1GB default
			return err
		}
	}

	return nil
}
