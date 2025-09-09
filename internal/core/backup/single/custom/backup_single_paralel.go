package backup_single_custom

import (
	"fmt"
	"os"
	"time"

	"sfDBTools/internal/logger"
	backup_utils "sfDBTools/utils/backup"
	"sfDBTools/utils/database"
	"sfDBTools/utils/database/info"
	"sfDBTools/utils/file"
)

// BackupSingle performs a backup of a single database
func BackupCustom(options backup_utils.BackupOptions) (*backup_utils.BackupResult, error) {
	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	errDir := file.ValidateDir(options.OutputDir)
	if errDir != nil {
		lg.Error(errDir.Error())
		fmt.Printf("Error: %v\n", errDir)
		os.Exit(1)
	}

	// Clean up old backups based on retention policy
	if removed, err := backup_utils.CleanupOldBackups(options.OutputDir, options.RetentionDays); err != nil {
		lg.Warn("Failed to cleanup old backups", logger.Error(err))
	} else if len(removed) > 0 {
		lg.Info("Old backup directories removed", logger.Strings("dirs", removed), logger.Int("count", len(removed)))
	} else {
		lg.Debug("No old backup directories to remove", logger.String("outputDir", options.OutputDir))
	}

	startTime := time.Now()
	lg.Info("Starting single database backup",
		logger.String("database", options.DBName),
		logger.String("host", options.Host),
		logger.Int("port", options.Port))

	result := backup_utils.InitializeBackupResult(options)

	if err := backup_utils.ValidateAndPrepareBackup(options); err != nil {
		result.Error = err
		return result, err
	}

	config := database.Config{
		Host: options.Host, Port: options.Port, User: options.User,
		Password: options.Password, DBName: options.DBName,
	}

	if timeManager, _ := database.SetupMaxStatementTimeManager(config, lg); timeManager != nil {
		defer database.CleanupMaxStatementTimeManager(timeManager)
	}

	dbInfo := info.CollectDatabaseInfo(config, lg)

	outputFile, metaFile, err := backup_utils.SetupBackupPaths(options)
	if err != nil {
		result.Error = err
		return result, err
	}
	result.OutputFile, result.BackupMetaFile = outputFile, metaFile

	if err := performBackup(options, outputFile); err != nil {
		result.Error = err
		return result, err
	}

	if err := backup_utils.FinalizeBackupResult(result, outputFile, startTime, options); err != nil {
		lg.Warn("Failed to finalize backup result", logger.Error(err))
	}

	if err := backup_utils.CreateMetadataFile(options, result, config, dbInfo); err != nil {
		lg.Warn("Failed to create metadata file", logger.Error(err))
	}

	return result, nil
}
