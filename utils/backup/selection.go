package backup_utils

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
	"sfDBTools/utils/database/info"

	"github.com/spf13/cobra"
)

// DatabaseSelectionWorkflow handles reusable database selection workflow for any backup type
type DatabaseSelectionWorkflow struct {
	UseInteractiveSelection bool
	DatabaseList            []string
	DBListFile              string
}

// SelectionBackupFunction defines the function signature for selection-based backup operations
type SelectionBackupFunction func(cmd *cobra.Command, backupConfig *BackupConfig, databases []string, dbConfig database.Config) error

// ExecuteSelectionBackupWorkflow executes selection-based backup workflow that can be reused by any backup type
func ExecuteSelectionBackupWorkflow(
	cmd *cobra.Command,
	workflow DatabaseSelectionWorkflow,
	backupFunc SelectionBackupFunction,
	backupType string,
) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	lg.Info("Starting selection backup workflow", logger.String("backup_type", backupType))

	// 1. Resolve backup configuration
	backupConfig, err := ResolveBackupConfigWithoutDB(cmd)
	if err != nil {
		return fmt.Errorf("failed to resolve backup configuration: %w", err)
	}

	// 2. Create database config and test connection
	dbConfig := CreateDatabaseConfig(backupConfig)
	if err := TestDatabaseConnection(dbConfig); err != nil {
		return err
	}

	// 3. Determine database selection method and get databases
	var selectedDatabases []string

	if workflow.UseInteractiveSelection {
		// Interactive multi-database selection
		selectedDatabases, err = info.SelectMultipleDatabasesInteractive(dbConfig)
		if err != nil {
			return fmt.Errorf("failed to select databases interactively: %w", err)
		}
	} else if workflow.DBListFile != "" {
		// Database list from file
		dbListResult, err := ProcessDatabaseList(workflow.DBListFile, dbConfig)
		if err != nil {
			return fmt.Errorf("failed to process database list: %w", err)
		}
		selectedDatabases = dbListResult.ValidDatabases
	} else if len(workflow.DatabaseList) > 0 {
		// Pre-provided database list
		selectedDatabases = workflow.DatabaseList
	} else {
		return fmt.Errorf("no database selection method provided")
	}

	if len(selectedDatabases) == 0 {
		return fmt.Errorf("no databases selected for %s backup", backupType)
	}

	lg.Info("Databases selected for backup",
		logger.String("backup_type", backupType),
		logger.Int("count", len(selectedDatabases)),
		logger.Strings("databases", selectedDatabases))

	// 4. Execute backup for selected databases using provided function
	if err := backupFunc(cmd, backupConfig, selectedDatabases, dbConfig); err != nil {
		return fmt.Errorf("%s backup failed: %w", backupType, err)
	}

	lg.Info("Selection backup workflow completed successfully", logger.String("backup_type", backupType))
	return nil
}

// ExecuteMultipleSelectionBackup executes backup for multiple databases using standard database backup function
func ExecuteMultipleSelectionBackup(
	backupConfig *BackupConfig,
	databases []string,
	backupFunc func(BackupOptions) (*BackupResult, error),
	operationType string,
) (*MultiBackupResult, error) {
	return ExecuteMultipleDatabaseBackup(backupConfig, databases, backupFunc, operationType)
}
