package backup_utils

import (
	"fmt"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/database"

	"github.com/spf13/cobra"
)

// ResolveBackupConfigWithoutDB resolves backup configuration without requiring a database name
func ResolveBackupConfigWithoutDB(cmd *cobra.Command) (*BackupConfig, error) {
	// Get default values from config
	_, _, _, defaultOutputDir,
		defaultCompress, defaultCompression, defaultCompressionLevel, defaultIncludeData,
		defaultEncrypt, defaultVerifyDisk, defaultRetentionDays, defaultCalculateChecksum, _ := config.GetBackupDefaults()

	backupConfig := &BackupConfig{}

	// Resolve database connection using the same logic as backup single
	host, port, user, password, source, err := ResolveDatabaseConnection(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve database connection: %w", err)
	}

	backupConfig.Host = host
	backupConfig.Port = port
	backupConfig.User = user
	backupConfig.Password = password

	// Display configuration source using the standard display function
	var details string
	configFile := common.GetStringFlagOrEnv(cmd, "config", "BACKUP_CONFIG", "")
	if configFile != "" {
		details = configFile
	} else {
		details = fmt.Sprintf("%s:%d (User: %s)", host, port, user)
	}
	DisplayConfigurationSource(source, details)

	// Resolve other backup options using common utilities
	backupConfig.OutputDir = common.GetStringFlagOrEnv(cmd, "output-dir", "OUTPUT_DIR", defaultOutputDir)
	backupConfig.Compress = common.GetBoolFlagOrEnv(cmd, "compress", "COMPRESS", defaultCompress)
	backupConfig.IncludeData = common.GetBoolFlagOrEnv(cmd, "data", "INCLUDE_DATA", defaultIncludeData)
	backupConfig.Encrypt = common.GetBoolFlagOrEnv(cmd, "encrypt", "ENCRYPT", defaultEncrypt)
	backupConfig.Compression = common.GetStringFlagOrEnv(cmd, "compression", "COMPRESSION", defaultCompression)
	backupConfig.CompressionLevel = common.GetStringFlagOrEnv(cmd, "compression-level", "COMPRESSION_LEVEL", defaultCompressionLevel)
	backupConfig.VerifyDisk = common.GetBoolFlagOrEnv(cmd, "verify-disk", "VERIFY_DISK", defaultVerifyDisk)
	backupConfig.RetentionDays = common.GetIntFlagOrEnv(cmd, "retention-days", "RETENTION_DAYS", defaultRetentionDays)
	backupConfig.CalculateChecksum = common.GetBoolFlagOrEnv(cmd, "calculate-checksum", "CALCULATE_CHECKSUM", defaultCalculateChecksum)

	if backupConfig.Compression == "" && backupConfig.Compress {
		backupConfig.Compression = "gzip"
	}

	return backupConfig, nil
}

// ExecuteSingleBackup executes backup for a single database with all standard operations
func ExecuteSingleBackup(backupConfig *BackupConfig, databaseName string, backupFunc func(BackupOptions) (*BackupResult, error)) (*BackupResult, error) {
	lg, _ := logger.Get()

	// Set database name for this backup
	backupConfig.DBName = databaseName
	options := backupConfig.ToBackupOptions()

	// Display parameters for this database
	DisplayBackupParameters(options)

	// Perform the backup
	result, err := backupFunc(options)
	if err != nil {
		lg.Error("Backup operation failed for database", logger.String("database", databaseName), logger.Error(err))
		return nil, fmt.Errorf("backup failed for %s: %w", databaseName, err)
	}

	if !result.Success {
		lg.Error("Backup completed with errors for database", logger.String("database", databaseName), logger.Error(result.Error))
		return nil, fmt.Errorf("backup completed with errors for %s: %w", databaseName, result.Error)
	}

	// Display comprehensive backup results (includes database metadata)
	DisplayBackupResults(result, options, databaseName)

	return result, nil
}

// CreateDatabaseConfig creates a database config from backup config
func CreateDatabaseConfig(backupConfig *BackupConfig) database.Config {
	return database.Config{
		Host:     backupConfig.Host,
		Port:     backupConfig.Port,
		User:     backupConfig.User,
		Password: backupConfig.Password,
		DBName:   "", // Connect without specific database for listing
	}
}

// ProcessDatabaseList reads and validates a database list from file
func ProcessDatabaseList(dbListFile string, dbConfig database.Config) (*DatabaseListResult, error) {
	lg, _ := logger.Get()

	// Read database list from file
	databases, err := common.ReadDatabaseList(dbListFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read database list: %w", err)
	}

	if len(databases) == 0 {
		return nil, fmt.Errorf("no databases found in the list file")
	}

	lg.Info("Database list loaded from file",
		logger.String("file", dbListFile),
		logger.Int("count", len(databases)),
		logger.Strings("databases", databases))

	// Validate databases from list
	result, err := ValidateDatabaseList(dbConfig, databases)
	if err != nil {
		return nil, fmt.Errorf("failed to validate database list: %w", err)
	}

	// Display validation results
	if err := DisplayDatabaseListValidation(result); err != nil {
		return nil, err
	}

	return result, nil
}
