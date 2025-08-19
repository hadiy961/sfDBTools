package command_backup

import (
	"fmt"
	"os"

	"sfDBTools/internal/config"
	backup_all_databases_mysqldump "sfDBTools/internal/core/backup/all_databases/mysqldump"
	"sfDBTools/internal/logger"
	backup_utils "sfDBTools/utils/backup"

	"github.com/spf13/cobra"
)

var BackupAllDatabasesCmd = &cobra.Command{
	Use:   "all",
	Short: "Backup all databases into a single file",
	Long: `This command backs up all databases from a MySQL/MariaDB server into a single backup file.
It will dump all user databases (excluding system databases like information_schema, performance_schema, mysql, sys)
into one consolidated backup file with proper separation and comments.

Features:
- Single file output containing all databases
- Compression and encryption support
- Proper database separation with comments
- Skip system databases automatically
- Metadata generation with backup details`,

	Example: `# Backup all databases with default settings
sfDBTools backup all --source_host localhost --source_user root

# Backup with custom output directory and compression
sfDBTools backup all --source_host localhost --source_user root --output-dir ./backups --compress --compression gzip

# Backup with encryption enabled
sfDBTools backup all --source_host localhost --source_user root --encrypt

# Backup schema only (no data)
sfDBTools backup all --source_host localhost --source_user root --data=false`,

	Annotations: map[string]string{
		"command":  "backup",
		"category": "backup",
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeAllDatabasesBackup(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("All databases backup failed", logger.Error(err))
			os.Exit(1)
		}
	},
}

// executeAllDatabasesBackup handles the main all databases backup execution logic
func executeAllDatabasesBackup(cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	lg.Info("Starting all databases backup process")

	// Execute the all databases backup workflow
	return backup_utils.ExecuteAllDatabasesBackup(cmd, backup_all_databases_mysqldump.BackupAllDatabases)
}

func init() {
	// Add all common backup flags
	backup_utils.AddCommonBackupFlags(BackupAllDatabasesCmd)

	// Additional backup options specific to all databases backup
	_, _, _, _,
		_, _, _, _,
		_, defaultVerifyDisk, defaultRetentionDays, defaultCalculateChecksum, _ := config.GetBackupDefaults()

	BackupAllDatabasesCmd.Flags().Bool("verify-disk", defaultVerifyDisk, "verify available disk space before backup")
	BackupAllDatabasesCmd.Flags().Int("retention-days", defaultRetentionDays, "retention period in days")
	BackupAllDatabasesCmd.Flags().Bool("calculate-checksum", defaultCalculateChecksum, "calculate SHA256 checksum of backup file")

	// Note: This command doesn't need database selection flags since it backs up all databases
	// source_db flag from AddCommonBackupFlags will be ignored in this context
}
