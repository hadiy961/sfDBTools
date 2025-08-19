package cmd

import (
	command_config "sfDBTools/cmd/config"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long:  "Configuration management commands for generating and managing encrypted database configurations.",
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.Get()
		if err != nil {
			lg.Error("Failed to get logger", logger.Error(err))
			return
		}
		lg.Info("Config command executed")
		cmd.Help()
	},
	Annotations: map[string]string{
		"command":  "config",
		"category": "configuration",
	},
}

func init() {
	rootCmd.AddCommand(ConfigCmd)
	ConfigCmd.AddCommand(command_config.GenerateCmd)
	ConfigCmd.AddCommand(command_config.ValidateCmd)
	ConfigCmd.AddCommand(command_config.ShowCmd)
	ConfigCmd.AddCommand(command_config.EditCmd)
}
