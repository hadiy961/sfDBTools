package install

import (
	"fmt"
	"strings"

	"sfDBTools/internal/core/mariadb/remove"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/terminal"
)

// InstallManager handles the installation coordination
type InstallManager struct {
	config           *InstallConfig
	packageManager   PackageManager
	repoSetupManager *RepoSetupManager
	osInfo           *common.OSInfo
	selectedVersion  *SelectableVersion
}

// NewInstallManager creates a new install manager
func NewInstallManager(config *InstallConfig, osInfo *common.OSInfo, selectedVersion *SelectableVersion) *InstallManager {
	return &InstallManager{
		config:          config,
		osInfo:          osInfo,
		selectedVersion: selectedVersion,
	}
}

// CheckExistingInstallation checks if MariaDB is already installed
func (i *InstallManager) CheckExistingInstallation() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Checking for existing MariaDB installation...")
	spinner.Start()

	// Create package manager
	i.packageManager = NewPackageManager(i.osInfo)
	if i.packageManager == nil {
		spinner.Stop()
		return fmt.Errorf("unsupported package manager for OS: %s", i.osInfo.ID)
	}

	// Check if MariaDB is installed
	// Try different package names that might be used
	packageNames := []string{"MariaDB-server", "mariadb-server", "mariadb"}

	var isInstalled bool
	var version string
	var checkErr error

	for _, pkgName := range packageNames {
		isInstalled, version, checkErr = i.packageManager.IsInstalled(pkgName)
		if checkErr == nil && isInstalled {
			break // Found installed package
		}
	}

	if checkErr != nil {
		spinner.Stop()
		lg.Warn("Failed to check existing installation", logger.Error(checkErr))
		return fmt.Errorf("unable to verify existing installation status: %w", checkErr)
	}

	spinner.Stop()

	if isInstalled {
		terminal.PrintWarning(fmt.Sprintf("MariaDB is already installed (version: %s)", version))

		// Track whether user explicitly confirmed removal
		userConfirmed := false

		if !i.config.RemoveExisting && !i.config.AutoConfirm {
			userConfirmed = i.confirmRemoveExisting()
			if !userConfirmed {
				return fmt.Errorf("installation cancelled: MariaDB is already installed")
			}
		}

		// Remove if config requests it, auto-confirm is enabled, or user confirmed interactively
		if i.config.RemoveExisting || i.config.AutoConfirm || userConfirmed {
			if err := i.removeExistingInstallation(); err != nil {
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
func (i *InstallManager) confirmRemoveExisting() bool {
	fmt.Print("Remove existing MariaDB installation? (y/N): ")

	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// removeExistingInstallation removes existing MariaDB installation using remove module
func (i *InstallManager) removeExistingInstallation() error {
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

// ConfigureRepository sets up MariaDB repository
func (i *InstallManager) ConfigureRepository() error {
	lg, _ := logger.Get()

	// Validate prerequisites
	if i.osInfo == nil {
		return fmt.Errorf("cannot configure repository: osInfo is not set")
	}
	if i.selectedVersion == nil {
		return fmt.Errorf("cannot configure repository: selected version is not set")
	}

	terminal.PrintInfo("Configuring MariaDB repository...")

	// Step 1: Initialize repository setup manager
	spinner := terminal.NewProgressSpinner("Initializing repository setup...")
	spinner.Start()

	// Create repository setup manager
	i.repoSetupManager = NewRepoSetupManager(i.osInfo)

	spinner.Stop()
	terminal.PrintSuccess("Repository setup manager initialized")

	// Step 2: Clean existing repositories
	spinner = terminal.NewProgressSpinner("Cleaning existing MariaDB repositories...")
	spinner.Start()

	if err := i.repoSetupManager.CleanRepositories(); err != nil {
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

	available, err := i.repoSetupManager.IsScriptAvailable()
	if err != nil || !available {
		spinner.Stop()
		return fmt.Errorf("MariaDB repository setup script is not available: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess("Repository setup script verified")

	// Step 4: Setup repository using official script
	spinner = terminal.NewProgressSpinner(fmt.Sprintf("Setting up MariaDB %s repository (this may take a moment)...", i.selectedVersion.Version))
	spinner.Start()

	if err := i.repoSetupManager.SetupRepository(i.selectedVersion.Version); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to setup repository: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess("MariaDB repository configured successfully")

	// Step 5: Update package cache
	spinner = terminal.NewProgressSpinner("Updating package cache...")
	spinner.Start()

	if err := i.repoSetupManager.UpdatePackageCache(); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to update package cache: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess("Package cache updated successfully")

	lg.Info("Repository configuration completed",
		logger.String("version", i.selectedVersion.Version))

	return nil
}

// InstallMariaDB performs the actual installation
func (i *InstallManager) InstallMariaDB() error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Installing MariaDB packages...")

	// Step 1: Determine package name
	spinner := terminal.NewProgressSpinner("Determining MariaDB package name...")
	spinner.Start()

	packageName := i.packageManager.GetPackageName(i.selectedVersion.Version)

	spinner.Stop()
	terminal.PrintSuccess(fmt.Sprintf("Package determined: %s", packageName))

	// Step 2: Install MariaDB packages
	spinner = terminal.NewProgressSpinner(fmt.Sprintf("Installing MariaDB %s (this may take several minutes)...", i.selectedVersion.LatestVersion))
	spinner.Start()

	if err := i.packageManager.Install(packageName, i.selectedVersion.Version); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to install MariaDB: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess(fmt.Sprintf("MariaDB %s installed successfully", i.selectedVersion.LatestVersion))

	lg.Info("MariaDB installation completed",
		logger.String("package", packageName),
		logger.String("version", i.selectedVersion.LatestVersion))

	return nil
}
