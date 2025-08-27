package install

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"sfDBTools/internal/core/mariadb/check_version"
	"sfDBTools/internal/core/mariadb/configure"
	"sfDBTools/internal/core/mariadb/remove"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/terminal"
)

// InstallRunner orchestrates the MariaDB installation process
type InstallRunner struct {
	config           *InstallConfig
	versionService   *check_version.VersionService
	versionSelector  *VersionSelector
	repoSetupManager *RepoSetupManager
	packageManager   PackageManager
	osInfo           *common.OSInfo
	selectedVersion  *SelectableVersion
}

// NewInstallRunner creates a new installation runner
func NewInstallRunner(config *InstallConfig) *InstallRunner {
	if config == nil {
		config = DefaultInstallConfig()
	}

	return &InstallRunner{
		config: config,
	}
}

// Run executes the complete MariaDB installation process
func (r *InstallRunner) Run() error {
	lg, _ := logger.Get()

	lg.Info("Starting MariaDB installation process")
	terminal.PrintHeader("MariaDB Installation")

	// Step 1: Check OS compatibility
	if err := r.checkOSCompatibility(); err != nil {
		return fmt.Errorf("OS compatibility check failed: %w", err)
	}

	// Step 2: Check internet connectivity
	if err := r.checkInternetConnectivity(); err != nil {
		return fmt.Errorf("internet connectivity check failed: %w", err)
	}

	// Step 3: Fetch available versions
	if err := r.fetchAvailableVersions(); err != nil {
		return fmt.Errorf("failed to fetch available versions: %w", err)
	}

	// Step 4: Select version
	if err := r.selectVersion(); err != nil {
		return fmt.Errorf("version selection failed: %w", err)
	}

	// Step 5: Check for existing installation
	if err := r.checkExistingInstallation(); err != nil {
		return fmt.Errorf("existing installation check failed: %w", err)
	}

	// Step 6: Configure repository
	if err := r.configureRepository(); err != nil {
		return fmt.Errorf("repository configuration failed: %w", err)
	}

	// Step 7: Install MariaDB
	if err := r.installMariaDB(); err != nil {
		return fmt.Errorf("MariaDB installation failed: %w", err)
	}

	// Step 8: Post-installation setup
	if err := r.postInstallationSetup(); err != nil {
		terminal.PrintWarning("Post-installation setup had issues, but MariaDB was installed successfully")
		lg.Warn("Post-installation setup failed", logger.Error(err))
	}

	terminal.PrintSuccess("MariaDB installation completed successfully!")
	lg.Info("MariaDB installation completed successfully",
		logger.String("version", r.selectedVersion.LatestVersion))

	return nil
}

// checkOSCompatibility checks if the OS is supported
func (r *InstallRunner) checkOSCompatibility() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Detecting operating system...")
	spinner.Start()

	// Detect OS using common utility
	detector := common.NewOSDetector()
	osInfo, err := detector.DetectOS()
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to detect OS: %w", err)
	}

	r.osInfo = osInfo

	// Check OS compatibility using MariaDB supported OS list
	supportedOS := common.MariaDBSupportedOS()
	if err := common.ValidateOperatingSystem(supportedOS); err != nil {
		spinner.Stop()
		return fmt.Errorf("OS compatibility check failed: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess(fmt.Sprintf("Operating system detected: %s %s (%s)",
		strings.ToUpper(string(osInfo.ID[0]))+osInfo.ID[1:], osInfo.Version, osInfo.Architecture))

	lg.Info("OS compatibility check passed",
		logger.String("os", osInfo.ID),
		logger.String("version", osInfo.Version),
		logger.String("arch", osInfo.Architecture))

	return nil
} // checkInternetConnectivity verifies internet connection
func (r *InstallRunner) checkInternetConnectivity() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Checking internet connectivity...")
	spinner.Start()

	if err := common.CheckInternetConnectivity(); err != nil {
		spinner.Stop()
		return fmt.Errorf("internet connectivity check failed: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess("Internet connectivity verified")

	lg.Info("Internet connectivity check passed")
	return nil
}

// fetchAvailableVersions retrieves available MariaDB versions
func (r *InstallRunner) fetchAvailableVersions() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Fetching available MariaDB versions...")
	spinner.Start()

	// Create version service
	versionConfig := check_version.DefaultCheckVersionConfig()
	r.versionService = check_version.NewVersionService(versionConfig)

	// Fetch versions
	versions, err := r.versionService.FetchAvailableVersions()
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to fetch MariaDB versions: %w", err)
	}

	if len(versions) == 0 {
		spinner.Stop()
		return fmt.Errorf("no MariaDB versions available for installation")
	}

	spinner.Stop()
	terminal.PrintSuccess(fmt.Sprintf("Found %d available MariaDB versions", len(versions)))

	// Convert to selectable versions
	selectableVersions := ConvertVersionInfo(versions)
	r.versionSelector = NewVersionSelector(selectableVersions)

	lg.Info("Available versions fetched successfully", logger.Int("count", len(versions)))
	return nil
}

