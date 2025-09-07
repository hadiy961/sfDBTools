package mariadb_cmd

import (
	"fmt"
	"os"

	"sfDBTools/internal/core/mariadb"
	"sfDBTools/internal/logger"
	mariadbUtils "sfDBTools/utils/mariadb"

	"github.com/spf13/cobra"
)

// CheckVersionCmd command for checking available MariaDB versions
var CheckVersionCmd = &cobra.Command{
	Use:   "check_version",
	Short: "Check available MariaDB versions",
	Long: `Check available MariaDB versions from official MariaDB download page.
This command fetches the list of supported MariaDB versions that can be installed
from the official MariaDB website.

Examples:
  # Standard display
  sfdbtools mariadb check_version

  # JSON output
  sfdbtools mariadb check_version --output json`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeVersionCheck(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("Version check failed", logger.Error(err))
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
	Annotations: map[string]string{
		"command":  "check_version",
		"category": "mariadb",
	},
}

func executeVersionCheck(cmd *cobra.Command) error {
	// 1. Resolve configuration first
	config, err := mariadbUtils.ResolveVersionConfig(cmd)
	if err != nil {
		return err
	}

	// 2. Get logger
	lg, err := logger.Get()
	if err != nil {
		return err
	}

	// 3. Log operation start
	lg.Info("Starting MariaDB version check operation")

	// 4. Get available versions using simple implementation
	versions, err := mariadb.GetAvailableVersions()
	if err != nil {
		return err
	}

	// 5. Display results
	return mariadb.DisplayVersions(versions, config.OutputFormat)
}

func init() {
	mariadbUtils.AddCommonVersionFlags(CheckVersionCmd)
}

func init() {
	CheckVersionCmd.Flags().String("output", "", "Output format (json for JSON output, empty for default)")
}
