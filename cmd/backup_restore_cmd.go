package cmd

import (
	command_backup_restore "sfDBTools/cmd/backup_restore"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

var BackupRestoreCmd = &cobra.Command{
	Use:   "backup-restore",
	Short: "Backup and restore databases within the same server",
	Long:  "Backup and restore command allows you to copy databases from production to secondary within the same server.",
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.Get()
		if err != nil {
			lg.Error("Failed to get logger", logger.Error(err))
			return
		}
		lg.Info("Backup restore command executed")
		cmd.Help()
	},
	Annotations: map[string]string{
		"command":  "backup-restore",
		"category": "backup-restore",
	},
}

func init() {
	rootCmd.AddCommand(BackupRestoreCmd)
	BackupRestoreCmd.AddCommand(command_backup_restore.BackupRestoreProductionCmd)
}
