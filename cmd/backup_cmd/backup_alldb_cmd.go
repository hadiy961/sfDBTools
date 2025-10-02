package backup_cmd

import (
	"context"
	"fmt"
	"os"
	"sfDBTools/internal/core/backup"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common/flags"

	"github.com/spf13/cobra"
)

var BackupAllDBCmd = &cobra.Command{
	Use:   "all-new",
	Short: "Backup all databases into a single file with flexible system database inclusion",
	Long: `This command backs up all databases from a MySQL/MariaDB server into a single backup file.
You can choose to include or exclude system databases and system users for replication purposes.

Features:
- Single file output containing all databases
- Flexible system database inclusion/exclusion
- User grants backup in separate file (uses SHOW GRANTS method)
- GTID information capture for replication
- Compression and encryption support (uses same password as config encryption)
- Metadata generation with backup details

Encryption:
- When --encrypt is enabled, you will be prompted for an encryption password
- Use the same password as your encrypted configuration files (.cnf.enc)
- You can set the SFDB_ENCRYPTION_PASSWORD environment variable to avoid prompts
- This ensures consistency between config and backup encryption`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeAllDBBackup(cmd, Lg); err != nil {
			Lg.Error("All databases backup failed", logger.Error(err))
			os.Exit(1)
		}
	},
}

// executeAllDatabasesBackup handles the main all databases backup execution logic
func executeAllDBBackup(cmd *cobra.Command, lg *logger.Logger) error {
	if lg == nil {
		return fmt.Errorf("logger is not initialized")
	}
	lg.Info("Starting all databases backup process")

	// Execute the all databases backup workflow
	executor := backup.NewBackupExecutor(lg, cmd)
	backupResult, err := executor.AllDB(context.Background())

	if err != nil {
		// Tampilkan result parsial (misalnya, GTID yang berhasil ditangkap sebelum gagal)
		fmt.Println("Partial backup result:", backupResult) // Asumsi fungsi ini ada
		return err
	}

	// return backup_utils.ExecuteAllDatabasesBackup(cmd, mysqldump.BackupAllDatabases)
	return nil
}

func init() {
	flags.AddBackupAllDBFlags(BackupAllDBCmd)
}
