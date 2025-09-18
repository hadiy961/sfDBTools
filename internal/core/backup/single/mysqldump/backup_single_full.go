package backup_single_mysqldump

import (
	"fmt"
	"os"
	"time"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	backup_utils "sfDBTools/utils/backup"
	"sfDBTools/utils/database"
	"sfDBTools/utils/database/info"
	"sfDBTools/utils/fs"
)

// BackupSingle performs a backup of a single database
func BackupSingle(options backup_utils.BackupOptions) (*backup_utils.BackupResult, error) {
	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	_, err = config.Get()
	if err != nil {
		lg.Error("Failed to load configuration", logger.Error(err))
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// defaultConfig := database.Config{
	// 	Host:     cfg.Database.Host,
	// 	Port:     cfg.Database.Port,
	// 	User:     cfg.Database.User,
	// 	Password: cfg.Database.Password,
	// 	DBName:   cfg.Database.DBName,
	// }

	// // Validate basic connection
	// if err := database.ValidateConnection(defaultConfig); err != nil {
	// 	lg.Error("Connection validation failed", logger.Error(err))
	// }

	// // Validate database exists
	// if err := database.ValidateDatabase(defaultConfig); err != nil {
	// 	lg.Error("Database validation failed", logger.Error(err))
	// }

	manager := fs.NewManager()
	if !manager.Dir().Exists(options.OutputDir) {
		if err := manager.Dir().Create(options.OutputDir); err != nil {
			lg.Error("Failed to create output directory", logger.Error(err))
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
	}
	if err := manager.Dir().IsWritable(options.OutputDir); err != nil {
		lg.Error("Output directory validation failed", logger.Error(err))
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
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
	lg.Info("Starting single database backup using mysqldump",
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

	// Collect replication information before backup
	replicationInfo, err := backup_utils.GetReplicationInfoForBackup(config)
	if err != nil {
		lg.Warn("Failed to collect replication information before backup", logger.Error(err))
	} else if replicationInfo != nil {
		lg.Info("Replication information collected successfully before backup")
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

	// backup_utils.LogBackupCompletion(options, result, lg)
	return result, nil
}
