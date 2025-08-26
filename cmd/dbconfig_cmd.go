package cmd

import (
	config_cmd "sfDBTools/cmd/dbconfig_cmd"
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
	ConfigCmd.AddCommand(config_cmd.GenerateCmd)
	ConfigCmd.AddCommand(config_cmd.ValidateCmd)
	ConfigCmd.AddCommand(config_cmd.ShowCmd)
	ConfigCmd.AddCommand(config_cmd.EditCmd)
	ConfigCmd.AddCommand(config_cmd.DeleteCmd)
}
