package cmd

import (
	mariadb_cmd "sfDBTools/cmd/mariadb_cmd"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)


var MariaDBCmd = &cobra.Command{
	Use:   "mariadb",
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
	rootCmd.AddCommand(MariaDBCmd)
	MariaDBCmd.AddCommand(mariadb_cmd.CheckVersionCmd)
	MariaDBCmd.AddCommand(mariadb_cmd.ConfigureMariadbCMD)
	MariaDBCmd.AddCommand(mariadb_cmd.InstallCmd)
	MariaDBCmd.AddCommand(mariadb_cmd.RemoveCmd)
}
