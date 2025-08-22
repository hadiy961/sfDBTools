package backup_restore_utils

import (
	"fmt"
	"os"
	"time"

	"sfDBTools/internal/config"
	"sfDBTools/internal/config/model"
	backup_single_mysqldump "sfDBTools/internal/core/backup/single/mysqldump"
	restore_single "sfDBTools/internal/core/restore/single"
	restoreUtils "sfDBTools/internal/core/restore/utils"
	"sfDBTools/internal/logger"
	backup_utils "sfDBTools/utils/backup"
	"sfDBTools/utils/database"

	_ "github.com/go-sql-driver/mysql"
)

// ExecuteBackupRestoreProduction executes the complete backup restore production flow
func ExecuteBackupRestoreProduction(options *BackupRestoreConfig) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	startTime := time.Now()
	lg.Info("Starting backup restore production execution",
		logger.String("account", options.Account),
		logger.String("target", options.Target),
		logger.String("production_db", options.ProductionDB),
		logger.String("target_db", options.TargetDB))

	// Create database config for connection
	dbConfig := database.Config{
		Host:     options.Host,
		Port:     options.Port,
		User:     options.User,
		Password: options.Password,
	}

	if options.DryRun {
		lg.Info("DRY RUN MODE: Showing what would be done")
		return executeDryRun(options, dbConfig)
	}

	// Step 1: Setup max_statement_time manager
	timeManager, err := database.SetupMaxStatementTimeManager(dbConfig, lg)
	if err != nil {
		lg.Warn("Failed to setup max_statement_time manager", logger.Error(err))
	}
	defer func() {
		if timeManager != nil {
			database.CleanupMaxStatementTimeManager(timeManager)
		}
	}()

	// Step 2: Verify production databases exist
	if err := verifyProductionDatabases(options, dbConfig); err != nil {
		return fmt.Errorf("production database verification failed: %w", err)
	}

	// Step 3: Create or verify target databases exist
	if err := createTargetDatabases(options, dbConfig); err != nil {
		return fmt.Errorf("target database creation failed: %w", err)
	}

	// Step 4: Check existing users and grant privileges
	if err := checkAndGrantUserPrivileges(options, dbConfig); err != nil {
		return fmt.Errorf("user privilege granting failed: %w", err)
	}

	// Step 5: Backup and restore production databases
	if err := backupAndRestoreDatabases(options, cfg, dbConfig); err != nil {
		return fmt.Errorf("backup and restore failed: %w", err)
	}

	duration := time.Since(startTime)
	lg.Info("Backup restore production completed successfully",
		logger.String("duration", duration.String()))

	return nil
}

// executeDryRun shows what would be done without actually executing
func executeDryRun(options *BackupRestoreConfig, dbConfig database.Config) error {
	lg, _ := logger.Get()

	lg.Info("DRY RUN: Would verify production databases exist",
		logger.String("production_db", options.ProductionDB),
		logger.String("dmart_db", options.ProductionDmartDB))

	lg.Info("DRY RUN: Would create/verify target databases",
		logger.String("target_db", options.TargetDB),
		logger.String("dmart_db", options.TargetDmartDB))

	lg.Info("DRY RUN: Would check existing users and grant privileges",
		logger.String("admin_user", fmt.Sprintf("sfnbc_%s_admin", options.Account)),
		logger.String("fin_user", fmt.Sprintf("sfnbc_%s_fin", options.Account)),
		logger.String("user", fmt.Sprintf("sfnbc_%s_user", options.Account)))

	lg.Info("DRY RUN: Would backup and restore databases",
		logger.String("from", options.ProductionDB),
		logger.String("to", options.TargetDB),
		logger.String("dmart_from", options.ProductionDmartDB),
		logger.String("dmart_to", options.TargetDmartDB))

	return nil
}

