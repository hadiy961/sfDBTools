package command_restore

import (
	"fmt"
	"os"

	restoreUser "sfDBTools/internal/core/restore/user"
	"sfDBTools/internal/logger"
	restore_utils "sfDBTools/utils/restore"

	"github.com/spf13/cobra"
)

var RestoreUserCMD = &cobra.Command{
	Use:   "user",
	Short: "Restore user grants from backup file",
	Long:  `This command allows you to restore user grants from a backup file with various options for database connection and validation.`,
	Example: `sfDBTools restore user --config ./config/mydb.cnf.enc --file ./backup/grants/system_users_2025_08_11_140853.sql.gz.enc
sfDBTools restore user --target_host localhost --target_port 3306 --target_user root --target_password my_password --file ./backup/grants/database_grants_2025_08_11_140853.sql.gz.enc
sfDBTools restore user --target_host localhost --target_user root --file ./backup/grants/system_users_2025_08_11_140853.sql.gz.enc  # Will prompt for password if needed
sfDBTools restore user --target_host localhost --target_user root  # Will prompt for grant file selection
sfDBTools restore user  # Fully interactive - will prompt for everything`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeRestoreUser(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("User grants restore failed", logger.Error(err))
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// executeRestoreUser handles the main user grants restore execution logic
func executeRestoreUser(cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	lg.Info("Starting user grants restore process")

	// Resolve restore configuration from various sources
	restoreConfig, err := restore_utils.ResolveRestoreUserConfig(cmd)
	if err != nil {
		return fmt.Errorf("failed to resolve restore configuration: %w", err)
	}

	// Convert to RestoreUserOptions for processing
	options := restoreConfig.ToRestoreUserOptions()

	// Display parameters before execution
	restore_utils.DisplayRestoreUserParameters(options)

	// Prompt for confirmation before proceeding
	if err := restore_utils.PromptRestoreUserConfirmation(options); err != nil {
		lg.Info("User grants restore operation cancelled", logger.String("reason", err.Error()))
		return err
	}

	// Perform the restore
	if err := restoreUser.RestoreUserGrants(options); err != nil {
		lg.Error("User grants restore operation failed", logger.Error(err))
		return fmt.Errorf("user grants restore failed: %w", err)
	}

	lg.Info("User grants restore process completed successfully")
	fmt.Println("✅ User grants restore completed successfully!")

	return nil
}

func init() {
	restore_utils.AddCommonRestoreUserFlags(RestoreUserCMD)
}
