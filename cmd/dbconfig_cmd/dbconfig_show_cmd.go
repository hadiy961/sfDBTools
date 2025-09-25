package dbconfig_cmd

import (
	"os"

	"sfDBTools/internal/core/dbconfig/show"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common/flags"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

var ShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show database configuration from encrypted files",
	Long: `Show database configuration from encrypted files.
If no file is specified, it will list all available encrypted config files
and allow you to choose one. Database password will be displayed in plain text.
You will always be prompted for the encryption password (environment variables are ignored for security).`,
	Run: func(cmd *cobra.Command, args []string) {
		// Clear screen and show header
		terminal.Headers("Show Database Configuration")

		if err := executeShow(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("Failed to show config", logger.Error(err))
			terminal.PrintError("Show operation failed")
			terminal.WaitForEnterWithMessage("Press Enter to continue...")
			os.Exit(1)
		}
	},
}

func executeShow(cmd *cobra.Command) error {
	// Resolve configuration with interactive selection if needed
	config, err := dbconfig.ResolveConfigWithInteractiveSelection(cmd)
	if err != nil {
		return err
	}

	// Execute show operation
	return show.ProcessShow(config)
}

func init() {
	// Add shared flags
	flags.AddCommonDbConfigFlags(ShowCmd)
}