// selectVersion handles version selection
func (r *InstallRunner) selectVersion() error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Please select a MariaDB version to install:")

	selectedVersion, err := r.versionSelector.SelectVersion(r.config.AutoConfirm, r.config.Version)
	if err != nil {
		return fmt.Errorf("version selection failed: %w", err)
	}

	r.selectedVersion = selectedVersion

	terminal.PrintSuccess(fmt.Sprintf("Selected MariaDB version: %s (%s)",
		selectedVersion.Version, selectedVersion.LatestVersion))

	lg.Info("Version selected",
		logger.String("major_version", selectedVersion.Version),
		logger.String("latest_version", selectedVersion.LatestVersion))

	return nil
}

// checkExistingInstallation checks if MariaDB is already installed
func (r *InstallRunner) checkExistingInstallation() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Checking for existing MariaDB installation...")
	spinner.Start()

	// Create package manager
	r.packageManager = NewPackageManager(r.osInfo)
	if r.packageManager == nil {
		spinner.Stop()
		return fmt.Errorf("unsupported package manager for OS: %s", r.osInfo.ID)
	}

	// Check if MariaDB is installed
	// Try different package names that might be used
	packageNames := []string{"MariaDB-server", "mariadb-server", "mariadb"}

	var isInstalled bool
	var version string
	var checkErr error

	for _, pkgName := range packageNames {
		isInstalled, version, checkErr = r.packageManager.IsInstalled(pkgName)
		if checkErr == nil && isInstalled {
			break // Found installed package
		}
	}

	if checkErr != nil {
		spinner.Stop()
		lg.Warn("Failed to check existing installation", logger.Error(checkErr))
		// Return error to avoid proceeding with installation when check fails
		return fmt.Errorf("unable to verify existing installation status: %w", checkErr)
	}

	spinner.Stop()

	if isInstalled {
		terminal.PrintWarning(fmt.Sprintf("MariaDB is already installed (version: %s)", version))

		var shouldRemove bool

		// Determine if we should remove existing installation
		if r.config.RemoveExisting {
			// Flag explicitly set
			shouldRemove = true
		} else if r.config.AutoConfirm {
			// Auto-confirm mode - remove existing
			shouldRemove = true
		} else {
			// Ask user for confirmation
			shouldRemove = r.confirmRemoveExisting()
			if !shouldRemove {
				return fmt.Errorf("installation cancelled: MariaDB is already installed")
			}
		}

		// Remove existing installation if confirmed
		if shouldRemove {
			if err := r.removeExistingInstallation(); err != nil {
				return fmt.Errorf("failed to remove existing installation: %w", err)
			}
		}
	} else {
		terminal.PrintSuccess("No existing MariaDB installation found")
	}

	lg.Info("Existing installation check completed", logger.Bool("was_installed", isInstalled))
	return nil
}

