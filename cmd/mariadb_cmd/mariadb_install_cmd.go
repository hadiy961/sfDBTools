package mariadb_cmd

import (
	"fmt"
	"os"

	"sfDBTools/internal/core/mariadb/install"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

// InstallCmd installs MariaDB interactively
var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install MariaDB server interactively",
	Long: `Install MariaDB server with interactive version selection.

This command performs the following steps:
1. Check for existing MariaDB service and packages
2. Verify internet connectivity
3. Detect operating system
4. Check repository availability  
5. Fetch available MariaDB versions
6. Allow user to select version to install
7. Setup MariaDB repository
8. Install MariaDB server and client
9. Start and enable MariaDB service
10. Verify installation

The installation process is fully interactive and requires no flags.
It will install the default MariaDB configuration without customization.

Examples:
  # Interactive MariaDB installation
  sfdbtools mariadb install
  
  # Dry run to test the installation flow
  sfdbtools mariadb install --dry-run`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeInstall(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("MariaDB installation failed", logger.Error(err))
			os.Exit(1)
		}
	},
	Annotations: map[string]string{
		"command":  "install",
		"category": "mariadb",
	},
}

func init() {
	InstallCmd.Flags().Bool("dry-run", false, "Perform a dry run without actual installation")
}

func executeInstall(cmd *cobra.Command) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if dryRun {
		fmt.Println("🧪 Running in dry-run mode - no actual installation will be performed")
		dryRunInstaller, err := install.NewDryRunInstaller()
		if err != nil {
			return err
		}
		_, err = dryRunInstaller.DryRun()
		return err
	}

	installer, err := install.NewInstaller(nil)
	if err != nil {
		return err
	}

	result, err := installer.Install()
	if err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("installation failed: %s", result.Message)
	}

	return nil
}
