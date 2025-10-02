package dbconfig_cmd

import (
	"os"

	"sfDBTools/internal/core/dbconfig/generate"
	"sfDBTools/utils/common/flags"
	"sfDBTools/utils/common/parsing"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate encrypted database configuration",
	Long: `Generate encrypted database configuration file.
This command will prompt for database connection details and encrypt them
using the password you provide.

For automation, you can use flags to provide database parameters:
--name, --host, --port, --user

Sensitive data (passwords) must be provided via environment variables:
- Encryption password: SFDB_ENCRYPTION_PASSWORD
- Database password: SFDB_DB_PASSWORD

If environment variables are not set, you will be prompted interactively.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := execDBConfigGenerate(cmd); err != nil {
			terminal.PrintError("Generation failed")
			terminal.WaitForEnterWithMessage("Press Enter to continue...")
			os.Exit(1)
		} else {
			terminal.PrintSuccess("Generation completed successfully")
			terminal.WaitForEnterWithMessage("Press Enter to continue...")
			return
		}
	},
}

func execDBConfigGenerate(cmd *cobra.Command) error {
	// Resolve configuration from flags
	DBConfig, err := parsing.ParseDBConfigGenerate(cmd)
	if err != nil {
		return err
	}
	// Execute generate operation
	return generate.ProcessGenerate(DBConfig, Lg)
}

func init() {
	// Add shared and generate-specific flags
	flags.AddGenerateFlags(GenerateCmd)
}
