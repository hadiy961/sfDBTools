package cmd

import (
	system_cmd "sfDBTools/cmd/system_cmd"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

var SystemCmd = &cobra.Command{
	Use:   "system",
	Short: "System-related utility commands",
	Long:  "System-related utility commands such as disk checks.",
	Run: func(cmd *cobra.Command, args []string) {
		lg, _ := logger.Get()
		lg.Info("System command executed")
		cmd.Help()
	},
	Annotations: map[string]string{
		"command":  "system",
		"category": "system",
	},
}

func init() {
	rootCmd.AddCommand(SystemCmd)
	SystemCmd.AddCommand(system_cmd.SystemDiskCmd)
	SystemCmd.AddCommand(system_cmd.SystemDiskMonitorCmd)
	SystemCmd.AddCommand(system_cmd.SystemDiskListCmd)
	SystemCmd.AddCommand(system_cmd.SystemStorageMonitorCmd)
}