// verifyProductionDatabases checks if production databases exist
func verifyProductionDatabases(options *BackupRestoreConfig, dbConfig database.Config) error {
	lg, _ := logger.Get()

	// Check main production database
	dbConfig.DBName = options.ProductionDB
	if err := database.ValidateDatabase(dbConfig); err != nil {
		return fmt.Errorf("production database %s does not exist: %w", options.ProductionDB, err)
	}
	lg.Info("Production database verified", logger.String("database", options.ProductionDB))

	// Check dmart database
	dbConfig.DBName = options.ProductionDmartDB
	if err := database.ValidateDatabase(dbConfig); err != nil {
		return fmt.Errorf("production dmart database %s does not exist: %w", options.ProductionDmartDB, err)
	}
	lg.Info("Production dmart database verified", logger.String("database", options.ProductionDmartDB))

	return nil
}

// createTargetDatabases creates target databases if they don't exist
func createTargetDatabases(options *BackupRestoreConfig, dbConfig database.Config) error {
	lg, _ := logger.Get()

	// Create main target database
	if err := createDatabaseIfNotExists(options.TargetDB, dbConfig); err != nil {
		return fmt.Errorf("failed to create target database %s: %w", options.TargetDB, err)
	}
	lg.Info("Target database ready", logger.String("database", options.TargetDB))

	// Create dmart target database
	if err := createDatabaseIfNotExists(options.TargetDmartDB, dbConfig); err != nil {
		return fmt.Errorf("failed to create target dmart database %s: %w", options.TargetDmartDB, err)
	}
	lg.Info("Target dmart database ready", logger.String("database", options.TargetDmartDB))

	return nil
}

// createDatabaseIfNotExists creates a database if it doesn't exist
func createDatabaseIfNotExists(dbName string, dbConfig database.Config) error {
	lg, _ := logger.Get()

	// Test if database exists
	testConfig := dbConfig
	testConfig.DBName = dbName
	if err := database.ValidateDatabase(testConfig); err == nil {
		lg.Info("Database already exists", logger.String("database", dbName))
		return nil
	}

	// Database doesn't exist, create it
	lg.Info("Creating database", logger.String("database", dbName))

	// Connect without database selection
	db, err := database.GetWithoutDB(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database server: %w", err)
	}
	defer db.Close()

	// Create database
	createSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName)
	if _, err := db.Exec(createSQL); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	lg.Info("Database created successfully", logger.String("database", dbName))
	return nil
}

// checkAndGrantUserPrivileges checks existing users and grants privileges to target databases
func checkAndGrantUserPrivileges(options *BackupRestoreConfig, dbConfig database.Config) error {
	lg, _ := logger.Get()

	users := []string{
		fmt.Sprintf("sfnbc_%s_admin", options.Account),
		fmt.Sprintf("sfnbc_%s_fin", options.Account),
		fmt.Sprintf("sfnbc_%s_user", options.Account),
	}

	for _, user := range users {
		if err := checkUserAndGrantPrivileges(user, options, dbConfig); err != nil {
			lg.Warn("Failed to grant privileges to user", logger.String("user", user), logger.Error(err))
			// Continue to next user instead of failing completely
		}
	}

	return nil
}

// checkUserAndGrantPrivileges checks if user exists and grants privileges to target databases
func checkUserAndGrantPrivileges(username string, options *BackupRestoreConfig, dbConfig database.Config) error {
	lg, _ := logger.Get()

	db, err := database.GetWithoutDB(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database server: %w", err)
	}
	defer db.Close()

	// Check if user exists
	var exists int
	checkSQL := "SELECT COUNT(*) FROM mysql.user WHERE User = ?"
	if err := db.QueryRow(checkSQL, username).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check if user exists: %w", err)
	}

	if exists == 0 {
		lg.Info("User does not exist, skipping", logger.String("user", username))
		return nil
	}

	lg.Info("User exists, granting privileges to target databases", logger.String("user", username))

	// Grant privileges to target databases
	databases := []string{options.TargetDB, options.TargetDmartDB}
	for _, dbName := range databases {
		grantSQL := fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'%%'", dbName, username)
		if _, err := db.Exec(grantSQL); err != nil {
			lg.Warn("Failed to grant privileges",
				logger.String("user", username),
				logger.String("database", dbName),
				logger.Error(err))
			// Continue to next database instead of failing
		} else {
			lg.Info("Privileges granted",
				logger.String("user", username),
				logger.String("database", dbName))
		}
	}

	// Flush privileges
	if _, err := db.Exec("FLUSH PRIVILEGES"); err != nil {
		lg.Warn("Failed to flush privileges", logger.Error(err))
	}

	return nil
}

