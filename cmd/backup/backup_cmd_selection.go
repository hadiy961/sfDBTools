package command_backup

import (
	"fmt"
	"os"

	"sfDBTools/internal/config"
	backup_single_mysqldump "sfDBTools/internal/core/backup/single/mysqldump"
	"sfDBTools/internal/logger"
	backup_utils "sfDBTools/utils/backup"
	"sfDBTools/utils/database/info"

	"github.com/spf13/cobra"
)

var BackupSelectionCmd = &cobra.Command{
	Use:   "selection",
	Short: "Backup databases using interactive selection, specific database, or from a list file",
	Long: `This command allows you to backup databases in three ways:
1. Interactive selection: Choose multiple databases from an interactive list
2. Specific database: Use --source_db to backup a single database
3. From file: Use --db_list to backup databases listed in a text file

Note: --source_db and --db_list flags are mutually exclusive.`,
	Example: `# Interactive selection
sfDBTools backup selection --source_host localhost --source_user root

# Specific database
sfDBTools backup selection --config ./config/mydb.cnf.enc --source_db mydb

# From database list file
sfDBTools backup selection --config ./config/mydb.cnf.enc --db_list ./databases.txt`,
	Run: func(cmd *cobra.Command, args []string) {
		lg, _ := logger.Get()

		// Get the values of both flags
		dbListPath, err := cmd.Flags().GetString("db_list")
		if err != nil {
			lg.Error("Failed to get db_list flag", logger.Error(err))
			os.Exit(1)
		}

		sourceDB, err := cmd.Flags().GetString("source_db")
		if err != nil {
			lg.Error("Failed to get source_db flag", logger.Error(err))
			os.Exit(1)
		}

		// Check for mutually exclusive flags
		if dbListPath != "" && sourceDB != "" {
			lg.Error("Cannot use both --source_db and --db_list flags simultaneously")
			fmt.Printf("Error: Cannot use both --source_db and --db_list flags simultaneously\n")
			os.Exit(1)
		}

		// Route to appropriate execution function
		if dbListPath != "" {
			// Execute list backup from file
			if err := executeListBackup(cmd); err != nil {
				lg.Error("List backup failed", logger.Error(err))
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Execute selection backup (either specific DB or interactive selection)
			if err := executeSelectionBackup(cmd); err != nil {
				lg.Error("Selection backup failed", logger.Error(err))
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

// executeSelectionBackup handles the main selection backup execution logic
func executeSelectionBackup(cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	lg.Info("Starting selection backup process")

	// 1. Resolve backup configuration without database
	backupConfig, err := backup_utils.ResolveBackupConfigWithoutDB(cmd)
	if err != nil {
		return fmt.Errorf("failed to resolve backup configuration: %w", err)
	}

	// 2. Create database config and test connection
	dbConfig := backup_utils.CreateDatabaseConfig(backupConfig)
	if err := backup_utils.TestDatabaseConnection(dbConfig); err != nil {
		return err
	}

	// 3. Get databases - either from source_db flag or interactive selection
	var selectedDatabases []string

	// Check if source_db flag is provided
	sourceDB, err := cmd.Flags().GetString("source_db")
	if err != nil {
		return fmt.Errorf("failed to get source_db flag: %w", err)
	}

	if sourceDB != "" {
		// If source_db is specified, use it directly
		selectedDatabases = []string{sourceDB}
		lg.Info("Using database from source_db flag",
			logger.String("database", sourceDB))
	} else {
		// Otherwise, let user select multiple databases interactively
		selectedDatabases, err = info.SelectMultipleDatabasesInteractive(dbConfig)
		if err != nil {
			return fmt.Errorf("failed to select databases: %w", err)
		}

		if len(selectedDatabases) == 0 {
			return fmt.Errorf("no databases selected for backup")
		}
	}

	lg.Info("Databases selected for backup",
		logger.Int("count", len(selectedDatabases)),
		logger.Strings("databases", selectedDatabases))

	// 4. Execute backup for selected databases
	_, err = backup_utils.ExecuteMultipleDatabaseBackup(
		backupConfig,
		selectedDatabases,
		backup_single_mysqldump.BackupSingle,
		"Selection",
	)

	return err
}

// executeListBackup handles the main list backup execution logic
func executeListBackup(cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	lg.Info("Starting list backup process")

	// Execute the complete list backup workflow using the reusable utility
	return backup_utils.ExecuteListBackupWorkflow(cmd, backup_single_mysqldump.BackupSingle)
}

func init() {
	backup_utils.AddCommonBackupFlags(BackupSelectionCmd)

	// Additional backup options
	_, _, _, _,
		_, _, _, _,
		_, defaultVerifyDisk, defaultRetentionDays, defaultCalculateChecksum, _ := config.GetBackupDefaults()

	BackupSelectionCmd.Flags().Bool("verify-disk", defaultVerifyDisk, "verify available disk space before backup")
	BackupSelectionCmd.Flags().Int("retention-days", defaultRetentionDays, "retention period in days")
	BackupSelectionCmd.Flags().Bool("calculate-checksum", defaultCalculateChecksum, "calculate SHA256 checksum of backup file")

	// Required flag for database list
	BackupSelectionCmd.Flags().String("db_list", "", "path to text file containing list of database names (optional, will show selection if not provided)")
}
