package cmd

import (
	command_mariadb "sfDBTools/cmd/mariadb"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

var MariaDBCmd = &cobra.Command{
	Use:   "mariadb",
	Short: "MariaDB/MySQL management and diagnostics commands",
	Long:  "MariaDB command provides health checks, diagnostics, and management utilities for MariaDB/MySQL servers.",
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.Get()
		if err != nil {
			lg.Error("Failed to get logger", logger.Error(err))
			return
		}
		lg.Info("MariaDB command executed")
		cmd.Help()
	},
	Annotations: map[string]string{
		"command":  "mariadb",
		"category": "database",
	},
}

func init() {
	rootCmd.AddCommand(MariaDBCmd)
	MariaDBCmd.AddCommand(command_mariadb.UninstallCmd)
	MariaDBCmd.AddCommand(command_mariadb.InstallCmd)
	MariaDBCmd.AddCommand(command_mariadb.ConfigureCmd)
	MariaDBCmd.AddCommand(command_mariadb.VersionsCmd)
}
