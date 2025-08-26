package dbconfig_cmd

import (
	"fmt"
	"os"
	"time"

	"sfDBTools/cmd/dbconfig/generate"
	"sfDBTools/internal/logger"
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
		terminal.ClearAndShowHeader("🔐 Generate Database Configuration")

		// Show generation info
		terminal.PrintSubHeader("📋 Configuration Generation")
		terminal.PrintInfo("This will create an encrypted database configuration file.")
		terminal.PrintInfo("You'll be prompted for database connection details.")
		fmt.Println()

		// Show spinner while preparing
		spinner := terminal.NewProgressSpinner("Preparing configuration generator...")
		spinner.Start()

		// Brief delay to show preparation
		time.Sleep(500 * time.Millisecond)

		// Stop spinner before interactive generation
		spinner.Stop()
		fmt.Println()

		if err := generate.GenerateEncryptedConfig(cmd, configName, dbHost, dbPort, dbUser, autoMode); err != nil {
			lg, _ := logger.Get()
			lg.Error("Failed to generate encrypted config", logger.Error(err))
			terminal.PrintError(fmt.Sprintf("Failed to generate configuration: %v", err))
			terminal.WaitForEnterWithMessage("Press Enter to continue...")
			os.Exit(1)
		}

		terminal.PrintSuccess("✅ Configuration generated successfully!")
		terminal.WaitForEnterWithMessage("Press Enter to continue...")
	},
}

// Flags for automation
var (
	configName string
	dbHost     string
	dbPort     int
	dbUser     string
	autoMode   bool
)

func init() {
	GenerateCmd.Flags().StringVarP(&configName, "name", "n", "", "Configuration name (without extension)")
	GenerateCmd.Flags().StringVarP(&dbHost, "host", "H", "", "Database host")
	GenerateCmd.Flags().IntVarP(&dbPort, "port", "p", 0, "Database port")
	GenerateCmd.Flags().StringVarP(&dbUser, "user", "u", "", "Database username")
	GenerateCmd.Flags().BoolVarP(&autoMode, "auto", "a", false, "Auto mode - skip confirmations")
}
