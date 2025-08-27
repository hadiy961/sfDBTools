package install

import (
	"fmt"
	"os/exec"
	"strings"

	"sfDBTools/internal/core/mariadb/check_version"
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
	isInstalled, version, err := r.packageManager.IsInstalled("mariadb-server")
	if err != nil {
		spinner.Stop()
		lg.Warn("Failed to check existing installation", logger.Error(err))
	}

	spinner.Stop()

	if isInstalled {
		terminal.PrintWarning(fmt.Sprintf("MariaDB is already installed (version: %s)", version))

		if !r.config.RemoveExisting && !r.config.AutoConfirm {
			if !r.confirmRemoveExisting() {
				return fmt.Errorf("installation cancelled: MariaDB is already installed")
			}
		}

		if r.config.RemoveExisting || r.config.AutoConfirm {
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

// removeExistingInstallation removes existing MariaDB installation
func (r *InstallRunner) removeExistingInstallation() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Removing existing MariaDB installation...")
	spinner.Start()

	// Stop MariaDB service if running
	if err := r.stopMariaDBService(); err != nil {
		lg.Warn("Failed to stop MariaDB service", logger.Error(err))
	}

	// Remove packages
	packages := []string{"mariadb-server", "mariadb-client", "mariadb-common"}
	for _, pkg := range packages {
		if err := r.packageManager.Remove(pkg); err != nil {
			lg.Warn("Failed to remove package", logger.String("package", pkg), logger.Error(err))
		}
	}

	spinner.Stop()
	terminal.PrintSuccess("Existing MariaDB installation removed")

	lg.Info("Existing installation removed successfully")
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

	spinner := terminal.NewProgressSpinner("Configuring MariaDB repository...")
	spinner.Start()

	// Create repository setup manager
	r.repoSetupManager = NewRepoSetupManager(r.osInfo)

	// Clean existing repositories to avoid conflicts
	if err := r.repoSetupManager.CleanRepositories(); err != nil {
		lg.Warn("Failed to clean existing repositories", logger.Error(err))
	}

	// Check if setup script is available
	available, err := r.repoSetupManager.IsScriptAvailable()
	if err != nil || !available {
		spinner.Stop()
		return fmt.Errorf("MariaDB repository setup script is not available: %w", err)
	}

	// Setup repository using official script
	if err := r.repoSetupManager.SetupRepository(r.selectedVersion.Version); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to setup repository: %w", err)
	}

	// Update package cache
	if err := r.repoSetupManager.UpdatePackageCache(); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to update package cache: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess("MariaDB repository configured successfully")

	lg.Info("Repository configuration completed",
		logger.String("version", r.selectedVersion.Version))

	return nil
}

// installMariaDB performs the actual installation
func (r *InstallRunner) installMariaDB() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner(fmt.Sprintf("Installing MariaDB %s...", r.selectedVersion.LatestVersion))
	spinner.Start()

	packageName := r.packageManager.GetPackageName(r.selectedVersion.Version)
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
