package mariadb_cmd

import (
	"os"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

// RemoveCmd represents the remove command
var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove MariaDB server",
	Long: `Remove MariaDB server including data cleanup and service removal.
Provides options for data backup before removal.`,
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.Get()
		if err != nil {
			terminal.PrintError("Failed to initialize logger")
			os.Exit(1)
		}

		lg.Info("MariaDB remove command called")

		// TODO: Implement removal logic in internal/core/mariadb/remove/
		terminal.PrintInfo("MariaDB remove feature - Coming Soon!")
		terminal.PrintWarning("This feature will be implemented in internal/core/mariadb/remove/")
	},
	Annotations: map[string]string{
		"command":  "remove",
		"category": "mariadb",
	},
}
