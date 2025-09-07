package install

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/repository"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// SystemChecker handles system validation operations
type SystemChecker struct {
	systemInfo  *SystemInfo
	pkgManager  system.PackageManager
	svcManager  system.ServiceManager
	repoManager *repository.Manager
}

// NewSystemChecker creates a new system checker instance
func NewSystemChecker(systemInfo *SystemInfo, pkgManager system.PackageManager,
	svcManager system.ServiceManager, repoManager *repository.Manager) *SystemChecker {
	return &SystemChecker{
		systemInfo:  systemInfo,
		pkgManager:  pkgManager,
		svcManager:  svcManager,
		repoManager: repoManager,
	}
}

// PerformAllChecks performs all system validation checks
func (sc *SystemChecker) PerformAllChecks() error {
	terminal.PrintInfo("Performing system checks...")

	if err := sc.CheckInternetConnectivity(); err != nil {
		return err
	}

	if err := sc.CheckRepositoryAvailability(); err != nil {
		return err
	}

	terminal.PrintSuccess("System checks completed successfully")
	return nil
}

// CheckInternetConnectivity verifies internet connectivity
func (sc *SystemChecker) CheckInternetConnectivity() error {
	lg, _ := logger.Get()

	if err := common.CheckMariaDBConnectivity(); err != nil {
		lg.Error("Internet connectivity check failed", logger.Error(err))
		sc.systemInfo.InternetAvailable = false
		terminal.PrintError("Internet connectivity check failed")
		return fmt.Errorf("internet connectivity required for installation: %w", err)
	}

	sc.systemInfo.InternetAvailable = true
	lg.Debug("Internet connectivity check passed")
	return nil
}

// CheckRepositoryAvailability verifies MariaDB repository accessibility
func (sc *SystemChecker) CheckRepositoryAvailability() error {
	lg, _ := logger.Get()

	available, err := sc.repoManager.IsAvailable()
	if err != nil || !available {
		lg.Error("Repository availability check failed", logger.Error(err))
		sc.systemInfo.RepoAvailable = false
		terminal.PrintError("Repository availability check failed")
		return fmt.Errorf("MariaDB repository is not accessible: %w", err)
	}

	sc.systemInfo.RepoAvailable = true
	lg.Debug("Repository availability check passed")
	return nil
}

// CheckExistingInstallation checks for existing MariaDB installations
func (sc *SystemChecker) CheckExistingInstallation() error {
	terminal.PrintInfo("Checking existing installations...")

	if err := sc.CheckExistingServices(); err != nil {
		return err
	}

	if err := sc.CheckExistingPackages(); err != nil {
		return err
	}

	return nil
}

// CheckExistingServices checks for existing MariaDB/MySQL services
func (sc *SystemChecker) CheckExistingServices() error {
	lg, _ := logger.Get()

	sc.systemInfo.ExistingService = sc.svcManager.IsActive("mariadb") || sc.svcManager.IsActive("mysql")

	if sc.systemInfo.ExistingService {
		lg.Warn("Existing MariaDB/MySQL service detected; aborting installation")
		terminal.PrintError("Existing MariaDB/MySQL service detected")
		return fmt.Errorf("existing MariaDB/MySQL service detected")
	}

	return nil
}

// CheckExistingPackages checks for existing MariaDB/MySQL packages
func (sc *SystemChecker) CheckExistingPackages() error {
	lg, _ := logger.Get()

	packages, err := sc.pkgManager.GetInstalledPackages()
	if err != nil {
		lg.Warn("Failed to get installed packages", logger.Error(err))
		return nil // Don't fail installation for this
	}

	sc.systemInfo.ExistingPackages = packages

	if len(sc.systemInfo.ExistingPackages) > 0 {
		terminal.PrintInfo(fmt.Sprintf("Found %d existing MariaDB/MySQL packages", len(sc.systemInfo.ExistingPackages)))
	} else {
		terminal.PrintSuccess("No existing MariaDB/MySQL packages found")
	}

	return nil
}
