package mariadb_cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// InstallCmd is a placeholder that prints coming soon
var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Coming soon",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Coming soon: install")
	},
	Annotations: map[string]string{
		"command":  "install",
		"category": "mariadb",
	},
}
