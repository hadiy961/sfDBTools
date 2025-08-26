package cmd

import (
	backup_cmd "sfDBTools/cmd/backup_cmd"
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
	BackupCmd.AddCommand(backup_cmd.BackupAllDatabasesCmd)
	BackupCmd.AddCommand(backup_cmd.BackupSelectionCmd)
	BackupCmd.AddCommand(backup_cmd.BackupUserCMD)
}
