package mariadb_cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// CheckVersionCmd is a placeholder command that prints a coming soon message
var CheckVersionCmd = &cobra.Command{
	Use:   "check_version",
	Short: "Coming soon",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Coming soon: check_version")
	},
	Annotations: map[string]string{
		"command":  "check_version",
		"category": "mariadb",
	},
}
