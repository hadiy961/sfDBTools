package command_migrate

import (
	"fmt"
	"os"
	"time"

	"sfDBTools/internal/core/restore/single"
	restoreUtils "sfDBTools/internal/core/restore/utils"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/database"
	"sfDBTools/utils/database/info"
	migrate_utils "sfDBTools/utils/migrate"

	"github.com/spf13/cobra"
)

var SelectionMigrateCmd = &cobra.Command{
	Use:   "selection",
	Short: "Migrate multiple databases using interactive selection between MySQL servers",
	Long: `This command allows you to select multiple databases interactively for migration from a source MySQL server to a target MySQL server.
You can choose from 1 to N databases from the available list or specify a database list file.

The migration process for each database includes:
1. Backup the target database (if exists)
2. Backup the source database (structure + data + users + grants)
3. Drop the target database
4. Create new target database
5. Restore the source database to target
6. Verify data integrity (optional)

The migration uses existing backup and restore functionality for reliability.`,
	Example: `sfDBTools migrate selection --source-config ./config/source.cnf.enc --target-config ./config/target.cnf.enc --db_list ./db_list.txt
sfDBTools migrate selection --source-host localhost --source-user root --target-host remote.server.com --target-user admin
sfDBTools migrate selection  # Fully interactive - will prompt for everything`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the value of the db_list flag
		dbListPath, err := cmd.Flags().GetString("db_list")
		if err != nil {
			lg, _ := logger.Get()
			lg.Error("Failed to get db_list flag", logger.Error(err))
			os.Exit(1)
		}

		if dbListPath == "" {
			if err := executeSelectionMigration(cmd); err != nil {
				lg, _ := logger.Get()
				lg.Error("Selection migration failed", logger.Error(err))
				os.Exit(1)
			}
		} else {
			if err := executeListMigration(cmd); err != nil {
				lg, _ := logger.Get()
				lg.Error("List migration failed", logger.Error(err))
				os.Exit(1)
			}
		}
	},
}

// executeSelectionMigration handles the main selection migration execution logic
func executeSelectionMigration(cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	lg.Info("Starting selection migration process")

	// 1. Resolve source and target configurations without specific database
	sourceConfig, targetConfig, err := resolveMigrationConfigurations(cmd)
	if err != nil {
		return fmt.Errorf("failed to resolve migration configurations: %w", err)
	}

	// 2. Get available databases from source and let user select multiple
	sourceDBConfig := database.Config{
		Host:     sourceConfig.SourceHost,
		Port:     sourceConfig.SourcePort,
		User:     sourceConfig.SourceUser,
		Password: sourceConfig.SourcePassword,
	}
	selectedDatabases, err := info.SelectMultipleDatabasesInteractive(sourceDBConfig)
	if err != nil {
		return fmt.Errorf("failed to select databases: %w", err)
	}

	if len(selectedDatabases) == 0 {
		return fmt.Errorf("no databases selected for migration")
	}

	lg.Info("Databases selected for migration",
		logger.Int("count", len(selectedDatabases)),
		logger.Strings("databases", selectedDatabases))

	// 3. Prompt for confirmation before proceeding with bulk migration
	if err := migrate_utils.PromptBulkMigrationConfirmation(sourceConfig, targetConfig, selectedDatabases); err != nil {
		lg.Info("Selection migration cancelled", logger.String("reason", err.Error()))
		return err
	}

	// 4. Execute migration for selected databases
	return executeBulkMigration(sourceConfig, targetConfig, selectedDatabases, lg)
}

// executeListMigration handles the main list migration execution logic
func executeListMigration(cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	lg.Info("Starting list migration process")

	// 1. Get the db_list file path
	dbListPath, _ := cmd.Flags().GetString("db_list")

	// 2. Read database list from file
	selectedDatabases, err := common.ReadDatabaseList(dbListPath)
	if err != nil {
		return fmt.Errorf("failed to read database list from file: %w", err)
	}

	if len(selectedDatabases) == 0 {
		return fmt.Errorf("no databases found in the list file")
	}

	lg.Info("Databases loaded from file for migration",
		logger.String("file", dbListPath),
		logger.Int("count", len(selectedDatabases)),
		logger.Strings("databases", selectedDatabases))

	// 3. Resolve source and target configurations
	sourceConfig, targetConfig, err := resolveMigrationConfigurations(cmd)
	if err != nil {
		return fmt.Errorf("failed to resolve migration configurations: %w", err)
	}

	// 4. Prompt for confirmation before proceeding with bulk migration
	if err := migrate_utils.PromptBulkMigrationConfirmation(sourceConfig, targetConfig, selectedDatabases); err != nil {
		lg.Info("List migration cancelled", logger.String("reason", err.Error()))
		return err
	}

	// 5. Execute migration for databases from list
	return executeBulkMigration(sourceConfig, targetConfig, selectedDatabases, lg)
}

