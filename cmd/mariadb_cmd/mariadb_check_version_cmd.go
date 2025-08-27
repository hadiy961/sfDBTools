package mariadb_cmd

import (
	"fmt"
	"os"

	"sfDBTools/internal/core/mariadb/check_version"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

// CheckVersionCmd represents the check_version command
var CheckVersionCmd = &cobra.Command{
	Use:   "check_version",
	Short: "Check available MariaDB versions",
	Long: `Check available MariaDB versions starting from 10.6 and above.
Shows only major version latest stable releases with EOL dates.
Only works on supported operating systems: CentOS, Ubuntu, RHEL, Rocky, AlmaLinux.`,
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.Get()
		if err != nil {
			terminal.PrintError("Failed to initialize logger")
			os.Exit(1)
		}

		// Create runner with default configuration
		config := check_version.DefaultCheckVersionConfig()
		runner := check_version.NewCheckVersionRunner(config)

		if err := runner.Run(); err != nil {
			lg.Error("Failed to check MariaDB versions", logger.Error(err))
			terminal.PrintError(fmt.Sprintf("Error: %v", err))
			os.Exit(1)
		}
	},
	Annotations: map[string]string{
		"command":  "check_version",
		"category": "mariadb",
	},
}
