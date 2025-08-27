package remove

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/terminal"
)

// RemovalRunner orchestrates the MariaDB removal process
type RemovalRunner struct {
	config           *RemovalConfig
	osInfo           *common.OSInfo
	detectionService *DetectionService
	backupService    *BackupService
	removalService   *RemovalService
	installation     *DetectedInstallation
}

// NewRemovalRunner creates a new removal runner
func NewRemovalRunner(config *RemovalConfig) *RemovalRunner {
	if config == nil {
		config = DefaultRemovalConfig()
	}

	return &RemovalRunner{
		config: config,
	}
}

// Run executes the complete MariaDB removal process
func (r *RemovalRunner) Run() error {
	lg, _ := logger.Get()

	lg.Info("Starting MariaDB removal process")
	terminal.PrintHeader("MariaDB Removal")

	// Step 1: Check OS compatibility
	if err := r.checkOSCompatibility(); err != nil {
		return fmt.Errorf("OS compatibility check failed: %w", err)
	}

	// Step 2: Detect existing installation
	if err := r.detectInstallation(); err != nil {
		return fmt.Errorf("installation detection failed: %w", err)
	}

	// Step 3: Show installation summary and confirm removal
	if err := r.confirmRemoval(); err != nil {
		return fmt.Errorf("removal confirmation failed: %w", err)
	}

	// Step 4: Backup data if requested
	if err := r.backupData(); err != nil {
		return fmt.Errorf("data backup failed: %w", err)
	}

	// Step 5: Stop and disable services
	if err := r.stopServices(); err != nil {
		return fmt.Errorf("service stopping failed: %w", err)
	}

	// Step 6: Remove packages
	if err := r.removePackages(); err != nil {
		return fmt.Errorf("package removal failed: %w", err)
	}

	// Step 7: Remove data directories
	if err := r.removeDataDirectories(); err != nil {
		return fmt.Errorf("data directory removal failed: %w", err)
	}

	// Step 8: Remove configuration files
	if err := r.removeConfigFiles(); err != nil {
		return fmt.Errorf("configuration file removal failed: %w", err)
	}

	// Step 9: Remove log files
	if err := r.removeLogFiles(); err != nil {
		return fmt.Errorf("log file removal failed: %w", err)
	}

	// Step 10: Remove repositories if requested
	if err := r.removeRepositories(); err != nil {
		return fmt.Errorf("repository removal failed: %w", err)
	}

	// Step 11: Cleanup residual files
	if err := r.cleanupResidualFiles(); err != nil {
		return fmt.Errorf("residual cleanup failed: %w", err)
	}

	terminal.PrintSuccess("MariaDB removal completed successfully!")
	lg.Info("MariaDB removal completed successfully")

	// Step 12: Show removal summary
	r.showRemovalSummary()

	return nil
}

// checkOSCompatibility checks if the OS is supported
func (r *RemovalRunner) checkOSCompatibility() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Detecting operating system...")
	spinner.Start()

	// Detect OS using common utility
	detector := common.NewOSDetector()
	osInfo, err := detector.DetectOS()
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to detect operating system: %w", err)
	}

	r.osInfo = osInfo

	// Validate OS compatibility
	if err := common.ValidateOperatingSystem(common.MariaDBSupportedOS()); err != nil {
		spinner.Stop()
		return fmt.Errorf("operating system not supported: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess(fmt.Sprintf("Operating system detected: %s %s (%s)",
		osInfo.Name, osInfo.Version, osInfo.Architecture))

	lg.Info("OS compatibility check passed",
		logger.String("os", osInfo.ID),
		logger.String("version", osInfo.Version),
		logger.String("arch", osInfo.Architecture))

	// Initialize services
	r.detectionService = NewDetectionService(r.osInfo)
	r.backupService = NewBackupService(r.osInfo)
	r.removalService = NewRemovalService(r.osInfo)

	return nil
}

// detectInstallation detects existing MariaDB installation
func (r *RemovalRunner) detectInstallation() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Detecting MariaDB installation...")
	spinner.Start()

	installation, err := r.detectionService.DetectInstallation()
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to detect installation: %w", err)
	}

	r.installation = installation
	spinner.Stop()

	if !installation.IsInstalled {
		terminal.PrintInfo("No MariaDB installation detected")
		lg.Info("No MariaDB installation found")
		return fmt.Errorf("no MariaDB installation found to remove")
	}

	terminal.PrintSuccess("MariaDB installation detected")
	lg.Info("MariaDB installation detected",
		logger.String("package", installation.PackageName),
		logger.String("version", installation.Version))

	return nil
}

