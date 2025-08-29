package mariadb_cmd

import (
	"context"
	"os"

	"sfDBTools/internal/core/mariadb/remove"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

// RemoveCmd represents the remove command
var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove MariaDB server",
	Long: `Remove MariaDB server including packages, services, and optionally data directories.

This command will:
1. Detect existing MariaDB installation
2. Optionally backup data before removal
3. Stop and disable MariaDB services
4. Remove MariaDB packages
5. Optionally remove data directories
6. Remove configuration and log files
7. Optionally remove MariaDB repositories
8. Clean up residual files

Examples:
  # Remove MariaDB but keep data
  sfdbtools mariadb remove

  # Remove MariaDB and all data (with backup)
  sfdbtools mariadb remove --remove-data --backup-data

  # Remove everything including repositories (auto-confirm)
  sfdbtools mariadb remove --remove-data --remove-repositories --auto-confirm

  # Remove with custom backup location
  sfdbtools mariadb remove --backup-data --backup-path /path/to/backup

Safety Features:
- Data is backed up by default before removal
- Confirmation is required unless --auto-confirm is used
- Data directories are preserved by default`,
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.Get()
		if err != nil {
			terminal.PrintError("Failed to initialize logger")
			os.Exit(1)
		}

		lg.Info("MariaDB remove command started")

		// Resolve removal configuration
		config, err := resolveRemovalConfig(cmd)
		if err != nil {
			terminal.PrintError("Configuration error: " + err.Error())
			lg.Error("Failed to resolve removal configuration", logger.Error(err))
			os.Exit(1)
		}

		// Create orchestrator with dependencies
		orchestrator := remove.NewOrchestrator(remove.Dependencies{
			PackageManager: system.NewPackageManager(),
			ServiceManager: system.NewServiceManager(),
			FileSystem:     system.NewSafeFileSystem(),
		})

		// Execute removal
		ctx := context.Background()
		if err := orchestrator.Execute(ctx, config); err != nil {
			terminal.PrintError("MariaDB removal failed: " + err.Error())
			lg.Error("MariaDB removal failed", logger.Error(err))
			os.Exit(1)
		}

		terminal.PrintSuccess("MariaDB removal completed successfully")
		lg.Info("MariaDB removal completed successfully")
	},
	Annotations: map[string]string{
		"command":  "remove",
		"category": "mariadb",
	},
}

func init() {
	// Add flags for removal configuration
	RemoveCmd.Flags().Bool("remove-data", false, "Remove data directories (WARNING: This will permanently delete all databases)")
	RemoveCmd.Flags().Bool("backup-data", true, "Create backup of data before removal")
	RemoveCmd.Flags().String("backup-path", "", "Custom path for data backup (default: ~/mariadb_backups)")
	RemoveCmd.Flags().Bool("remove-repositories", false, "Remove MariaDB repositories")
	RemoveCmd.Flags().BoolP("auto-confirm", "y", false, "Automatically confirm all prompts")
}

// resolveRemovalConfig resolves the removal configuration from flags and interactive input
func resolveRemovalConfig(cmd *cobra.Command) (*remove.RemovalConfig, error) {
	// Check if any flags were provided
	flagsProvided := cmd.Flags().Changed("remove-data") ||
		cmd.Flags().Changed("backup-data") ||
		cmd.Flags().Changed("backup-path") ||
		cmd.Flags().Changed("remove-repositories") ||
		cmd.Flags().Changed("auto-confirm")

	// If no flags provided, use interactive wizard
	if !flagsProvided {
		wizard := terminal.NewRemovalWizard()
		wizardConfig, err := wizard.GatherRemovalConfig()
		if err != nil {
			return nil, err
		}

		// Convert wizard config to removal config
		return &remove.RemovalConfig{
			RemoveData:         wizardConfig.RemoveData,
			BackupData:         wizardConfig.BackupData,
			BackupPath:         wizardConfig.BackupPath,
			RemoveRepositories: wizardConfig.RemoveRepositories,
			AutoConfirm:        false, // Wizard handles confirmations
		}, nil
	}

	// Use flag values
	removeData, _ := cmd.Flags().GetBool("remove-data")
	backupData, _ := cmd.Flags().GetBool("backup-data")
	backupPath, _ := cmd.Flags().GetString("backup-path")
	removeRepositories, _ := cmd.Flags().GetBool("remove-repositories")
	autoConfirm, _ := cmd.Flags().GetBool("auto-confirm")

	return &remove.RemovalConfig{
		RemoveData:         removeData,
		BackupData:         backupData,
		BackupPath:         backupPath,
		RemoveRepositories: removeRepositories,
		AutoConfirm:        autoConfirm,
	}, nil
}
