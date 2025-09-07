package mariadb_cmd

import (
	"fmt"
	"os"

	"sfDBTools/internal/core/mariadb"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

// CheckVersionCmd command for checking available MariaDB versions
var CheckVersionCmd = &cobra.Command{
	Use:   "check_version",
	Short: "Check available MariaDB versions",
	Long: `Check available MariaDB versions from official MariaDB REST API.
Shows supported MariaDB versions that can be installed.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeVersionCheck(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("Version check failed", logger.Error(err))
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func executeVersionCheck(cmd *cobra.Command) error {
	lg, _ := logger.Get()
	lg.Info("Starting MariaDB version check")

	// Get output format from flag (if provided)
	outputFormat, _ := cmd.Flags().GetString("output")

	// Get available versions
	versions, err := mariadb.GetAvailableVersions()
	if err != nil {
		return err
	}

	// Display results
	return mariadb.DisplayVersions(versions, outputFormat)
}

func init() {
	CheckVersionCmd.Flags().String("output", "", "Output format (json for JSON output)")
}
