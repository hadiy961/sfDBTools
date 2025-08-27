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
	upgradeRemoveExisting  bool
	upgradeStartService    bool
	upgradeEnableSecurity  bool
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
  # Interactive upgrade to latest available LTS version
  sfdbtools mariadb upgrade

  # Auto-confirm upgrade to specific version
  sfdbtools mariadb upgrade --target-version=10.11 --auto-confirm

  # Upgrade and remove existing installation
  sfdbtools mariadb upgrade --remove-existing --auto-confirm

  # Upgrade without backup (dangerous)
  sfdbtools mariadb upgrade --skip-backup --auto-confirm

  # Test mode (dry-run) to see what would be upgraded
  sfdbtools mariadb upgrade --test-mode
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
			RemoveExisting:  upgradeRemoveExisting,
			StartService:    upgradeStartService,
			EnableSecurity:  upgradeEnableSecurity,
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
	// Version selection (consistent with install command)
	UpgradeMariaDBCMD.Flags().StringVarP(&upgradeTargetVersion, "target-version", "v", "",
		"MariaDB version to upgrade to (e.g., 10.11, 11.4, default: latest LTS)")

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

	// Installation options (similar to install command)
	UpgradeMariaDBCMD.Flags().BoolVar(&upgradeRemoveExisting, "remove-existing", false,
		"Remove existing MariaDB installation before upgrade")

	// Service options
	UpgradeMariaDBCMD.Flags().BoolVar(&upgradeStartService, "start-service", true,
		"Start MariaDB service after upgrade (default: true)")

	UpgradeMariaDBCMD.Flags().BoolVar(&upgradeEnableSecurity, "enable-security", true,
		"Enable security setup after upgrade (mysql_secure_installation will need to be run manually)")
}
