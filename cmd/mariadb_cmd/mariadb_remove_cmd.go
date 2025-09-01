package mariadb_cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RemoveCmd is a placeholder that prints coming soon
var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Coming soon",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Coming soon: remove")
	},
	Annotations: map[string]string{
		"command":  "remove",
		"category": "mariadb",
	},
}
