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
	Short: "Backup all databases into a single file with flexible system database inclusion",
	Long: `This command backs up all databases from a MySQL/MariaDB server into a single backup file.
You can choose to include or exclude system databases and system users for replication purposes.

Features:
- Single file output containing all databases
- Flexible system database inclusion/exclusion
- User grants backup in separate file (uses SHOW GRANTS method)
- GTID information capture for replication
- Compression and encryption support
- Proper database separation with comments
- Metadata generation with backup details`,

	Example: `# Backup all user databases (exclude system databases - default)
sfDBTools backup all --source_host localhost --source_user root

# Backup all databases including system databases (for full server backup)
sfDBTools backup all --source_host localhost --source_user root --include-system-databases

# Backup for replication setup (includes user grants in separate file + GTID info)
sfDBTools backup all --source_host localhost --source_user root --include-system-databases --include-user

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

	// New flags for system database and user inclusion
	BackupAllDatabasesCmd.Flags().Bool("include-system-databases", false, "include system databases (mysql, information_schema, performance_schema, sys)")
	BackupAllDatabasesCmd.Flags().Bool("include-user", false, "include user grants in separate file (uses SHOW GRANTS method)")
	BackupAllDatabasesCmd.Flags().Bool("capture-gtid", true, "capture GTID information for replication (includes BINLOG_GTID_POS)")

	// Note: This command doesn't need database selection flags since it backs up all databases
	// source_db flag from AddCommonBackupFlags will be ignored in this context
}
