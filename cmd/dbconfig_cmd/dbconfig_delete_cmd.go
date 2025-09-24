package dbconfig_cmd

import (
	"os"

	"sfDBTools/internal/core/dbconfig/delete"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete encrypted database configuration files",
	Long: `Delete encrypted database configuration files.
⚠️  WARNING: Deleted files cannot be recovered. Use with caution.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeDelete(cmd, args); err != nil {
			lg, _ := logger.Get()
			lg.Error("Failed to delete config", logger.Error(err))
			terminal.PrintError("Delete operation failed")
			os.Exit(1)
		}
	},
}

func executeDelete(cmd *cobra.Command, args []string) error {
	// Resolve configuration from flags and arguments
	config, err := dbconfig.ResDBConfigFlag(cmd)
	if err != nil {
		return err
	}

	// Execute delete operation
	return delete.ProcessDelete(config, args)
}

func init() {
	// Add shared and delete-specific flags
	dbconfig.AddCommonDbConfigFlags(DeleteCmd)
	dbconfig.AddDeleteFlags(DeleteCmd)
}
