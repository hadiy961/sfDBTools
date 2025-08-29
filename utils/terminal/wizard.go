package terminal

import (
	"fmt"
	"strings"
)

// RemovalWizard provides a conversational interface for MariaDB removal
type RemovalWizard struct{}

// NewRemovalWizard creates a new removal wizard
func NewRemovalWizard() *RemovalWizard {
	return &RemovalWizard{}
}

// GatherRemovalConfig walks user through removal configuration
func (w *RemovalWizard) GatherRemovalConfig() (*RemovalConfig, error) {
	PrintHeader("🗑️  MariaDB Removal Configuration Wizard")
	PrintInfo("This wizard will help you safely configure MariaDB removal options.")
	PrintInfo("")

	config := &RemovalConfig{}

	// Step 1: Data removal decision
	w.showDataRemovalInfo()
	config.RemoveData = AskWithContext(
		"Remove data directories?",
		"⚠️  This will permanently delete all databases and cannot be undone",
		false,
	)

	// Step 2: Backup decision
	w.showBackupInfo(config.RemoveData)
	if config.RemoveData {
		config.BackupData = AskWithContext(
			"Create backup of data before removal?",
			"🔒 Highly recommended when removing data - provides safety net for recovery",
			true,
		)
	} else {
		config.BackupData = AskWithContext(
			"Create backup of data before removal?",
			"📦 Backup current data for safety, even when keeping data directories",
			true,
		)
	}

	// Step 3: Backup path (if backup enabled)
	if config.BackupData {
		config.BackupPath = w.getBackupPath()
	}

	// Step 4: Repository removal
	w.showRepositoryInfo()
	config.RemoveRepositories = AskYesNo("Remove MariaDB repositories?", false)

	// Step 5: Show summary and final confirmation
	w.showRemovalSummary(config)

	confirmed := AskWithContext(
		"Proceed with removal using these settings?",
		"⚡ This will start the removal process - make sure you're ready!",
		false,
	)

	if !confirmed {
		return nil, fmt.Errorf("removal cancelled by user")
	}

	return config, nil
}

// showDataRemovalInfo explains data removal implications
func (w *RemovalWizard) showDataRemovalInfo() {
	PrintSubHeader("📊 Data Directory Configuration")
	PrintInfo("MariaDB stores all your databases in data directories.")
	PrintInfo("Removing data directories will:")
	PrintInfo("  ✗ Permanently delete all databases")
	PrintInfo("  ✗ Remove all user data")
	PrintInfo("  ✗ Delete transaction logs")
	PrintInfo("")
	PrintInfo("Keeping data directories allows you to:")
	PrintInfo("  ✓ Reinstall MariaDB later with existing data")
	PrintInfo("  ✓ Migrate data to another server")
	PrintInfo("  ✓ Keep databases for backup purposes")
	PrintInfo("")
}

// showBackupInfo explains backup options
func (w *RemovalWizard) showBackupInfo(removingData bool) {
	PrintSubHeader("💾 Backup Strategy")
	if removingData {
		PrintInfo("Since you're removing data, backup is your safety net:")
		PrintInfo("  ✓ Allows complete recovery if needed")
		PrintInfo("  ✓ Enables data migration to new installation")
		PrintInfo("  ✓ Archives data for compliance/audit requirements")
	} else {
		PrintInfo("Even when keeping data, backup provides extra protection:")
		PrintInfo("  ✓ Guards against accidental removal")
		PrintInfo("  ✓ Creates clean snapshot before changes")
		PrintInfo("  ✓ Enables easy rollback if issues occur")
	}
	PrintInfo("")
}

// getBackupPath gets backup path from user
func (w *RemovalWizard) getBackupPath() string {
	PrintSubHeader("📁 Backup Location")
	PrintInfo("Choose where to store your MariaDB backup:")
	PrintInfo("  Default: ~/mariadb_backups/mariadb_backup_TIMESTAMP")
	PrintInfo("  Custom:  Specify your own path")
	PrintInfo("")

	backupPath := AskString("Enter custom backup path (leave empty for default)", "")

	if backupPath != "" {
		PrintSuccess(fmt.Sprintf("✓ Using custom backup path: %s", backupPath))
	} else {
		PrintInfo("✓ Using default backup location")
	}
	PrintInfo("")

	return backupPath
}

// showRepositoryInfo explains repository removal
func (w *RemovalWizard) showRepositoryInfo() {
	PrintSubHeader("📦 Repository Management")
	PrintInfo("MariaDB repositories provide package updates and dependencies.")
	PrintInfo("")
	PrintInfo("Remove repositories if:")
	PrintInfo("  ✓ You won't reinstall MariaDB on this system")
	PrintInfo("  ✓ You want complete cleanup")
	PrintInfo("  ✓ You'll use different package sources later")
	PrintInfo("")
	PrintInfo("Keep repositories if:")
	PrintInfo("  ✓ You might reinstall MariaDB later")
	PrintInfo("  ✓ Other applications depend on them")
	PrintInfo("  ✓ You prefer minimal changes")
	PrintInfo("")
}

// showRemovalSummary displays the complete removal plan
func (w *RemovalWizard) showRemovalSummary(config *RemovalConfig) {
	PrintHeader("📋 Removal Plan Summary")

	w.showImpactBox(map[string]string{
		"📦 Packages":     w.getPackagesSummary(),
		"🔧 Services":     "Stop and disable MariaDB services",
		"📁 Data":         w.getDataSummary(config),
		"⚙️  Config":     "Remove configuration files",
		"📚 Repositories": w.getRepositoriesSummary(config),
		"💾 Backup":       w.getBackupSummary(config),
	})

	PrintInfo("")
	PrintInfo("⏱️  Estimated time: 2-5 minutes")
	PrintInfo("🔄 Rollback: Available for most operations")
	PrintInfo("")
}

// showImpactBox displays removal impact in a formatted box
func (w *RemovalWizard) showImpactBox(impacts map[string]string) {
	maxKeyLen := 0
	for key := range impacts {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}

	boxWidth := maxKeyLen + 50

	// Top border
	fmt.Println("┌" + strings.Repeat("─", boxWidth-2) + "┐")

	// Content
	for key, value := range impacts {
		padding := strings.Repeat(" ", maxKeyLen-len(key))
		fmt.Printf("│ %s%s │ %s\n", key, padding, value)
	}

	// Bottom border
	fmt.Println("└" + strings.Repeat("─", boxWidth-2) + "┘")
}

// Helper functions for summary generation
func (w *RemovalWizard) getPackagesSummary() string {
	return "Remove all MariaDB packages"
}

func (w *RemovalWizard) getDataSummary(config *RemovalConfig) string {
	if config.RemoveData {
		return "⚠️  REMOVE all databases permanently"
	}
	return "✓ Keep all databases intact"
}

func (w *RemovalWizard) getRepositoriesSummary(config *RemovalConfig) string {
	if config.RemoveRepositories {
		return "Remove MariaDB package repositories"
	}
	return "Keep MariaDB package repositories"
}

func (w *RemovalWizard) getBackupSummary(config *RemovalConfig) string {
	if config.BackupData {
		if config.BackupPath != "" {
			return fmt.Sprintf("Create backup at: %s", config.BackupPath)
		}
		return "Create backup at default location"
	}
	return "No backup will be created"
}

// RemovalConfig represents removal configuration from wizard
type RemovalConfig struct {
	RemoveData         bool
	BackupData         bool
	BackupPath         string
	RemoveRepositories bool
}
