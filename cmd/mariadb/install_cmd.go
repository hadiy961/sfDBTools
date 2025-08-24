package command_mariadb

import (
	"fmt"
	"os"

	installcmd "sfDBTools/cmd/mariadb/install_cmd"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

// InstallCmd is the entrypoint cobra command for mariadb install. The heavy
// lifting has been moved to files under cmd/mariadb/install_cmd/ for better
// modularity and testability.
var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install MariaDB with custom configuration",
	Long:  "Install MariaDB with custom configuration (see sub-files for implementation)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := installcmd.Execute(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("MariaDB install failed", logger.Error(err))
			fmt.Fprintln(os.Stderr, "MariaDB install failed:", err)
			os.Exit(1)
		}
	},
	Annotations: map[string]string{
		"command":  "install",
		"category": "mariadb",
	},
}

func init() {
	installcmd.InitFlags(InstallCmd)
}