// confirmRemoval shows installation details and confirms removal
func (r *RemovalRunner) confirmRemoval() error {
	lg, _ := logger.Get()

	// Show installation summary
	summary := r.detectionService.GetInstallationSummary(r.installation)
	terminal.PrintInfo("Current Installation:")
	fmt.Print(summary)

	// Show what will be removed
	terminal.PrintInfo("\nThe following will be removed:")

	if r.installation.IsInstalled {
		terminal.PrintInfo(fmt.Sprintf("  ✓ MariaDB packages (%s)", r.installation.PackageName))
	}

	if r.installation.ServiceName != "" {
		terminal.PrintInfo(fmt.Sprintf("  ✓ MariaDB service (%s)", r.installation.ServiceName))
	}

	if r.config.RemoveData && r.installation.DataDirectoryExists {
		size := r.detectionService.formatSize(r.installation.DataDirectorySize)
		terminal.PrintInfo(fmt.Sprintf("  ✓ Data directories (%s)", size))
	}

	if len(r.installation.ConfigFiles) > 0 {
		terminal.PrintInfo(fmt.Sprintf("  ✓ Configuration files (%d files)", len(r.installation.ConfigFiles)))
	}

	if len(r.installation.LogFiles) > 0 {
		terminal.PrintInfo(fmt.Sprintf("  ✓ Log files (%d files)", len(r.installation.LogFiles)))
	}

	if r.config.RemoveRepositories {
		terminal.PrintInfo("  ✓ MariaDB repositories")
	}

	// Show what will be preserved
	if !r.config.RemoveData {
		terminal.PrintWarning("\nData directories will be preserved (use --remove-data to remove)")
	}

	if r.config.BackupData && r.installation.DataDirectoryExists {
		terminal.PrintInfo(fmt.Sprintf("\nData will be backed up to: %s", r.config.BackupPath))
	}

	// Confirm removal unless auto-confirm is enabled
	if !r.config.AutoConfirm {
		terminal.PrintWarning("\nThis will permanently remove MariaDB from your system!")
		fmt.Print("Do you want to proceed with the removal? (y/N): ")

		var response string
		fmt.Scanln(&response)

		if response != "y" && response != "Y" && response != "yes" && response != "YES" {
			lg.Info("Removal cancelled by user")
			return fmt.Errorf("removal cancelled by user")
		}
	}

	lg.Info("Removal confirmed by user")
	return nil
}

// backupData creates a backup of MariaDB data if requested
func (r *RemovalRunner) backupData() error {
	if !r.config.BackupData || !r.installation.DataDirectoryExists {
		return nil
	}

	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Creating data backup...")
	spinner.Start()

	// Set default backup path if not specified
	if r.config.BackupPath == "" {
		homeDir, _ := os.UserHomeDir()
		r.config.BackupPath = filepath.Join(homeDir, "mariadb_backups")
	}

	err := r.backupService.BackupData(r.installation, r.config.BackupPath)
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("failed to backup data: %w", err)
	}

	terminal.PrintSuccess(fmt.Sprintf("Data backed up to: %s", r.config.BackupPath))
	lg.Info("Data backup completed successfully",
		logger.String("backup_path", r.config.BackupPath))

	return nil
}

// stopServices stops and disables MariaDB services
func (r *RemovalRunner) stopServices() error {
	if r.installation.ServiceName == "" {
		return nil
	}

	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Stopping MariaDB services...")
	spinner.Start()

	err := r.removalService.StopAndDisableServices(r.installation)
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	terminal.PrintSuccess("MariaDB services stopped and disabled")
	lg.Info("Services stopped and disabled successfully")

	return nil
}

// removePackages removes MariaDB packages
func (r *RemovalRunner) removePackages() error {
	if !r.installation.IsInstalled {
		return nil
	}

	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Removing MariaDB packages...")
	spinner.Start()

	err := r.removalService.RemovePackages(r.installation)
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("failed to remove packages: %w", err)
	}

	terminal.PrintSuccess("MariaDB packages removed successfully")
	lg.Info("Packages removed successfully")

	return nil
}

