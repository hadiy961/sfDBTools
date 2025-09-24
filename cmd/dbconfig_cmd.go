package cmd

import (
	"sfDBTools/cmd/dbconfig_cmd"
	"sfDBTools/internal/core/menu"

	"github.com/spf13/cobra"
)

var DBConfigCMD = &cobra.Command{
	Use:   "dbconfig",
	Short: "Database configuration management commands",
	Long: `Database configuration management commands for generating, validating, editing, viewing, and deleting encrypted database configurations.
All database configurations are stored in encrypted format for security.`,
	Run: func(cmd *cobra.Command, args []string) {
		// use the cfg/lg provided to the dbconfig_cmd package
		menu.DBConfigMenu(dbconfig_cmd.Lg, dbconfig_cmd.Cfg)
	},
}

func init() {
	rootCmd.AddCommand(DBConfigCMD)
	DBConfigCMD.AddCommand(dbconfig_cmd.GenerateCmd)
	DBConfigCMD.AddCommand(dbconfig_cmd.ValidateCmd)
	DBConfigCMD.AddCommand(dbconfig_cmd.ShowCmd)
	DBConfigCMD.AddCommand(dbconfig_cmd.EditCmd)
	DBConfigCMD.AddCommand(dbconfig_cmd.DeleteCmd)
}
