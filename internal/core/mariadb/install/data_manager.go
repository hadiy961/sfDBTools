package install

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// DataManager handles data directory operations
type DataManager struct {
	selectedVersion *SelectableVersion
}

// NewDataManager creates a new data manager
func NewDataManager(selectedVersion *SelectableVersion) *DataManager {
	return &DataManager{
		selectedVersion: selectedVersion,
	}
}

// CheckDataDirectoryCompatibility checks if existing data directory is compatible
func (d *DataManager) CheckDataDirectoryCompatibility() error {
	lg, _ := logger.Get()

	// Check if data directory exists
	dataDir := "/var/lib/mysql"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		return nil // No data directory, no conflict
	}

	// Check for version info file
	versionFile := dataDir + "/mariadb_upgrade_info"
	if _, err := os.Stat(versionFile); os.IsNotExist(err) {
		return nil // No version info, assume compatibility
	}

	// Read existing version
	content, err := os.ReadFile(versionFile)
	if err != nil {
		lg.Warn("Failed to read version info", logger.Error(err))
		return nil // Can't read, assume compatibility
	}

	existingVersion := strings.TrimSpace(string(content))
	currentVersion := d.selectedVersion.LatestVersion

	lg.Info("Checking data directory compatibility",
		logger.String("existing_version", existingVersion),
		logger.String("installing_version", currentVersion))

	// Parse versions for comparison
	if d.isVersionDowngrade(existingVersion, currentVersion) {
		return fmt.Errorf("data directory contains newer version (%s), cannot downgrade to %s",
			existingVersion, currentVersion)
	}

	return nil
}

// HandleDataDirectoryConflict resolves data directory version conflicts
func (d *DataManager) HandleDataDirectoryConflict(autoConfirm bool) error {
	terminal.PrintWarning("Incompatible MariaDB data directory detected!")
	terminal.PrintInfo("The existing data directory contains a newer MariaDB version.")
	terminal.PrintInfo("Options:")
	terminal.PrintInfo("  1. Backup and reinitialize data directory (recommended)")
	terminal.PrintInfo("  2. Cancel installation")

	if autoConfirm {
		terminal.PrintInfo("Auto-confirm enabled: backing up and reinitializing data directory")
		return d.backupAndReinitializeDataDirectory()
	}

	fmt.Print("Backup and reinitialize data directory? (y/N): ")
	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	if response == "y" || response == "yes" {
		return d.backupAndReinitializeDataDirectory()
	}

	return fmt.Errorf("installation cancelled due to data directory conflict")
}

// backupAndReinitializeDataDirectory backs up existing data and creates fresh data directory
func (d *DataManager) backupAndReinitializeDataDirectory() error {
	lg, _ := logger.Get()

	dataDir := "/var/lib/mysql"
	backupDir := fmt.Sprintf("/var/lib/mysql.backup.%s",
		time.Now().Format("20060102_150405"))

	// Stop MariaDB service if running
	spinner := terminal.NewProgressSpinner("Stopping MariaDB service...")
	spinner.Start()

	serviceManager := NewServiceManager()
	_ = serviceManager.StopMariaDBService() // Ignore error if not running

	spinner.Stop()
	terminal.PrintSuccess("MariaDB service stopped")

	// Backup existing data directory
	spinner = terminal.NewProgressSpinner("Backing up existing data directory...")
	spinner.Start()

	cmd := exec.Command("mv", dataDir, backupDir)
	if err := cmd.Run(); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to backup data directory: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess(fmt.Sprintf("Data directory backed up to: %s", backupDir))

	// Reinitialize data directory
	spinner = terminal.NewProgressSpinner("Initializing new data directory...")
	spinner.Start()

	cmd = exec.Command("mysql_install_db", "--user=mysql", "--basedir=/usr", "--datadir="+dataDir)
	if err := cmd.Run(); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to initialize data directory: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess("New data directory initialized")

	lg.Info("Data directory conflict resolved",
		logger.String("backup_location", backupDir),
		logger.String("new_data_dir", dataDir))

	return nil
}

// isVersionDowngrade checks if target version is older than existing version
func (d *DataManager) isVersionDowngrade(existing, target string) bool {
	// Extract major.minor version numbers
	existingParts := strings.Split(strings.Split(existing, "-")[0], ".")
	targetParts := strings.Split(target, ".")

	// Convert to integers for comparison
	existingMajor, existingMinor := parseVersionParts(existingParts)
	targetMajor, targetMinor := parseVersionParts(targetParts)

	// Compare major version
	if targetMajor < existingMajor {
		return true
	}
	if targetMajor > existingMajor {
		return false
	}

	// Same major version, compare minor
	return targetMinor < existingMinor
}
