package mariadb_cmd

import (
	"github.com/spf13/cobra"
)

// RemoveCmd removes MariaDB installation
var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove MariaDB installation (dangerous)",
	Long: `Remove MariaDB server, packages, data directories and configuration.
This action is destructive. Use --yes to skip confirmations.`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}
