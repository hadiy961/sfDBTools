package mariadb_cmd

import (
	"fmt"
	"os"

	"sfDBTools/internal/core/mariadb"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

// RemoveCmd removes MariaDB installation
var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove MariaDB installation (dangerous)",
	Long: `Remove MariaDB server, packages, data directories and configuration.
This action is destructive. Use --yes to skip confirmations.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeRemove(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("MariaDB remove failed", logger.Error(err))
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	RemoveCmd.Flags().Bool("yes", false, "Skip confirmations and run non-interactively (dangerous)")
}

func executeRemove(cmd *cobra.Command) error {
	skipConfirm, _ := cmd.Flags().GetBool("yes")
	return mariadb.RemoveMariaDB(skipConfirm)
}
