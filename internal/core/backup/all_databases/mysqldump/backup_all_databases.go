package backup_all_databases_mysqldump

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	backup_utils "sfDBTools/utils/backup"
	"sfDBTools/utils/database"
	"sfDBTools/utils/disk"
)

// BackupAllDatabases performs a backup of all databases into a single file
func BackupAllDatabases(options backup_utils.AllDatabasesBackupOptions) (*backup_utils.AllDatabasesBackupResult, error) {
	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	_, err = config.Get()
	if err != nil {
		lg.Error("Failed to load configuration", logger.Error(err))
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate output directory
	errDir := disk.ValidateOutputDir(options.OutputDir)
	if errDir != nil {
		lg.Error(errDir.Error())
		return nil, errDir
	}

	// Clean up old backups based on retention policy
	if removed, err := backup_utils.CleanupOldBackups(options.OutputDir, options.RetentionDays); err != nil {
		lg.Warn("Failed to cleanup old backups", logger.Error(err))
	} else if len(removed) > 0 {
		lg.Info("Old backup directories removed", logger.Strings("dirs", removed), logger.Int("count", len(removed)))
	} else {
		lg.Info("No old backup directories to remove", logger.String("outputDir", options.OutputDir))
	}

	startTime := time.Now()
	lg.Info("Starting all databases backup using mysqldump",
		logger.String("host", options.Host),
		logger.Int("port", options.Port),
		logger.Bool("exclude_system", options.ExcludeSystemDatabases))

	// Initialize result
	result := &backup_utils.AllDatabasesBackupResult{
		BackupResult: backup_utils.BackupResult{
			Success:         false,
			CompressionUsed: options.Compression,
			Encrypted:       options.Encrypt,
			IncludedData:    options.IncludeData,
		},
		ProcessedDatabases: []string{},
		SkippedDatabases:   []string{},
	}

	// Validate options
	if err := backup_utils.ValidateAllDatabasesBackupOptions(options); err != nil {
		result.BackupResult.Error = err
		return result, err
	}

	// Setup database connection
	dbConfig := database.Config{
		Host:     options.Host,
		Port:     options.Port,
		User:     options.User,
		Password: options.Password,
	}

	// Setup max statement time manager
	if timeManager, _ := database.SetupMaxStatementTimeManager(dbConfig, lg); timeManager != nil {
		defer database.CleanupMaxStatementTimeManager(timeManager)
	}

	// Collect replication information before backup
	replicationInfo, err := backup_utils.GetReplicationInfoForBackup(dbConfig)
	if err != nil {
		lg.Warn("Failed to collect replication information before all databases backup", logger.Error(err))
	} else if replicationInfo != nil {
		lg.Info("Replication information collected successfully before all databases backup")
	}

	// Get all databases
	databases, err := backup_utils.GetAllDatabasesList(dbConfig, options.ExcludeSystemDatabases)
	if err != nil {
		result.BackupResult.Error = err
		return result, err
	}

	result.TotalDatabases = len(databases)
	lg.Info("Found databases to backup", logger.Int("count", len(databases)), logger.Strings("databases", databases))

	// Generate output paths
	outputFile, metaFile := backup_utils.GenerateAllDatabasesOutputPaths(options)
	result.BackupResult.OutputFile = outputFile
	result.BackupResult.BackupMetaFile = metaFile

	// Perform the backup
	processedDatabases, skippedDatabases, err := performAllDatabasesBackup(options, outputFile, databases)
	if err != nil {
		result.BackupResult.Error = err
		return result, err
	}

	result.ProcessedDatabases = processedDatabases
	result.SkippedDatabases = skippedDatabases

	// Finalize backup result
	if err := backup_utils.FinalizeBackupResult(&result.BackupResult, outputFile, startTime, options.BackupOptions); err != nil {
		lg.Warn("Failed to finalize backup result", logger.Error(err))
	}

	// Create metadata file
	metadata := backup_utils.CreateAllDatabasesMetadata(options, result, dbConfig)
	if err := backup_utils.CreateMetadataFile(options.BackupOptions, &result.BackupResult, dbConfig, nil); err != nil {
		lg.Warn("Failed to create metadata file", logger.Error(err))
	}

	// Save custom metadata for all databases
	if err := saveAllDatabasesMetadata(metaFile, metadata); err != nil {
		lg.Warn("Failed to save all databases metadata", logger.Error(err))
	}

	result.BackupResult.Success = true
	lg.Info("All databases backup completed successfully",
		logger.Int("total_databases", result.TotalDatabases),
		logger.Int("processed", len(result.ProcessedDatabases)),
		logger.Int("skipped", len(result.SkippedDatabases)),
		logger.String("output_file", result.BackupResult.OutputFile),
		logger.String("duration", result.BackupResult.Duration.String()))

	return result, nil
}

// performAllDatabasesBackup performs the actual backup operation for all databases
func performAllDatabasesBackup(options backup_utils.AllDatabasesBackupOptions, outputFile string, databases []string) ([]string, []string, error) {
	lg, _ := logger.Get()

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Execute mysqldump for all databases
	processedDatabases, skippedDatabases, err := executeAllDatabasesMysqldump(options, outputFile, databases)
	if err != nil {
		lg.Error("mysqldump execution failed", logger.Error(err))
		return processedDatabases, skippedDatabases, fmt.Errorf("mysqldump failed: %w", err)
	}

	lg.Info("All databases mysqldump completed successfully",
		logger.Int("processed", len(processedDatabases)),
		logger.Int("skipped", len(skippedDatabases)))

	return processedDatabases, skippedDatabases, nil
}
