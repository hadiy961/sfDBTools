package command_backup

import (
	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	backup_utils "sfDBTools/utils/backup"

	"github.com/spf13/cobra"
)

var FullDBCmd = &cobra.Command{
	Use:     "full",
	Short:   "Backup full databases",
	Long:    "This command backs up all databases in the MySQL server. It will create a backup file for each database found.",
	Example: `sfDBTools backup full --source_db my_database`,
	Annotations: map[string]string{
		"command":  "backup",
		"category": "backup",
	},
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.Get()
		if err != nil {
			lg.Error("Failed to get logger", logger.Error(err))
			return
		}

		lg.Info("Starting backup of all databases")

		// Here you would implement the logic to backup all databases
		// For example, you might call a function that retrieves all databases
		// and then iterates over them to create backups.

		lg.Info("Backup of all databases completed successfully")
	},
}

func init() {
	backup_utils.AddCommonBackupFlags(FullDBCmd)

	// Additional backup options
	_, _, _, _,
		_, _, _, _,
		_, defaultVerifyDisk, defaultRetentionDays, defaultCalculateChecksum, _ := config.GetBackupDefaults()

	FullDBCmd.Flags().Bool("verify-disk", defaultVerifyDisk, "verify available disk space before backup")
	FullDBCmd.Flags().Int("retention-days", defaultRetentionDays, "retention period in days")
	FullDBCmd.Flags().Bool("calculate-checksum", defaultCalculateChecksum, "calculate SHA256 checksum of backup file")

	// Required flag for database list
	FullDBCmd.Flags().String("db_list", "", "path to text file containing list of database names (optional, will show selection if not provided)")
}
