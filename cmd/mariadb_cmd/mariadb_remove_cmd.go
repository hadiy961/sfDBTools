package mariadb_cmd

import (
	"fmt"
	"os"

	"sfDBTools/internal/core/mariadb/remove"
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
			os.Exit(1)
		}
	},
	Annotations: map[string]string{
		"command":  "remove",
		"category": "mariadb",
	},
}

func init() {
	RemoveCmd.Flags().Bool("yes", false, "Skip confirmations and run non-interactively (dangerous)")
}

func executeRemove(cmd *cobra.Command) error {
	skipConfirm, _ := cmd.Flags().GetBool("yes")

	cfg := &remove.Config{SkipConfirm: skipConfirm}

	r, err := remove.NewRemover(cfg)
	if err != nil {
		return err
	}

	res, err := r.Remove()
	if err != nil {
		return err
	}

	if !res.Success {
		// Don't treat "no services found" as an error
		if res.Message == "no MariaDB services found" {
			return nil
		}
		return fmt.Errorf("remove failed: %s", res.Message)
	}

	return nil
}
