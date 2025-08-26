package cmd

import (
	mariadb_cmd "sfDBTools/cmd/mariadb_cmd"
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
		lg.Info("MariaDB command executed")
		cmd.Help()
	},
	Annotations: map[string]string{
		"command":  "mariadb",
		"category": "mariadb",
	},
}

func init() {
	rootCmd.AddCommand(MariaDBCMD)
	// MariaDB-specific commands
	MariaDBCMD.AddCommand(mariadb_cmd.CheckVersionCmd)
	MariaDBCMD.AddCommand(mariadb_cmd.InstallCmd)
	MariaDBCMD.AddCommand(mariadb_cmd.RemoveCmd)
	MariaDBCMD.AddCommand(mariadb_cmd.CheckConfigCmd)
	// TODO: Add more commands as they are implemented
	// MariaDBCMD.AddCommand(mariadb_cmd.TuneConfigCmd)
	// MariaDBCMD.AddCommand(mariadb_cmd.MonitorCmd)
	// MariaDBCMD.AddCommand(mariadb_cmd.EditConfigCmd)
}
