package command_mariadb

import (
	"fmt"
	"os"
	"sfDBTools/internal/core/mariadb"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

var ConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure MariaDB with custom settings",
	Long: `Configure command provides comprehensive MariaDB configuration management.

This command allows you to:
- Customize MariaDB configuration using server.cnf template
- Set server ID, ports, and directory paths
- Configure InnoDB settings (data directory, log directory)
- Enable/disable encryption with custom key files
- Set up binary logging
- Apply security best practices

The configuration process will:
1. Backup existing configuration files
2. Prompt for custom settings interactively
3. Generate new configuration based on server.cnf template
4. Copy encryption keys if encryption is enabled
5. Restart MariaDB service to apply changes

Examples:
  # Interactive configuration
  sfDBTools mariadb configure

  # The command will guide you through setting up:
  # - Server ID and port
  # - Data and log directories
  # - Encryption settings
  # - Binary log configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeConfigure(); err != nil {
			lg, _ := logger.Get()
			lg.Error("MariaDB configuration failed", logger.Error(err))
			fmt.Printf("❌ Configuration failed: %v\n", err)
			os.Exit(1)
		}
	},
	Annotations: map[string]string{
		"command":  "configure",
		"category": "mariadb",
	},
}

func executeConfigure() error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting MariaDB configuration process")

	// Execute configuration
	result, err := mariadb.ConfigureMariaDB()
	if err != nil {
		return fmt.Errorf("configuration failed: %w", err)
	}

	// Display results
	mariadb.DisplayConfigResult(result)

	lg.Info("MariaDB configuration completed",
		logger.Bool("success", result.Success),
		logger.String("config_path", result.ConfigPath))

	return nil
}

func init() {
	// No flags needed for now since it's interactive
}
