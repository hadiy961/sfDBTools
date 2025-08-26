package cmd

import (
	command_config "sfDBTools/cmd/dbconfig"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

var MariaDBCMD = &cobra.Command{
	Use:   "mariadb",
	Short: "MariaDB management commands",
	Long: `MariaDB management commands for generating, validating, editing, viewing, and deleting encrypted database configurations.
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
	rootCmd.AddCommand(MariaDBCMD)
	MariaDBCMD.AddCommand(command_config.GenerateCmd)
	MariaDBCMD.AddCommand(command_config.ValidateCmd)
	MariaDBCMD.AddCommand(command_config.ShowCmd)
	MariaDBCMD.AddCommand(command_config.EditCmd)
	MariaDBCMD.AddCommand(command_config.DeleteCmd)
}