// confirmRemoveExisting asks user to confirm removal of existing installation
func (r *InstallRunner) confirmRemoveExisting() bool {
	fmt.Print("Remove existing MariaDB installation? (y/N): ")

	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// removeExistingInstallation removes existing MariaDB installation using remove module
func (r *InstallRunner) removeExistingInstallation() error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Using comprehensive removal process...")

	// Create remove configuration for installation context
	removeConfig := &remove.RemovalConfig{
		RemoveData:         false, // Keep data during installation removal
		BackupData:         true,  // Always backup data for safety
		BackupPath:         "",    // Use default backup path
		RemoveRepositories: false, // Keep repositories for new installation
		AutoConfirm:        true,  // Auto-confirm since we're in install flow
		DataDirectory:      "",    // Auto-detect
		ConfigDirectory:    "",    // Auto-detect
		LogDirectory:       "",    // Auto-detect
	}

	// Create remove runner
	removeRunner := remove.NewRemovalRunner(removeConfig)

	// Execute removal process
	if err := removeRunner.Run(); err != nil {
		return fmt.Errorf("comprehensive removal failed: %w", err)
	}

	terminal.PrintSuccess("Existing MariaDB installation removed successfully using comprehensive removal")
	lg.Info("Existing installation removed using remove module")
	return nil
}

// stopMariaDBService stops the MariaDB service
func (r *InstallRunner) stopMariaDBService() error {
	services := []string{"mariadb", "mysql", "mysqld"}

	for _, service := range services {
		cmd := exec.Command("systemctl", "stop", service)
		if err := cmd.Run(); err == nil {
			return nil // Successfully stopped
		}
	}

	return fmt.Errorf("failed to stop any MariaDB service")
}

// installMariaDB performs the actual installation

// configureRepository sets up MariaDB repository
func (r *InstallRunner) configureRepository() error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Configuring MariaDB repository...")

	// Step 1: Initialize repository setup manager
	spinner := terminal.NewProgressSpinner("Initializing repository setup...")
	spinner.Start()

	// Create repository setup manager
	r.repoSetupManager = NewRepoSetupManager(r.osInfo)

	spinner.Stop()
	terminal.PrintSuccess("Repository setup manager initialized")

	// Step 2: Clean existing repositories
	spinner = terminal.NewProgressSpinner("Cleaning existing MariaDB repositories...")
	spinner.Start()

	if err := r.repoSetupManager.CleanRepositories(); err != nil {
		spinner.Stop()
		lg.Warn("Failed to clean existing repositories", logger.Error(err))
		terminal.PrintWarning("Failed to clean existing repositories, continuing...")
	} else {
		spinner.Stop()
		terminal.PrintSuccess("Existing MariaDB repositories cleaned")
	}

	// Step 3: Check repository setup script availability
	spinner = terminal.NewProgressSpinner("Verifying MariaDB repository setup script...")
	spinner.Start()

	available, err := r.repoSetupManager.IsScriptAvailable()
	if err != nil || !available {
		spinner.Stop()
		return fmt.Errorf("MariaDB repository setup script is not available: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess("Repository setup script verified")

	// Step 4: Setup repository using official script
	spinner = terminal.NewProgressSpinner(fmt.Sprintf("Setting up MariaDB %s repository (this may take a moment)...", r.selectedVersion.Version))
	spinner.Start()

	if err := r.repoSetupManager.SetupRepository(r.selectedVersion.Version); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to setup repository: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess("MariaDB repository configured successfully")

	// // Step 5: Update package cache
	// spinner = terminal.NewProgressSpinner("Updating package cache...")
	// spinner.Start()

	// if err := r.repoSetupManager.UpdatePackageCache(); err != nil {
	// 	spinner.Stop()
	// 	return fmt.Errorf("failed to update package cache: %w", err)
	// }

	spinner.Stop()
	terminal.PrintSuccess("Package cache updated successfully")

	lg.Info("Repository configuration completed",
		logger.String("version", r.selectedVersion.Version))

	return nil
}

// installMariaDB performs the actual installation
func (r *InstallRunner) installMariaDB() error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Installing MariaDB packages...")

	// Step 1: Determine package name
	spinner := terminal.NewProgressSpinner("Determining MariaDB package name...")
	spinner.Start()

	packageName := r.packageManager.GetPackageName(r.selectedVersion.Version)

	spinner.Stop()
	terminal.PrintSuccess(fmt.Sprintf("Package determined: %s", packageName))

	// Step 2: Install MariaDB packages
	spinner = terminal.NewProgressSpinner(fmt.Sprintf("Installing MariaDB %s (this may take several minutes)...", r.selectedVersion.LatestVersion))
	spinner.Start()

	if err := r.packageManager.Install(packageName, r.selectedVersion.Version); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to install MariaDB: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess(fmt.Sprintf("MariaDB %s installed successfully", r.selectedVersion.LatestVersion))

	lg.Info("MariaDB installation completed",
		logger.String("package", packageName),
		logger.String("version", r.selectedVersion.LatestVersion))

	return nil
}

