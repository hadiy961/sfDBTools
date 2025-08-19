package cmd

import (
	command_restore "sfDBTools/cmd/restore"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

var RestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore main command",
	Long:  "Restore command allows you to restore database backups.",
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.Get()
		if err != nil {
			return
		}
		lg.Info("Restore command executed")
		cmd.Help()
	},
	Example: `sfDBTools restore single --help`,
	Annotations: map[string]string{
		"command":  "restore",
		"category": "restore",
	},
}

func init() {
	rootCmd.AddCommand(RestoreCmd)
	RestoreCmd.AddCommand(command_restore.SingleRestoreCmd)
	RestoreCmd.AddCommand(command_restore.RestoreUserCMD)
}
