package mariadb_cmd

import (
	"os"

	"sfDBTools/internal/core/mariadb/upgrade"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

// Upgrade configuration variables
var (
	upgradeTargetVersion   string
	upgradeAutoConfirm     bool
	upgradeBackupData      bool
	upgradeBackupPath      string
	upgradeSkipBackup      bool
	upgradeForceUpgrade    bool
	upgradeSkipPostUpgrade bool
	upgradeTestMode        bool
)

// UpgradeMariaDBCMD represents the upgrade command
var UpgradeMariaDBCMD = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade MariaDB version for existing installation",
	Long: `Upgrade MariaDB to a newer version.

This command will:
1. Validate current installation and target version
2. Create a backup of existing data (unless --skip-backup)
3. Stop MariaDB service temporarily
4. Update MariaDB repository
5. Upgrade MariaDB packages
6. Start MariaDB service
7. Run mysql_upgrade (unless --skip-post-upgrade)
8. Verify upgrade success

Examples:
  # Upgrade to latest available LTS version
  sfdbtools mariadb upgrade

  # Upgrade to specific version
  sfdbtools mariadb upgrade --target-version=11.4

  # Upgrade with auto-confirmation (no prompts)
  sfdbtools mariadb upgrade --auto-confirm

  # Upgrade without backup (dangerous)
  sfdbtools mariadb upgrade --skip-backup

  # Test mode (dry-run)
  sfdbtools mariadb upgrade --test-mode

Safety Features:
- Data backup created by default before upgrade
- Validation checks before proceeding
- Rollback information provided if upgrade fails
- Service verification after upgrade`,
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.Get()
		if err != nil {
			terminal.PrintError("Failed to initialize logger")
			os.Exit(1)
		}

		lg.Info("MariaDB upgrade command started")

		// Create upgrade configuration
		config := &upgrade.UpgradeConfig{
			TargetVersion:   upgradeTargetVersion,
			AutoConfirm:     upgradeAutoConfirm,
			BackupData:      upgradeBackupData,
			BackupPath:      upgradeBackupPath,
			SkipBackup:      upgradeSkipBackup,
			ForceUpgrade:    upgradeForceUpgrade,
			SkipPostUpgrade: upgradeSkipPostUpgrade,
			TestMode:        upgradeTestMode,
		}

		// Create and run upgrade runner
		runner := upgrade.NewUpgradeRunner(config)
		if err := runner.Run(); err != nil {
			terminal.PrintError("MariaDB upgrade failed: " + err.Error())
			lg.Error("MariaDB upgrade failed", logger.Error(err))
			os.Exit(1)
		}

		lg.Info("MariaDB upgrade completed successfully")
	},
	Annotations: map[string]string{
		"command":  "upgrade",
		"category": "mariadb",
	},
}

func init() {
	// Target version
	UpgradeMariaDBCMD.Flags().StringVar(&upgradeTargetVersion, "target-version", "",
		"Target MariaDB version to upgrade to (default: latest LTS)")

	// Confirmation options
	UpgradeMariaDBCMD.Flags().BoolVarP(&upgradeAutoConfirm, "auto-confirm", "y", false,
		"Automatically confirm all prompts and proceed without user interaction")

	// Backup options
	UpgradeMariaDBCMD.Flags().BoolVar(&upgradeBackupData, "backup-data", true,
		"Create backup of data before upgrade")

	UpgradeMariaDBCMD.Flags().StringVar(&upgradeBackupPath, "backup-path", "",
		"Custom path for data backup (default: ~/mariadb_backups/upgrade_backup_TIMESTAMP)")

	UpgradeMariaDBCMD.Flags().BoolVar(&upgradeSkipBackup, "skip-backup", false,
		"Skip data backup (WARNING: This is dangerous)")

	// Upgrade options
	UpgradeMariaDBCMD.Flags().BoolVar(&upgradeForceUpgrade, "force-upgrade", false,
		"Force upgrade even if target version is lower (downgrade)")

	UpgradeMariaDBCMD.Flags().BoolVar(&upgradeSkipPostUpgrade, "skip-post-upgrade", false,
		"Skip running mysql_upgrade after package upgrade")

	// Testing options
	UpgradeMariaDBCMD.Flags().BoolVar(&upgradeTestMode, "test-mode", false,
		"Run in test mode (dry-run) - validate and plan but don't execute")
}