// postInstallationSetup performs post-installation configuration
func (r *InstallRunner) postInstallationSetup() error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Performing post-installation setup...")

	// Check for data directory conflicts before starting service
	if err := r.checkDataDirectoryCompatibility(); err != nil {
		lg.Warn("Data directory compatibility issue detected", logger.Error(err))
		terminal.PrintWarning("Data directory compatibility issue detected")

		if err := r.handleDataDirectoryConflict(); err != nil {
			return fmt.Errorf("failed to resolve data directory conflict: %w", err)
		}
	}

	// Start MariaDB service
	if r.config.StartService {
		if err := r.startMariaDBService(); err != nil {
			lg.Error("Failed to start MariaDB service", logger.Error(err))
			return fmt.Errorf("failed to start MariaDB service: %w", err)
		}
		terminal.PrintSuccess("MariaDB service started successfully")
	}

	// Enable service on boot
	if err := r.enableMariaDBService(); err != nil {
		lg.Warn("Failed to enable MariaDB service on boot", logger.Error(err))
		terminal.PrintWarning("MariaDB service may not start automatically on boot")
	} else {
		terminal.PrintSuccess("MariaDB service enabled on boot")
	}

	// Run security setup if enabled
	if r.config.EnableSecurity {
		terminal.PrintInfo("Security setup will need to be run manually using: mysql_secure_installation")
	}

	// Ask if user wants to configure MariaDB with custom settings
	if r.shouldRunConfiguration() {
		if err := r.runMariaDBConfiguration(); err != nil {
			lg.Warn("MariaDB configuration failed", logger.Error(err))
			terminal.PrintWarning("MariaDB configuration had issues, but installation was successful")
			terminal.PrintInfo("You can run configuration manually using: sfdbtools mariadb configure")
		} else {
			terminal.PrintSuccess("MariaDB configuration completed successfully")
		}
	} else {
		terminal.PrintInfo("MariaDB configuration skipped")
		terminal.PrintInfo("You can run configuration later using: sfdbtools mariadb configure")
	}

	lg.Info("Post-installation setup completed")
	return nil
}

// startMariaDBService starts the MariaDB service
func (r *InstallRunner) startMariaDBService() error {
	services := []string{"mariadb", "mysql"}

	for _, service := range services {
		cmd := exec.Command("systemctl", "start", service)
		if err := cmd.Run(); err == nil {
			return nil // Successfully started
		}
	}

	return fmt.Errorf("failed to start any MariaDB service")
}

// enableMariaDBService enables MariaDB service on boot
func (r *InstallRunner) enableMariaDBService() error {
	services := []string{"mariadb", "mysql"}

	for _, service := range services {
		cmd := exec.Command("systemctl", "enable", service)
		if err := cmd.Run(); err == nil {
			return nil // Successfully enabled
		}
	}

	return fmt.Errorf("failed to enable any MariaDB service")
}

