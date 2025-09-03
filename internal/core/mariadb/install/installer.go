package install

import (
	"fmt"
	"time"

	"sfDBTools/internal/core/mariadb/check_version"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/repository"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// Installer handles MariaDB installation operations
type Installer struct {
	config      *Config
	systemInfo  *SystemInfo
	pkgManager  system.PackageManager
	svcManager  system.ServiceManager
	repoManager *repository.Manager
}

// NewInstaller creates a new installer instance
func NewInstaller(config *Config) (*Installer, error) {
	if config == nil {
		config = DefaultConfig()
	}

	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	// Detect OS
	osDetector := common.NewOSDetector()
	osInfo, err := osDetector.DetectOS()
	if err != nil {
		lg.Error("Failed to detect OS", logger.Error(err))
		return nil, fmt.Errorf("failed to detect OS: %w", err)
	}

	systemInfo := NewSystemInfo()
	systemInfo.OSInfo = osInfo

	installer := &Installer{
		config:      config,
		systemInfo:  systemInfo,
		pkgManager:  system.NewPackageManager(),
		svcManager:  system.NewServiceManager(),
		repoManager: repository.NewManager(osInfo),
	}

	return installer, nil
}

// Install performs the complete MariaDB installation process
func (i *Installer) Install() (*InstallResult, error) {
	startTime := time.Now()
	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting MariaDB installation process")

	// Clear screen and show header
	terminal.ClearAndShowHeader("MariaDB Installation")

	result := &InstallResult{
		InstalledAt: time.Now(),
	}

	// Step 1: System checks
	if err := i.performSystemChecks(); err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("System checks failed: %v", err)
		result.Duration = time.Since(startTime)
		return result, err
	}

	// Step 2: Check existing installation
	if err := i.checkExistingInstallation(); err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to check existing installation: %v", err)
		result.Duration = time.Since(startTime)
		return result, err
	}

	// Step 3: Get available versions and let user choose
	selectedVersion, err := i.selectVersion()
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Version selection failed: %v", err)
		result.Duration = time.Since(startTime)
		return result, err
	}

	result.Version = selectedVersion

	// Step 4: Setup repository
	if err := i.setupRepository(selectedVersion); err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Repository setup failed: %v", err)
		result.Duration = time.Since(startTime)
		return result, err
	}

	// Step 5: Install packages
	installedCount, err := i.installPackages()
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Package installation failed: %v", err)
		result.Duration = time.Since(startTime)
		return result, err
	}

	result.PackagesCount = installedCount

	// Step 6: Start and enable service
	if err := i.configureService(); err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Service configuration failed: %v", err)
		result.Duration = time.Since(startTime)
		return result, err
	}

	// Step 7: Verify installation
	serviceStatus, err := i.verifyInstallation()
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Installation verification failed: %v", err)
		result.Duration = time.Since(startTime)
		return result, err
	}

	result.ServiceStatus = serviceStatus
	result.Success = true
	result.Message = "MariaDB installed successfully"
	result.Duration = time.Since(startTime)

	lg.Info("MariaDB installation completed successfully",
		logger.String("version", selectedVersion),
		logger.String("duration", result.Duration.String()))

	return result, nil
}

// performSystemChecks performs initial system validation
func (i *Installer) performSystemChecks() error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Performing system checks...")

	// Check internet connectivity
	if err := common.CheckMariaDBConnectivity(); err != nil {
		lg.Error("Internet connectivity check failed", logger.Error(err))
		i.systemInfo.InternetAvailable = false
		terminal.PrintError("Internet connectivity check failed")
		return fmt.Errorf("internet connectivity required for installation: %w", err)
	}
	i.systemInfo.InternetAvailable = true
	lg.Debug("Internet connectivity check passed")

	// Check repository availability
	available, err := i.repoManager.IsAvailable()
	if err != nil || !available {
		lg.Error("Repository availability check failed", logger.Error(err))
		i.systemInfo.RepoAvailable = false
		terminal.PrintError("Repository availability check failed")
		return fmt.Errorf("MariaDB repository is not accessible: %w", err)
	}
	i.systemInfo.RepoAvailable = true
	lg.Debug("Repository availability check passed")

	terminal.PrintSuccess("System checks completed successfully")
	return nil
}

