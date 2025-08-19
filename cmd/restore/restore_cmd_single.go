package command_restore

import (
	"fmt"
	"os"

	restore "sfDBTools/internal/core/restore/single"
	restoreUtils "sfDBTools/internal/core/restore/utils"
	"sfDBTools/internal/logger"
	restore_utils "sfDBTools/utils/restore"

	"github.com/spf13/cobra"
)

var SingleRestoreCmd = &cobra.Command{
	Use:   "single",
	Short: "Restore a single backup file",
	Long:  `This command allows you to restore a single backup file with various options for database connection and validation.`,
	Example: `sfDBTools restore single --config ./config/mydb.cnf.enc --file ./backup/database_backup.sql.gz
sfDBTools restore single --target_db my_database --target_host localhost --target_port 3306 --target_user root --target_password my_password --file ./backup/database_backup.sql.gz
sfDBTools restore single --target_host localhost --target_user root --file ./backup/database_backup.sql.gz  # Will prompt for database selection
sfDBTools restore single --target_host localhost --target_user root  # Will prompt for backup file and database selection
sfDBTools restore single  # Fully interactive - will prompt for everything

# Create new database options:
sfDBTools restore single --create-new-db --file ./backup/database_backup.sql.gz  # Create new database with manual name input
sfDBTools restore single --create-new-db --db-from-filename --file ./backup/database_backup.sql.gz  # Create new database using name from filename
sfDBTools restore single --target_host localhost --target_user root --create-new-db  # Interactive mode with new database option`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeRestore(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("Restore failed", logger.Error(err))
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// executeRestore handles the main restore execution logic
func executeRestore(cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	lg.Info("Starting restore process")

	// Resolve restore configuration from various sources
	restoreConfig, err := restore_utils.ResolveRestoreConfig(cmd)
	if err != nil {
		return fmt.Errorf("failed to resolve restore configuration: %w", err)
	}

	// Convert to RestoreOptions for backward compatibility
	options := restoreConfig.ToRestoreOptions()

	// Display parameters before execution
	restore_utils.DisplayRestoreParameters(options)

	// Prompt for confirmation before proceeding
	if err := restore_utils.PromptRestoreConfirmation(options); err != nil {
		lg.Info("Restore operation cancelled", logger.String("reason", err.Error()))
		return err
	}

	// Convert to internal RestoreOptions for backward compatibility
	internalOptions := restoreUtils.RestoreOptions{
		Host:           options.Host,
		Port:           options.Port,
		User:           options.User,
		Password:       options.Password,
		DBName:         options.DBName,
		File:           options.File,
		VerifyChecksum: options.VerifyChecksum,
	}

	// Perform the restore
	if err := restore.RestoreSingle(internalOptions); err != nil {
		lg.Error("Restore operation failed", logger.Error(err))
		return fmt.Errorf("restore failed: %w", err)
	}

	lg.Info("Restore process completed successfully")
	fmt.Println("✅ Restore completed successfully!")

	return nil
}

func init() {
	restore_utils.AddCommonRestoreFlags(SingleRestoreCmd)
}
