package backup_utils

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

// ExecuteGrantBackup executes the backup for user grants
func ExecuteGrantBackup() {

}

// ExecuteMultipleDatabaseBackup executes backup for multiple databases
func ExecuteMultipleDatabaseBackup(
	backupConfig *BackupConfig,
	databases []string,
	backupFunc func(BackupOptions) (*BackupResult, error),
	operationType string,
) (*MultiBackupResult, error) {
	lg, _ := logger.Get()
	terminal.ClearAndShowHeader("Backup Tools - Multi-Database Backup")
	lg.Debug("Starting multi-database backup",
		logger.String("operation", operationType),
		logger.Int("total_databases", len(databases)),
		logger.Strings("databases", databases))

	result := &MultiBackupResult{
		TotalProcessed:  len(databases),
		SuccessCount:    0,
		FailedDatabases: []string{},
	}

	for i, dbName := range databases {
		terminal.PrintSubHeader(fmt.Sprintf("Processing Database (%d/%d): %s", i+1, len(databases), dbName))
		lg.Debug("Processing database",
			logger.String("database", dbName),
			logger.Int("current", i+1),
			logger.Int("total", len(databases)))

		_, err := ExecuteSingleBackup(backupConfig, dbName, backupFunc)
		if err != nil {
			lg.Error("Database backup failed",
				logger.String("database", dbName),
				logger.Error(err))
			result.FailedDatabases = append(result.FailedDatabases, dbName)
			continue
		}

		result.SuccessCount++
	}

	// Final summary
	terminal.PrintSubHeader("Backup Summary")
	lg.Info("Multi-database backup completed",
		logger.String("operation", operationType),
		logger.Int("total_processed", result.TotalProcessed),
		logger.Int("successful", result.SuccessCount),
		logger.Int("failed", len(result.FailedDatabases)),
		logger.Strings("failed_databases", result.FailedDatabases))

	if len(result.FailedDatabases) > 0 {
		return result, fmt.Errorf("some databases failed to backup")
	}

	return result, nil
}

// ExecuteListBackupWorkflow executes the complete list backup workflow
func ExecuteListBackupWorkflow(
	cmd *cobra.Command,
	backupFunc func(BackupOptions) (*BackupResult, error),
) error {
	lg, _ := logger.Get()
	lg.Info("Starting list backup process")

	// 1. Resolve backup configuration
	backupConfig, err := ResolveBackupConfigWithoutDB(cmd)
	if err != nil {
		return fmt.Errorf("failed to resolve backup configuration: %w", err)
	}

	fmt.Println()

	// 2. Resolve database list file
	dbListFile, err := ResolveDBListFile(cmd)
	if err != nil {
		return fmt.Errorf("failed to resolve db_list file: %w", err)
	}

	// 3. Create database config and test connection
	dbConfig := CreateDatabaseConfig(backupConfig)
	if err := TestDatabaseConnection(dbConfig); err != nil {
		return err
	}

	// 4. Process and validate database list
	dbListResult, err := ProcessDatabaseList(dbListFile, dbConfig)
	if err != nil {
		return err
	}

	// 5. Execute backup for valid databases
	multiResult, err := ExecuteMultipleDatabaseBackup(
		backupConfig,
		dbListResult.ValidDatabases,
		backupFunc,
		"List",
	)

	if multiResult != nil {
		multiResult.InvalidDatabases = dbListResult.InvalidDatabases
	}

	return err
}
