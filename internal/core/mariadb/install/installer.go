package install

import (
	"fmt"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/repository"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// Installer handles MariaDB installation operations
type Installer struct {
	config           *Config
	systemInfo       *SystemInfo
	systemChecker    *SystemChecker
	versionSelector  *VersionSelector
	repositorySetup  *RepositorySetupManager
	packageInstaller *PackageInstaller
	serviceManager   *ServiceConfigManager
	validator        *InstallationValidator
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

	// Initialize service providers
	pkgManager := system.NewPackageManager()
	svcManager := system.NewServiceManager()
	repoManager := repository.NewManager(osInfo)

	// Initialize specialized components
	systemChecker := NewSystemChecker(systemInfo, pkgManager, svcManager, repoManager)
	versionSelector := NewVersionSelector(config, osInfo)
	repositorySetup := NewRepositorySetupManager(repoManager)
	packageInstaller := NewPackageInstaller(pkgManager, osInfo)
	serviceManager := NewServiceConfigManager(svcManager)
	validator := NewInstallationValidator()

	installer := &Installer{
		config:           config,
		systemInfo:       systemInfo,
		systemChecker:    systemChecker,
		versionSelector:  versionSelector,
		repositorySetup:  repositorySetup,
		packageInstaller: packageInstaller,
		serviceManager:   serviceManager,
		validator:        validator,
	}

	return installer, nil
}

// Install performs the complete MariaDB installation process
func (i *Installer) Install() (*InstallResult, error) {
	startTime := time.Now()

	// Validate configuration and log start
	if err := i.validator.ValidateConfig(i.config); err != nil {
		return i.validator.CreateErrorResult(err, "Configuration validation", startTime), err
	}

	if err := i.validator.LogInstallationStart(); err != nil {
		return i.validator.CreateErrorResult(err, "Logging initialization", startTime), err
	}

	// Clear screen and show header
	terminal.ClearAndShowHeader("MariaDB Installation")

	// Step 1: System checks
	if err := i.systemChecker.PerformAllChecks(); err != nil {
		return i.validator.CreateErrorResult(err, "System checks", startTime), err
	}

	// Step 2: Check existing installation
	if err := i.systemChecker.CheckExistingInstallation(); err != nil {
		return i.validator.CreateErrorResult(err, "Existing installation check", startTime), err
	}

	// Step 3: Get available versions and let user choose
	selectedVersion, err := i.versionSelector.SelectVersion()
	if err != nil {
		return i.validator.CreateErrorResult(err, "Version selection", startTime), err
	}

	// Step 4: Setup repository
	if err := i.repositorySetup.SetupRepository(selectedVersion); err != nil {
		return i.validator.CreateErrorResult(err, "Repository setup", startTime), err
	}

	// Step 5: Install packages
	installedCount, err := i.packageInstaller.InstallPackages()
	if err != nil {
		return i.validator.CreateErrorResult(err, "Package installation", startTime), err
	}

	// Step 6: Start and enable service
	if err := i.serviceManager.ConfigureService(); err != nil {
		return i.validator.CreateErrorResult(err, "Service configuration", startTime), err
	}

	// Step 7: Verify installation
	serviceStatus, err := i.serviceManager.VerifyInstallation()
	if err != nil {
		return i.validator.CreateErrorResult(err, "Installation verification", startTime), err
	}

	// Create success result
	result := i.validator.CreateSuccessResult(selectedVersion, installedCount, serviceStatus, startTime)
	i.validator.LogInstallationSuccess(selectedVersion, result.Duration)

	return result, nil
}
