package mariadb_cmd

import (
	"fmt"
	"os"

	"sfDBTools/internal/core/mariadb/configure"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

var (
	configureAutoConfirm   bool
	configureSkipUserSetup bool
	configureSkipDBSetup   bool
)

// ConfigureMariadbCMD represents the configure command
var ConfigureMariadbCMD = &cobra.Command{
	Use:   "configure",
	Short: "Configure MariaDB after installation",
	Long: `Configure MariaDB server after installation with custom settings.
This command will:
- Stop MariaDB service temporarily
- Configure custom data, binlog, and log directories  
- Setup configuration file from template
- Configure systemd service
- Setup firewall rules
- Migrate existing data
- Configure SELinux contexts (CentOS/RHEL)
- Create default databases and users
- Start and enable MariaDB service

The configuration will be applied according to settings in /etc/sfDBTools/config/config.yaml

Examples:
  # Interactive configuration  
  sfdbtools mariadb configure

  # Auto-confirm configuration with defaults
  sfdbtools mariadb configure --auto-confirm

  # Skip user and database setup
  sfdbtools mariadb configure --skip-user-setup --skip-db-setup`,
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.Get()
		if err != nil {
			terminal.PrintError("Failed to initialize logger")
			os.Exit(1)
		}

		lg.Info("MariaDB configure command started")

		// Create configuration
		config := &configure.ConfigureConfig{
			AutoConfirm:   configureAutoConfirm,
			SkipUserSetup: configureSkipUserSetup,
			SkipDBSetup:   configureSkipDBSetup,
		}

		// Create and run configure runner
		runner := configure.NewConfigureRunner(config)
		if err := runner.Run(); err != nil {
			lg.Error("MariaDB configuration failed", logger.Error(err))
			terminal.PrintError(fmt.Sprintf("Configuration failed: %v", err))
			os.Exit(1)
		}

		lg.Info("MariaDB configuration completed successfully")
	},
	Annotations: map[string]string{
		"command":  "configure",
		"category": "mariadb",
	},
}

func init() {
	// Configuration options
	ConfigureMariadbCMD.Flags().BoolVarP(&configureAutoConfirm, "auto-confirm", "y", false,
		"Automatically confirm all prompts and use default values")

	ConfigureMariadbCMD.Flags().BoolVar(&configureSkipUserSetup, "skip-user-setup", false,
		"Skip default user creation")

	ConfigureMariadbCMD.Flags().BoolVar(&configureSkipDBSetup, "skip-db-setup", false,
		"Skip default database creation")
}