// removeDataDirectories removes MariaDB data directories
func (r *RemovalRunner) removeDataDirectories() error {
	lg, _ := logger.Get()

	if !r.config.RemoveData {
		lg.Info("Data removal is disabled, skipping")
		return nil
	}

	spinner := terminal.NewProgressSpinner("Removing data directories...")
	spinner.Start()

	err := r.removalService.RemoveDataDirectories(r.installation, r.config)
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("failed to remove data directories: %w", err)
	}

	terminal.PrintSuccess("Data directories removed successfully")
	lg.Info("Data directories removed successfully")

	return nil
}

// removeConfigFiles removes MariaDB configuration files
func (r *RemovalRunner) removeConfigFiles() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Removing configuration files...")
	spinner.Start()

	err := r.removalService.RemoveConfigFiles(r.installation)
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("failed to remove configuration files: %w", err)
	}

	terminal.PrintSuccess("Configuration files removed successfully")
	lg.Info("Configuration files removed successfully")

	return nil
}

// removeLogFiles removes MariaDB log files
func (r *RemovalRunner) removeLogFiles() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Removing log files...")
	spinner.Start()

	err := r.removalService.RemoveLogFiles(r.installation)
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("failed to remove log files: %w", err)
	}

	terminal.PrintSuccess("Log files removed successfully")
	lg.Info("Log files removed successfully")

	return nil
}

// removeRepositories removes MariaDB repositories
func (r *RemovalRunner) removeRepositories() error {
	lg, _ := logger.Get()

	if !r.config.RemoveRepositories {
		lg.Info("Repository removal is disabled, skipping")
		return nil
	}

	spinner := terminal.NewProgressSpinner("Removing MariaDB repositories...")
	spinner.Start()

	err := r.removalService.RemoveRepositories(r.config)
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("failed to remove repositories: %w", err)
	}

	terminal.PrintSuccess("MariaDB repositories removed successfully")
	lg.Info("Repositories removed successfully")

	return nil
}

// cleanupResidualFiles removes any remaining MariaDB files
func (r *RemovalRunner) cleanupResidualFiles() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Cleaning up residual files...")
	spinner.Start()

	err := r.removalService.CleanupResidualFiles()
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("failed to cleanup residual files: %w", err)
	}

	terminal.PrintSuccess("Residual files cleaned up successfully")
	lg.Info("Residual cleanup completed successfully")

	return nil
}

// showRemovalSummary shows a summary of what was removed
func (r *RemovalRunner) showRemovalSummary() {
	separator := strings.Repeat("=", 70)

	terminal.PrintInfo("\n" + separator)
	terminal.PrintInfo("REMOVAL SUMMARY")
	terminal.PrintInfo(separator)

	if r.installation.IsInstalled {
		terminal.PrintSuccess(fmt.Sprintf("✓ Removed MariaDB %s", r.installation.Version))
	}

	if r.installation.ServiceName != "" {
		terminal.PrintSuccess(fmt.Sprintf("✓ Stopped and disabled service: %s", r.installation.ServiceName))
	}

	if r.config.BackupData && r.installation.DataDirectoryExists {
		terminal.PrintSuccess(fmt.Sprintf("✓ Data backed up to: %s", r.config.BackupPath))
	}

	if r.config.RemoveData {
		terminal.PrintSuccess("✓ Data directories removed")
	} else {
		terminal.PrintInfo("ℹ️ Data directories preserved")
	}

	if len(r.installation.ConfigFiles) > 0 {
		terminal.PrintSuccess(fmt.Sprintf("✓ Removed %d configuration files", len(r.installation.ConfigFiles)))
	}

	if len(r.installation.LogFiles) > 0 {
		terminal.PrintSuccess(fmt.Sprintf("✓ Removed %d log files", len(r.installation.LogFiles)))
	}

	if r.config.RemoveRepositories {
		terminal.PrintSuccess("✓ MariaDB repositories removed")
	}

	terminal.PrintSuccess("✓ Residual files cleaned up")
	terminal.PrintInfo(separator)
	terminal.PrintSuccess("MariaDB has been completely removed from your system!")
}
