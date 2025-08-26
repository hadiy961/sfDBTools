package backup_restore

import (
	"fmt"
	"os"

	"sfDBTools/internal/logger"
	backup_restore_utils "sfDBTools/utils/backup_restore"

	"github.com/spf13/cobra"
)

var BackupRestoreProductionCmd = &cobra.Command{
	Use:   "prod_to_secondary",
	Short: "Backup and restore production databases to secondary within the same server",
	Long: `This command copies production databases to secondary databases within the same server.

Flow:
1. Find production databases (dbsf_nbc_{{acc}} and dbsf_nbc_{{acc}}_dmart)
2. Create target databases if not exist (dbsf_nbc_{{acc}}_secondary_{{target}} and dbsf_nbc_{{acc}}_secondary_{{target}}_dmart)
3. Check existing users (sfnbc_{{acc}}_admin, sfnbc_{{acc}}_fin, sfnbc_{{acc}}_user)
4. Set max_statement_time to 0
5. Backup production databases
6. Restore target databases from production backups
7. Grant privileges to existing users for target databases
8. Restore original max_statement_time

Database Naming:
- Production: dbsf_nbc_{{acc}}
- Production Dmart: dbsf_nbc_{{acc}}_dmart
- Target: dbsf_nbc_{{acc}}_secondary_{{target}}
- Target Dmart: dbsf_nbc_{{acc}}_secondary_{{target}}_dmart

User Pattern:
- sfnbc_{{acc}}_admin
- sfnbc_{{acc}}_fin
- sfnbc_{{acc}}_user`,

	Example: `# Copy production to training for dataon account
sfDBTools backup-restore production --target=training --acc=dataon

# Copy production to staging for client123 account
sfDBTools backup-restore production --target=staging --acc=client123 --config mydb.cnf.enc`,

	Annotations: map[string]string{
		"command":  "backup-restore",
		"category": "backup-restore",
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeBackupRestoreProduction(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("Backup restore production failed", logger.Error(err))
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// executeBackupRestoreProduction handles the main backup restore execution logic
func executeBackupRestoreProduction(cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	lg.Info("Starting backup restore production process")

	// Resolve configuration
	config, err := backup_restore_utils.ResolveBackupRestoreConfig(cmd)
	if err != nil {
		return fmt.Errorf("failed to resolve configuration: %w", err)
	}

	// Display configuration
	backup_restore_utils.DisplayBackupRestoreConfig(config)

	// Prompt for confirmation
	if err := backup_restore_utils.PromptBackupRestoreConfirmation(config); err != nil {
		lg.Info("Backup restore operation cancelled", logger.String("reason", err.Error()))
		return err
	}

	// Execute backup restore process
	if err := backup_restore_utils.ExecuteBackupRestoreProduction(config); err != nil {
		lg.Error("Backup restore operation failed", logger.Error(err))
		return fmt.Errorf("backup restore failed: %w", err)
	}

	lg.Info("Backup restore production process completed successfully")
	fmt.Println("✅ Backup restore completed successfully!")

	return nil
}

func init() {
	// Add flags for backup restore
	BackupRestoreProductionCmd.Flags().String("target", "", "target environment name (e.g., training, staging)")
	BackupRestoreProductionCmd.Flags().String("acc", "", "account name for database naming")
	BackupRestoreProductionCmd.Flags().String("config", "", "encrypted configuration file (.cnf.enc)")
	BackupRestoreProductionCmd.Flags().Bool("encrypt", false, "encrypt backup files (default: false)")
	BackupRestoreProductionCmd.Flags().Bool("dry-run", false, "show what would be done without executing")
	BackupRestoreProductionCmd.Flags().Bool("yes", false, "skip confirmation prompts")

	// Mark required flags
	BackupRestoreProductionCmd.MarkFlagRequired("target")
	BackupRestoreProductionCmd.MarkFlagRequired("acc")
}
