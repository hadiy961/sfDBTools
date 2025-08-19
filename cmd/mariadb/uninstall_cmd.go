package command_mariadb

import (
	"fmt"
	"os"

	"sfDBTools/internal/core/mariadb"
	"sfDBTools/internal/logger"
	mariadb_utils "sfDBTools/utils/mariadb"

	"github.com/spf13/cobra"
)

var UninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Completely uninstall MariaDB/MySQL from the system",
	Long: `Uninstall command provides a comprehensive solution for completely removing MariaDB/MySQL from your system.

This command supports:
- Cross-platform support: CentOS, RHEL, AlmaLinux, Rocky Linux, Ubuntu, Debian
- Complete package removal: All MariaDB/MySQL server, client, and common packages
- Data cleanup: Removes data directories, configuration files, and logs
- Service management: Stops and disables MariaDB services
- Verification: Confirms complete removal and reports any remaining components
- Safety prompts: Interactive confirmation with detailed warnings about data loss

⚠️  WARNING: This will permanently delete all databases, users, and configuration!
Always backup your data before uninstalling.

Examples:
  # Interactive uninstall with confirmation prompts
  sfDBTools mariadb uninstall

  # Force uninstall without confirmation (use with caution)
  sfDBTools mariadb uninstall --force

  # Keep data directories (remove only packages and configs)
  sfDBTools mariadb uninstall --keep-data

  # Keep configuration files (remove only packages and data)
  sfDBTools mariadb uninstall --keep-config`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeUninstall(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("Command failed", logger.Error(err))
			os.Exit(1)
		}
	},
	Annotations: map[string]string{
		"command":  "uninstall",
		"category": "mariadb",
	},
}

func executeUninstall(cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	// Get flags
	force, _ := cmd.Flags().GetBool("force")
	keepData, _ := cmd.Flags().GetBool("keep-data")
	keepConfig, _ := cmd.Flags().GetBool("keep-config")
	backupFirst, _ := cmd.Flags().GetBool("backup-first")
	backupDir, _ := cmd.Flags().GetString("backup-dir")

	lg.Info("Starting MariaDB uninstall process",
		logger.Bool("force", force),
		logger.Bool("keep_data", keepData),
		logger.Bool("keep_config", keepConfig),
		logger.Bool("backup_first", backupFirst))

	// Show warning and get confirmation (unless force mode)
	if !force {
		mariadb_utils.ShowUninstallWarning()

		if !mariadb_utils.PromptConfirmation() {
			lg.Info("Uninstall cancelled by user")
			fmt.Println("❌ Uninstall cancelled.")
			return nil
		}
	}

	// Show summary
	options := mariadb_utils.UninstallOptions{
		Force:       force,
		KeepData:    keepData,
		KeepConfig:  keepConfig,
		BackupFirst: backupFirst,
		BackupDir:   backupDir,
	}

	if !force {
		mariadb_utils.ShowUninstallSummary(options)
	}

	// Execute uninstall
	result, err := mariadb.UninstallMariaDB(options)
	if err != nil {
		lg.Error("Uninstall failed", logger.Error(err))
		if result != nil {
			mariadb_utils.DisplayUninstallResults(result)
		}
		return fmt.Errorf("uninstall failed: %w", err)
	}

	// Display results
	mariadb_utils.DisplayUninstallResults(result)

	// Log results
	if result.Success {
		lg.Info("MariaDB uninstall completed successfully",
			logger.String("duration", result.Duration.String()),
			logger.Int("packages_removed", result.PackagesRemoved),
			logger.Int("directories_removed", len(result.DirectoriesRemoved)))
	} else {
		lg.Warn("MariaDB uninstall completed with issues",
			logger.String("duration", result.Duration.String()),
			logger.Int("warnings", len(result.Warnings)),
			logger.Int("errors", len(result.Errors)))
	}

	return nil
}

func init() {
	UninstallCmd.Flags().Bool("force", false, "Skip confirmation prompts (use with caution)")
	UninstallCmd.Flags().Bool("keep-data", false, "Keep data directories (only remove packages)")
	UninstallCmd.Flags().Bool("keep-config", false, "Keep configuration files")
	UninstallCmd.Flags().Bool("backup-first", false, "Create backup before uninstalling")
	UninstallCmd.Flags().String("backup-dir", "./mariadb_backup", "Directory for backup files")
}