// shouldRunConfiguration asks user if they want to configure MariaDB
func (r *InstallRunner) shouldRunConfiguration() bool {
	// Skip prompt if auto-confirm is enabled (don't auto-configure)
	if r.config.AutoConfirm {
		return false // Let user run configuration manually later
	}

	terminal.PrintInfo("\nMariaDB has been installed successfully!")
	terminal.PrintInfo("Would you like to configure MariaDB with custom settings now?")
	terminal.PrintInfo("This will setup:")
	terminal.PrintInfo("  • Custom data and binlog directories")
	terminal.PrintInfo("  • Configuration file optimization")
	terminal.PrintInfo("  • Firewall and SELinux settings")
	terminal.PrintInfo("  • Default databases and users")

	fmt.Print("Run MariaDB configuration now? (y/N): ")

	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// runMariaDBConfiguration executes the MariaDB configuration process
func (r *InstallRunner) runMariaDBConfiguration() error {
	lg, _ := logger.Get()

	lg.Info("Starting MariaDB configuration from install process")
	terminal.PrintInfo("Configuring MariaDB with optimized settings...")

	// Create configure config with auto-confirm mode to avoid double prompting
	configureConfig := &configure.ConfigureConfig{
		AutoConfirm:   false, // Auto-confirm since user already agreed
		SkipUserSetup: false, // Include user setup
		SkipDBSetup:   false, // Include database setup
	}

	// Create configure runner
	configRunner := configure.NewConfigureRunner(configureConfig)

	// Run configuration
	if err := configRunner.Run(); err != nil {
		return fmt.Errorf("MariaDB configuration failed: %w", err)
	}

	lg.Info("MariaDB configuration completed from install process")
	return nil
}

// checkDataDirectoryCompatibility checks if existing data directory is compatible
func (r *InstallRunner) checkDataDirectoryCompatibility() error {
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
	currentVersion := r.selectedVersion.LatestVersion

	lg.Info("Checking data directory compatibility",
		logger.String("existing_version", existingVersion),
		logger.String("installing_version", currentVersion))

	// Parse versions for comparison
	if r.isVersionDowngrade(existingVersion, currentVersion) {
		return fmt.Errorf("data directory contains newer version (%s), cannot downgrade to %s",
			existingVersion, currentVersion)
	}

	return nil
}

// isVersionDowngrade checks if target version is older than existing version
func (r *InstallRunner) isVersionDowngrade(existing, target string) bool {
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

// parseVersionParts extracts major and minor version numbers
func parseVersionParts(parts []string) (int, int) {
	major, minor := 0, 0

	if len(parts) >= 1 {
		if val, err := strconv.Atoi(parts[0]); err == nil {
			major = val
		}
	}
	if len(parts) >= 2 {
		if val, err := strconv.Atoi(parts[1]); err == nil {
			minor = val
		}
	}

	return major, minor
}

// handleDataDirectoryConflict resolves data directory version conflicts
func (r *InstallRunner) handleDataDirectoryConflict() error {
	terminal.PrintWarning("Incompatible MariaDB data directory detected!")
	terminal.PrintInfo("The existing data directory contains a newer MariaDB version.")
	terminal.PrintInfo("Options:")
	terminal.PrintInfo("  1. Backup and reinitialize data directory (recommended)")
	terminal.PrintInfo("  2. Cancel installation")

	if r.config.AutoConfirm {
		terminal.PrintInfo("Auto-confirm enabled: backing up and reinitializing data directory")
		return r.backupAndReinitializeDataDirectory()
	}

	fmt.Print("Backup and reinitialize data directory? (y/N): ")
	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	if response == "y" || response == "yes" {
		return r.backupAndReinitializeDataDirectory()
	}

	return fmt.Errorf("installation cancelled due to data directory conflict")
}

// backupAndReinitializeDataDirectory backs up existing data and creates fresh data directory
func (r *InstallRunner) backupAndReinitializeDataDirectory() error {
	lg, _ := logger.Get()

	dataDir := "/var/lib/mysql"
	backupDir := fmt.Sprintf("/var/lib/mysql.backup.%s",
		time.Now().Format("20060102_150405"))

	// Stop MariaDB service if running
	spinner := terminal.NewProgressSpinner("Stopping MariaDB service...")
	spinner.Start()

	_ = r.stopMariaDBService() // Ignore error if not running

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

// GetInstallationSteps returns the list of installation steps for progress tracking
func (r *InstallRunner) GetInstallationSteps() []InstallationStep {
	return []InstallationStep{
		{Name: "os_check", Description: "Checking OS compatibility", Required: true},
		{Name: "internet_check", Description: "Checking internet connectivity", Required: true},
		{Name: "fetch_versions", Description: "Fetching available versions", Required: true},
		{Name: "select_version", Description: "Selecting MariaDB version", Required: true},
		{Name: "check_existing", Description: "Checking existing installation", Required: true},
		{Name: "configure_repo", Description: "Configuring repository", Required: true},
		{Name: "install", Description: "Installing MariaDB", Required: true},
		{Name: "post_setup", Description: "Post-installation setup", Required: false},
	}
}
