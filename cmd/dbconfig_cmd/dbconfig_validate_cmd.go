package dbconfig_cmd

import (
	"os"

	"sfDBTools/internal/core/dbconfig/validate"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

var ValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate encrypted database configuration and test connection",
	Long: `Validate that the encrypted database configuration can be properly decrypted
and test the actual database connection. If no file is specified, it will list all 
available encrypted config files and allow you to choose one.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Clear screen and show header
		terminal.ClearAndShowHeader("✅ Validate Database Configuration")

		if err := executeValidate(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("Failed to validate config", logger.Error(err))
			terminal.PrintError("Validation failed")
			terminal.WaitForEnterWithMessage("Press Enter to continue...")
			os.Exit(1)
		}
	},
}

func executeValidate(cmd *cobra.Command) error {
	// Resolve configuration with interactive selection if needed
	config, err := dbconfig.ResolveConfigWithInteractiveSelection(cmd)
	if err != nil {
		return err
	}

	// Execute validation operation
	return validate.ProcessValidate(config)
}

func init() {
	// Add shared flags
	dbconfig.AddCommonDbConfigFlags(ValidateCmd)
}