// backupAndRestoreDatabases performs the backup and restore operations
func backupAndRestoreDatabases(options *BackupRestoreConfig, cfg *model.Config, dbConfig database.Config) error {
	// Backup and restore main database
	if err := backupAndRestoreDatabase(options.ProductionDB, options.TargetDB, options, cfg, dbConfig); err != nil {
		return fmt.Errorf("failed to backup/restore main database: %w", err)
	}

	// Backup and restore dmart database
	if err := backupAndRestoreDatabase(options.ProductionDmartDB, options.TargetDmartDB, options, cfg, dbConfig); err != nil {
		return fmt.Errorf("failed to backup/restore dmart database: %w", err)
	}

	return nil
}

// backupAndRestoreDatabase performs backup and restore for a single database
func backupAndRestoreDatabase(sourceDB, targetDB string, options *BackupRestoreConfig, cfg *model.Config, dbConfig database.Config) error {
	lg, _ := logger.Get()

	lg.Info("Starting backup and restore operation",
		logger.String("source", sourceDB),
		logger.String("target", targetDB))

	// Step 1: Backup source database
	backupOptions := backup_utils.BackupOptions{
		Host:              dbConfig.Host,
		Port:              dbConfig.Port,
		User:              dbConfig.User,
		Password:          dbConfig.Password,
		DBName:            sourceDB,
		OutputDir:         cfg.Backup.OutputDir,
		Compress:          true,
		Compression:       "gzip",
		CompressionLevel:  "6",
		IncludeData:       true,
		Encrypt:           false, // Set to false for intermediate backup
		VerifyDisk:        false,
		RetentionDays:     1, // Clean up after 1 day for temp backups
		CalculateChecksum: false,
		IncludeSystem:     false,
		SystemUsers:       false,
	}

	lg.Info("Backing up source database", logger.String("database", sourceDB))
	result, err := backup_single_mysqldump.BackupSingle(backupOptions)
	if err != nil {
		return fmt.Errorf("backup failed for %s: %w", sourceDB, err)
	}

	if !result.Success {
		return fmt.Errorf("backup failed for %s: %v", sourceDB, result.Error)
	}

	lg.Info("Backup completed successfully",
		logger.String("source", sourceDB),
		logger.String("file", result.OutputFile),
		logger.String("size", fmt.Sprintf("%.2f MB", float64(result.OutputSize)/(1024*1024))))

	// Step 2: Restore to target database
	restoreOptions := restoreUtils.RestoreOptions{
		Host:     dbConfig.Host,
		Port:     dbConfig.Port,
		User:     dbConfig.User,
		Password: dbConfig.Password,
		DBName:   targetDB,
		File:     result.OutputFile,
	}

	lg.Info("Restoring to target database", logger.String("database", targetDB))
	if err := restore_single.RestoreSingle(restoreOptions); err != nil {
		return fmt.Errorf("restore failed for %s: %w", targetDB, err)
	}

	lg.Info("Restore completed successfully",
		logger.String("target", targetDB),
		logger.String("from_file", result.OutputFile))

	// Step 3: Clean up temporary backup file
	if err := os.Remove(result.OutputFile); err != nil {
		lg.Warn("Failed to clean up temporary backup file",
			logger.String("file", result.OutputFile),
			logger.Error(err))
	} else {
		lg.Info("Temporary backup file cleaned up", logger.String("file", result.OutputFile))
	}

	// Clean up metadata file if exists
	if result.BackupMetaFile != "" {
		if err := os.Remove(result.BackupMetaFile); err != nil {
			lg.Warn("Failed to clean up temporary metadata file",
				logger.String("file", result.BackupMetaFile),
				logger.Error(err))
		}
	}

	return nil
}
