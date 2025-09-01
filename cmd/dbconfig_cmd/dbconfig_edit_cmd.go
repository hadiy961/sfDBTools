package dbconfig_cmd

import (
	"os"

	"sfDBTools/internal/core/dbconfig/edit"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

var EditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit encrypted database configuration",
	Long: `Edit encrypted database configuration file.
If no file is specified, it will list all available encrypted config files
and allow you to choose one. You can modify name, host, port, user, and password.
If the name changes, the file will be renamed accordingly.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Clear screen and show header
		terminal.ClearAndShowHeader("✏️ Edit Database Configuration")

		if err := executeEdit(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("Failed to edit config", logger.Error(err))
			terminal.PrintError("Edit operation failed")
			terminal.WaitForEnterWithMessage("Press Enter to continue...")
			os.Exit(1)
		}

		terminal.PrintSuccess("✅ Configuration updated successfully!")
		terminal.WaitForEnterWithMessage("Press Enter to continue...")
	},
}

func executeEdit(cmd *cobra.Command) error {
	// Resolve configuration with interactive selection if needed
	config, err := dbconfig.ResolveConfigWithInteractiveSelection(cmd)
	if err != nil {
		return err
	}

	// Execute edit operation
	return edit.ProcessEdit(config)
}

func init() {
	// Add shared flags
	dbconfig.AddCommonDbConfigFlags(EditCmd)
}
