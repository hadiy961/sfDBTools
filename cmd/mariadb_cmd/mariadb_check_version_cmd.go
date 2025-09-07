package mariadb_cmd

import (
	"github.com/spf13/cobra"
)

// CheckVersionCmd command for checking available MariaDB versions
var CheckVersionCmd = &cobra.Command{
	Use:   "check_version",
	Short: "Check available MariaDB versions",
	Long: `Check available MariaDB versions from official MariaDB REPO SCRIPT.
Shows supported MariaDB versions that can be installed.`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	// No flags needed since we only have one display format
}