// checkExistingInstallation checks for existing MariaDB installations
func (i *Installer) checkExistingInstallation() error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Checking existing installations...")

	// Check for existing service
	i.systemInfo.ExistingService = i.svcManager.IsActive("mariadb") || i.svcManager.IsActive("mysql")

	// Check for existing packages
	packages, err := i.pkgManager.GetInstalledPackages()
	if err != nil {
		lg.Warn("Failed to get installed packages", logger.Error(err))
	} else {
		i.systemInfo.ExistingPackages = packages
	}

	if i.systemInfo.ExistingService {
		lg.Warn("Existing MariaDB/MySQL service detected; aborting installation")
		return fmt.Errorf("existing MariaDB/MySQL service detected")
	}

	if len(i.systemInfo.ExistingPackages) > 0 {
		terminal.PrintInfo(fmt.Sprintf("Found %d existing MariaDB/MySQL packages", len(i.systemInfo.ExistingPackages)))
	} else {
		terminal.PrintSuccess("No existing MariaDB/MySQL packages found")
	}

	return nil
}

// selectVersion allows user to select MariaDB version to install
func (i *Installer) selectVersion() (string, error) {
	lg, _ := logger.Get()

	// Use minimal spinner for version fetching since it might have logs
	spinner := terminal.NewProgressSpinnerWithStyle("Fetching available MariaDB versions...", terminal.SpinnerMinimal)
	spinner.Start()

	// Get available versions using existing check_version functionality
	checkerConfig := check_version.DefaultConfig()
	checker, err := check_version.NewChecker(checkerConfig)
	if err != nil {
		spinner.StopWithError("Failed to create version checker")
		return "", fmt.Errorf("failed to create version checker: %w", err)
	}

	result, err := checker.CheckAvailableVersions()
	if err != nil {
		spinner.StopWithError("Failed to fetch available versions")
		return "", fmt.Errorf("failed to get available versions: %w", err)
	}

	if len(result.AvailableVersions) == 0 {
		spinner.StopWithError("No MariaDB versions available")
		return "", fmt.Errorf("no MariaDB versions available")
	}

	spinner.StopWithSuccess("Available MariaDB versions retrieved")

	// Convert to menu options (only stable versions for installation)
	var stableVersions []string
	var versionOptions []string

	for _, version := range result.AvailableVersions {
		if version.Type == "stable" {
			stableVersions = append(stableVersions, version.Version)
			option := fmt.Sprintf("MariaDB %s (Stable)", version.Version)
			if version.Version == result.CurrentStable {
				option += " [Recommended]"
			}
			versionOptions = append(versionOptions, option)
		}
	}

	if len(versionOptions) == 0 {
		terminal.PrintError("No stable versions available for installation")
		return "", fmt.Errorf("no stable versions available for installation")
	}

	// Show version selection menu
	terminal.ClearAndShowHeader("MariaDB Version Selection")
	terminal.PrintInfo("Select a MariaDB version to install:")

	// If SkipConfirm is enabled, auto-select recommended/current stable version
	if i.config != nil && i.config.SkipConfirm {
		// prefer current stable
		if result.CurrentStable != "" {
			terminal.PrintInfo(fmt.Sprintf("Auto-selecting recommended version: MariaDB %s", result.CurrentStable))
			terminal.PrintSuccess(fmt.Sprintf("Selected MariaDB %s for installation", result.CurrentStable))
			return result.CurrentStable, nil
		}

		// fallback to first stable
		if len(stableVersions) > 0 {
			terminal.PrintInfo(fmt.Sprintf("Auto-selecting first stable version: MariaDB %s", stableVersions[0]))
			terminal.PrintSuccess(fmt.Sprintf("Selected MariaDB %s for installation", stableVersions[0]))
			return stableVersions[0], nil
		}
	}

	selected, err := terminal.ShowMenuAndClear("Available Versions", versionOptions)
	if err != nil {
		return "", fmt.Errorf("version selection failed: %w", err)
	}

	selectedVersion := stableVersions[selected-1]

	lg.Info("User selected version", logger.String("version", selectedVersion))
	terminal.PrintSuccess(fmt.Sprintf("Selected MariaDB %s for installation", selectedVersion))

	return selectedVersion, nil
}

