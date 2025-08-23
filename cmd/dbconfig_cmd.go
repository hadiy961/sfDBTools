package cmd

import (
	command_config "sfDBTools/cmd/dbconfig"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use:   "dbconfig",
	Short: "Database configuration management commands",
	Long: `Database configuration management commands for generating, validating, editing, viewing, and deleting encrypted database configurations.
All database configurations are stored in encrypted format for security.`,
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.Get()
		if err != nil {
			lg.Error("Failed to get logger", logger.Error(err))
			return
		}
		lg.Info("Database config command executed")
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
	ConfigCmd.AddCommand(command_config.DeleteCmd)
}
