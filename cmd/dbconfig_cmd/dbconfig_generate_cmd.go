package dbconfig_cmd

import (
	"fmt"
	"os"
	"time"

	"sfDBTools/internal/core/dbconfig/generate"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate encrypted database configuration",
	Long: `Generate encrypted database configuration file.
This command will prompt for database connection details and encrypt them
using the application configuration values (client_code, app_name, version, author)
combined with an encryption password from environment variable or interactive input.

For automation, you can use flags to provide database parameters:
--name, --host, --port, --user

Sensitive data (passwords) must be provided via environment variables:
- Encryption password: SFDB_ENCRYPTION_PASSWORD
- Database password: SFDB_PASSWORD

If environment variables are not set, you will be prompted interactively.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Clear screen and show header
		terminal.ClearAndShowHeader("Generate database configuration")

		// Show generation info
		terminal.PrintSubHeader("Configuration generation")
		terminal.PrintInfo("This will create an encrypted database configuration file.")
		terminal.PrintInfo("You'll be prompted for database connection details.")
		fmt.Println()

		// Show spinner while preparing
		spinner := terminal.NewProgressSpinner("Preparing configuration generator...")
		spinner.Start()
		time.Sleep(500 * time.Millisecond)
		spinner.Stop()
		fmt.Println()

		if err := executeGenerate(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("Failed to generate encrypted config", logger.Error(err))
			terminal.PrintError("Generation failed")
			terminal.WaitForEnterWithMessage("Press Enter to continue...")
			os.Exit(1)
		}

		terminal.PrintSuccess("Configuration generated successfully!")
		terminal.WaitForEnterWithMessage("Press Enter to continue...")
	},
}

func executeGenerate(cmd *cobra.Command) error {
	// Resolve configuration from flags
	config, err := dbconfig.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	// Execute generate operation
	return generate.ProcessGenerate(config)
}

func init() {
	// Add shared and generate-specific flags
	dbconfig.AddCommonDbConfigFlags(GenerateCmd)
	dbconfig.AddGenerateFlags(GenerateCmd)
}