// setupRepository sets up MariaDB repository for the selected version
func (i *Installer) setupRepository(version string) error {
	lg, _ := logger.Get()

	spinner := terminal.NewProcessingSpinner("Setting up MariaDB repository...")
	spinner.Start()

	// Clean existing repositories first
	spinner.UpdateMessage("Cleaning existing repositories...")
	if err := i.repoManager.Clean(); err != nil {
		lg.Warn("Failed to clean existing repositories", logger.Error(err))
	}

	// Setup official repository
	spinner.UpdateMessage(fmt.Sprintf("Setting up official MariaDB %s repository...", version))
	if err := i.repoManager.SetupOfficial(version); err != nil {
		spinner.StopWithError("Failed to setup repository")
		return fmt.Errorf("failed to setup repository: %w", err)
	}

	// Update package cache
	spinner.UpdateMessage("Updating package cache...")
	if err := i.repoManager.UpdateCache(); err != nil {
		spinner.StopWithError("Failed to update package cache")
		return fmt.Errorf("failed to update package cache: %w", err)
	}

	spinner.StopWithSuccess("Repository setup completed")
	return nil
}

// installPackages installs MariaDB server and client packages
func (i *Installer) installPackages() (int, error) {
	lg, _ := logger.Get()

	// Determine packages to install based on OS
	packages := i.getPackagesToInstall()

	spinner := terminal.NewInstallSpinner(fmt.Sprintf("Installing %d MariaDB packages...", len(packages)))
	spinner.Start()

	spinner.UpdateMessage(fmt.Sprintf("Installing packages: %v", packages))

	if err := i.pkgManager.Install(packages); err != nil {
		spinner.StopWithError("Package installation failed")
		return 0, fmt.Errorf("package installation failed: %w", err)
	}

	lg.Info("Packages installed successfully", logger.Strings("packages", packages))
	spinner.StopWithSuccess("Package installation completed")

	return len(packages), nil
}

// configureService starts and enables MariaDB service
func (i *Installer) configureService() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProcessingSpinner("Configuring MariaDB service...")
	spinner.Start()

	serviceName := "mariadb"

	// Start the service
	spinner.UpdateMessage("Starting MariaDB service...")
	if err := i.svcManager.Start(serviceName); err != nil {
		spinner.StopWithError("Failed to start service")
		return fmt.Errorf("failed to start service: %w", err)
	}

	// Enable the service
	spinner.UpdateMessage("Enabling MariaDB service...")
	if err := i.svcManager.Enable(serviceName); err != nil {
		spinner.StopWithError("Failed to enable service")
		return fmt.Errorf("failed to enable service: %w", err)
	}

	lg.Info("Service configured successfully", logger.String("service", serviceName))
	spinner.StopWithSuccess("Service configuration completed")

	return nil
}

// verifyInstallation verifies the installation was successful
func (i *Installer) verifyInstallation() (string, error) {
	lg, _ := logger.Get()

	spinner := terminal.NewLoadingSpinner("Verifying installation...")
	spinner.Start()

	serviceName := "mariadb"

	// Check service status
	spinner.UpdateMessage("Checking service status...")
	status, err := i.svcManager.GetStatus(serviceName)
	if err != nil {
		spinner.StopWithError("Failed to get service status")
		return "", fmt.Errorf("failed to get service status: %w", err)
	}

	if !status.Active || !status.Enabled {
		statusMsg := fmt.Sprintf("Service issues - Active: %v, Enabled: %v", status.Active, status.Enabled)
		spinner.StopWithWarning("Service is not properly configured")
		return statusMsg, fmt.Errorf("service is not properly configured")
	}

	lg.Info("Installation verification completed",
		logger.Bool("active", status.Active),
		logger.Bool("enabled", status.Enabled))

	spinner.StopWithSuccess("Installation verification completed")

	return "Active and Enabled", nil
}

// getPackagesToInstall returns the list of packages to install based on OS
func (i *Installer) getPackagesToInstall() []string {
	switch i.systemInfo.OSInfo.PackageType {
	case "deb":
		return []string{"mariadb-server", "mariadb-client"}
	case "rpm":
		return []string{"MariaDB-server", "MariaDB-client"}
	default:
		// Default to generic names
		return []string{"mariadb-server", "mariadb-client"}
	}
}
