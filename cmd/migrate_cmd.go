package cmd

import (
	command_migrate "sfDBTools/cmd/migrate"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

var MigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands",
	Long:  "Database migration commands for transferring data from source to target database with backup safety.",
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.Get()
		if err != nil {
			lg.Error("Failed to get logger", logger.Error(err))
			return
		}
		lg.Info("Migrate command executed")
		cmd.Help()
	},
	Annotations: map[string]string{
		"command":  "migrate",
		"category": "migration",
	},
}

func init() {
	rootCmd.AddCommand(MigrateCmd)
	MigrateCmd.AddCommand(command_migrate.SelectionMigrateCmd)
}
