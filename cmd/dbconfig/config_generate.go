package dbconfig

import (
	"fmt"
	"os"

	"sfDBTools/cmd/dbconfig/generate"
	"sfDBTools/internal/logger"

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
		if err := generate.GenerateEncryptedConfig(cmd, configName, dbHost, dbPort, dbUser, autoMode); err != nil {
			lg, _ := logger.Get()
			lg.Error("Failed to generate encrypted config", logger.Error(err))
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
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
