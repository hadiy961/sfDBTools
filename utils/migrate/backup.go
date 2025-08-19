package migrate_utils

import (
	"fmt"

	backup_single_mysqldump "sfDBTools/internal/core/backup/single/mysqldump"
	"sfDBTools/internal/logger"
	backup_utils "sfDBTools/utils/backup"
)

// BackupDatabaseForMigration creates a backup of a database (source or target) for migration
func BackupDatabaseForMigration(config *MigrationConfig, isSource bool, lg *logger.Logger) (string, error) {
	// Get database connection info
	dbInfo := getDBInfo(config, isSource)
	dbType := getDBType(isSource)

	lg.Info("Backing up database for migration",
		logger.String("database", dbInfo.DBName),
		logger.String("type", dbType))

	// Create backup options with defaults
	backupOptions := createBackupOptions(dbInfo, config.MigrateData)

	// Perform backup
	result, err := backup_single_mysqldump.BackupSingle(backupOptions)
	if err != nil {
		return "", fmt.Errorf("failed to backup %s database: %w", dbType, err)
	}

	return result.OutputFile, nil
}

// getDBInfo extracts database connection info based on source/target
func getDBInfo(config *MigrationConfig, isSource bool) backup_utils.BackupOptions {
	if isSource {
		return backup_utils.BackupOptions{
			Host:     config.SourceHost,
			Port:     config.SourcePort,
			User:     config.SourceUser,
			Password: config.SourcePassword,
			DBName:   config.SourceDBName,
		}
	}
	return backup_utils.BackupOptions{
		Host:     config.TargetHost,
		Port:     config.TargetPort,
		User:     config.TargetUser,
		Password: config.TargetPassword,
		DBName:   config.TargetDBName,
	}
}

// getDBType returns string representation of database type
func getDBType(isSource bool) string {
	if isSource {
		return "source"
	}
	return "target"
}

// createBackupOptions creates backup options with default settings
func createBackupOptions(dbInfo backup_utils.BackupOptions, includeData bool) backup_utils.BackupOptions {
	return backup_utils.BackupOptions{
		Host:              dbInfo.Host,
		Port:              dbInfo.Port,
		User:              dbInfo.User,
		Password:          dbInfo.Password,
		DBName:            dbInfo.DBName,
		OutputDir:         "./backup",
		Compress:          true,
		Compression:       "gzip",
		CompressionLevel:  "default",
		IncludeData:       includeData,
		Encrypt:           true,
		VerifyDisk:        true,
		RetentionDays:     30,
		CalculateChecksum: true,
	}
}