// resolveMigrationConfigurations resolves source and target configurations without specific database
func resolveMigrationConfigurations(cmd *cobra.Command) (*migrate_utils.MigrationConfig, *migrate_utils.MigrationConfig, error) {
	// Resolve source database connection
	sourceHost, sourcePort, sourceUser, sourcePassword, _, err := migrate_utils.ResolveSourceDatabaseConnection(cmd)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve source database connection: %w", err)
	}

	// Resolve target database connection
	targetHost, targetPort, targetUser, targetPassword, _, err := migrate_utils.ResolveTargetDatabaseConnection(cmd)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve target database connection: %w", err)
	}

	// Create source configuration template
	sourceConfig := &migrate_utils.MigrationConfig{
		SourceHost:       sourceHost,
		SourcePort:       sourcePort,
		SourceUser:       sourceUser,
		SourcePassword:   sourcePassword,
		MigrateUsers:     true,
		MigrateData:      true,
		MigrateStructure: true,
		VerifyData:       true,
		BackupTarget:     true,
		DropTarget:       true,
		CreateTarget:     true,
	}

	// Create target configuration template
	targetConfig := &migrate_utils.MigrationConfig{
		TargetHost:       targetHost,
		TargetPort:       targetPort,
		TargetUser:       targetUser,
		TargetPassword:   targetPassword,
		MigrateUsers:     true,
		MigrateData:      true,
		MigrateStructure: true,
		VerifyData:       true,
		BackupTarget:     true,
		DropTarget:       true,
		CreateTarget:     true,
	}

	// Display configuration sources
	lg, _ := logger.Get()
	lg.Info("Source configuration loaded",
		logger.String("host", sourceHost),
		logger.Int("port", sourcePort),
		logger.String("user", sourceUser))
	lg.Info("Target configuration loaded",
		logger.String("host", targetHost),
		logger.Int("port", targetPort),
		logger.String("user", targetUser))

	return sourceConfig, targetConfig, nil
}

// executeBulkMigration executes migration for multiple databases
func executeBulkMigration(sourceConfig, targetConfig *migrate_utils.MigrationConfig, databases []string, lg *logger.Logger) error {
	startTime := time.Now()
	successCount := 0
	errorCount := 0
	var errors []string

	lg.Info("Starting bulk database migration",
		logger.Int("database_count", len(databases)),
		logger.String("source_host", sourceConfig.SourceHost),
		logger.Int("source_port", sourceConfig.SourcePort),
		logger.String("source_user", sourceConfig.SourceUser),
		logger.String("target_host", targetConfig.TargetHost),
		logger.Int("target_port", targetConfig.TargetPort),
		logger.String("target_user", targetConfig.TargetUser))

	for i, dbName := range databases {
		lg.Info("Starting database migration",
			logger.Int("current", i+1),
			logger.Int("total", len(databases)),
			logger.String("database", dbName))

		// Create specific migration config for this database
		migrationConfig := &migrate_utils.MigrationConfig{
			SourceHost:       sourceConfig.SourceHost,
			SourcePort:       sourceConfig.SourcePort,
			SourceUser:       sourceConfig.SourceUser,
			SourcePassword:   sourceConfig.SourcePassword,
			SourceDBName:     dbName,
			TargetHost:       targetConfig.TargetHost,
			TargetPort:       targetConfig.TargetPort,
			TargetUser:       targetConfig.TargetUser,
			TargetPassword:   targetConfig.TargetPassword,
			TargetDBName:     dbName, // Same name on target
			MigrateUsers:     sourceConfig.MigrateUsers,
			MigrateData:      sourceConfig.MigrateData,
			MigrateStructure: sourceConfig.MigrateStructure,
			VerifyData:       sourceConfig.VerifyData,
			BackupTarget:     sourceConfig.BackupTarget,
			DropTarget:       sourceConfig.DropTarget,
			CreateTarget:     sourceConfig.CreateTarget,
		}

		// Execute migration for this database
		err := executeSingleDatabaseMigration(migrationConfig, lg)
		if err != nil {
			errorCount++
			errMsg := fmt.Sprintf("Database %s: %v", dbName, err)
			errors = append(errors, errMsg)
			lg.Error("Database migration failed",
				logger.String("database", dbName),
				logger.Error(err))
		} else {
			successCount++
			lg.Info("Database migration completed successfully",
				logger.String("database", dbName))
		}
	}

	// Display final summary
	duration := time.Since(startTime)
	migrate_utils.DisplayMigrationSummary(databases, successCount, errorCount, duration.String(), lg)

	if len(errors) > 0 {
		lg.Warn("Migration completed with errors", logger.Strings("errors", errors))
	}

	lg.Info("Bulk migration completed",
		logger.Int("total", len(databases)),
		logger.Int("successful", successCount),
		logger.Int("failed", errorCount),
		logger.String("duration", duration.String()))

	if errorCount > 0 {
		return fmt.Errorf("migration completed with %d errors out of %d databases", errorCount, len(databases))
	}

	lg.Info("All database migrations completed successfully")
	return nil
}

