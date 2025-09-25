package dbconfig_cmd

import (
	"os"

	"sfDBTools/internal/core/dbconfig/edit"
	"sfDBTools/utils/common/flags"
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
		if err := executeEdit(cmd); err != nil {
			terminal.PrintError("Edit operation failed")
			terminal.WaitForEnterWithMessage("Press Enter to continue...")
			os.Exit(1)
		} else {
			terminal.PrintSuccess("Configuration updated successfully!")
			terminal.WaitForEnterWithMessage("Press Enter to continue...")
			return
		}
	},
}

func executeEdit(cmd *cobra.Command) error {
	// Resolve configuration with interactive selection if needed
	terminal.Headers("Edit Database Configuration")

	config, err := dbconfig.ResolveConfigWithInteractiveSelection(cmd)
	if err != nil {
		return err
	}

	// Execute edit operation
	return edit.ProcessEdit(config, Lg)
}

func init() {
	// Add shared flags
	flags.AddCommonDbConfigFlags(EditCmd)
}
