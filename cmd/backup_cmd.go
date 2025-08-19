package cmd

import (
	command_backup "sfDBTools/cmd/backup"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup main command",
	Long:  "Backup command allows you to backup databases, tables, and grants.",
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.Get()
		if err != nil {
			lg.Error("Failed to get logger", logger.Error(err))
			return
		}
		lg.Info("Backup command executed")
		cmd.Help()
	},
	Annotations: map[string]string{
		"command":  "backup",
		"category": "backup",
	},
}

func init() {
	rootCmd.AddCommand(BackupCmd)
	BackupCmd.AddCommand(command_backup.BackupAllDatabasesCmd)
	BackupCmd.AddCommand(command_backup.FullDBCmd)
	BackupCmd.AddCommand(command_backup.BackupSelectionCmd)
	BackupCmd.AddCommand(command_backup.BackupUserCMD)
}
