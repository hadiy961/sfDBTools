package mariadb_cmd

import (
	"fmt"
	"os"

	"sfDBTools/internal/core/mariadb/install"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

var (
	installVersion        string
	installAutoConfirm    bool
	installRemoveExisting bool
	installEnableSecurity bool
	installStartService   bool
)

// InstallCmd represents the install command
var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install MariaDB server",
	Long: `Install MariaDB server with specified configuration.
Supports automated installation with security setup and configuration tuning.

Available operating systems:
- CentOS 7, 8, 9
- Ubuntu 18.04, 20.04, 22.04, 24.04
- RHEL 7, 8, 9  
- Rocky Linux 8, 9
- AlmaLinux 8, 9

The installation process includes:
1. Operating system compatibility check
2. Internet connectivity verification
3. Fetching available MariaDB versions
4. Version selection (interactive or automatic)
5. Repository configuration
6. Package installation
7. Post-installation setup

Examples:
  # Interactive installation
  sfdbtools mariadb install

  # Auto-confirm with specific version
  sfdbtools mariadb install --version 10.11 --auto-confirm

  # Install and remove existing installation
  sfdbtools mariadb install --remove-existing --auto-confirm`,
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.Get()
		if err != nil {
			terminal.PrintError("Failed to initialize logger")
			os.Exit(1)
		}

		lg.Info("MariaDB install command started")

		// Create installation configuration
		config := &install.InstallConfig{
			Version:        installVersion,
			AutoConfirm:    installAutoConfirm,
			RemoveExisting: installRemoveExisting,
			EnableSecurity: installEnableSecurity,
			StartService:   installStartService,
		}

		// Create and run installer
		runner := install.NewInstallRunner(config)
		if err := runner.Run(); err != nil {
			lg.Error("MariaDB installation failed", logger.Error(err))
			terminal.PrintError(fmt.Sprintf("Installation failed: %v", err))
			os.Exit(1)
		}

		lg.Info("MariaDB installation completed successfully")
	},
	Annotations: map[string]string{
		"command":  "install",
		"category": "mariadb",
	},
}

func init() {
	// Version selection
	InstallCmd.Flags().StringVarP(&installVersion, "version", "v", "",
		"MariaDB major version to install (e.g., 10.11, 10.6)")

	// Installation options
	InstallCmd.Flags().BoolVarP(&installAutoConfirm, "auto-confirm", "y", false,
		"Automatically confirm all prompts")

	InstallCmd.Flags().BoolVar(&installRemoveExisting, "remove-existing", false,
		"Remove existing MariaDB installation if found")

	// Service options
	InstallCmd.Flags().BoolVar(&installEnableSecurity, "enable-security", true,
		"Enable security setup (mysql_secure_installation will need to be run manually)")

	InstallCmd.Flags().BoolVar(&installStartService, "start-service", true,
		"Start MariaDB service after installation")
}
