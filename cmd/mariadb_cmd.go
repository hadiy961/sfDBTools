package cmd

import (
	mariadb_cmd "sfDBTools/cmd/mariadb_cmd"
	"sfDBTools/internal/core/menu"

	"github.com/spf13/cobra"
)

var MariaDBCmd = &cobra.Command{
	Use:   "mariadb",
	Short: "Database configuration management commands",
	Long: `Database configuration management commands for generating, validating, editing, viewing, and deleting encrypted database configurations.
All database configurations are stored in encrypted format for security.`,
	Run: func(cmd *cobra.Command, args []string) {
		menu.MariaDBMenu(mariadb_cmd.Lg, mariadb_cmd.Cfg)
	},
}

func init() {
	rootCmd.AddCommand(MariaDBCmd)
	MariaDBCmd.AddCommand(mariadb_cmd.Check)
	MariaDBCmd.AddCommand(mariadb_cmd.ConfigureMariadbCMD)
	MariaDBCmd.AddCommand(mariadb_cmd.InstallCmd)
	MariaDBCmd.AddCommand(mariadb_cmd.RemoveCmd)
}