// executeSingleDatabaseMigration executes migration for a single database
func executeSingleDatabaseMigration(config *migrate_utils.MigrationConfig, lg *logger.Logger) error {
	// Display migration information for this database
	lg.Info("Preparing database migration",
		logger.String("source_database", config.SourceDBName),
		logger.String("target_database", config.TargetDBName),
		logger.String("source_host", fmt.Sprintf("%s:%d", config.SourceHost, config.SourcePort)),
		logger.String("target_host", fmt.Sprintf("%s:%d", config.TargetHost, config.TargetPort)))

	// Step 1: Backup target database (if exists)
	if config.BackupTarget {
		lg.Info("Starting target database backup", logger.String("database", config.TargetDBName))
		targetBackupFile, err := migrate_utils.BackupDatabaseForMigration(config, false, lg)

		if err != nil {
			lg.Warn("Failed to backup target database (may not exist)",
				logger.String("database", config.TargetDBName),
				logger.Error(err))
		} else {
			lg.Info("Target database backed up successfully",
				logger.String("database", config.TargetDBName),
				logger.String("backup_file", targetBackupFile))
		}
	}

	// Step 2: Backup source database
	lg.Info("Starting source database backup", logger.String("database", config.SourceDBName))
	sourceBackupFile, err := migrate_utils.BackupDatabaseForMigration(config, true, lg)
	if err != nil {
		return fmt.Errorf("failed to backup source database: %w", err)
	}
	lg.Info("Source database backed up successfully",
		logger.String("database", config.SourceDBName),
		logger.String("backup_file", sourceBackupFile))

	// Step 3: Restore source backup to target
	lg.Info("Starting restore to target database",
		logger.String("source_file", sourceBackupFile),
		logger.String("target_database", config.TargetDBName))
	err = restoreSelectionToTarget(config, sourceBackupFile, lg)
	if err != nil {
		return fmt.Errorf("failed to restore to target: %w", err)
	}
	lg.Info("Database restored to target successfully",
		logger.String("target_database", config.TargetDBName))

	return nil
}

// restoreToTarget restores the source backup to the target database
func restoreSelectionToTarget(config *migrate_utils.MigrationConfig, sourceBackupFile string, lg *logger.Logger) error {
	// Create restore options for target database
	restoreOptions := restoreUtils.RestoreOptions{
		Host:           config.TargetHost,
		Port:           config.TargetPort,
		User:           config.TargetUser,
		Password:       config.TargetPassword,
		DBName:         config.TargetDBName,
		File:           sourceBackupFile,
		VerifyChecksum: config.VerifyData,
	}

	// Perform restore using existing restore functionality
	err := single.RestoreSingle(restoreOptions)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	migrate_utils.AddCommonMigrationFlags(SelectionMigrateCmd)

	// Add specific flag for database list
	SelectionMigrateCmd.Flags().String("db_list", "", "path to text file containing list of database names (optional, will show selection if not provided)")
}
