package install

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// PackageInstaller handles MariaDB package installation operations
type PackageInstaller struct {
	pkgManager system.PackageManager
	osInfo     *system.OSInfo
}

// NewPackageInstaller creates a new package installer instance
func NewPackageInstaller(pkgManager system.PackageManager, osInfo *system.OSInfo) *PackageInstaller {
	return &PackageInstaller{
		pkgManager: pkgManager,
		osInfo:     osInfo,
	}
}

// InstallPackages installs MariaDB server and client packages
func (pi *PackageInstaller) InstallPackages() (int, error) {
	lg, _ := logger.Get()

	packages := pi.getPackagesToInstall()

	spinner := terminal.NewInstallSpinner(fmt.Sprintf("Installing %d MariaDB packages...", len(packages)))
	spinner.Start()

	spinner.UpdateMessage(fmt.Sprintf("Installing packages: %v", packages))

	if err := pi.pkgManager.Install(packages); err != nil {
		spinner.StopWithError("Package installation failed")
		return 0, fmt.Errorf("package installation failed: %w", err)
	}

	lg.Info("Packages installed successfully", logger.Strings("packages", packages))
	spinner.StopWithSuccess("Package installation completed")

	return len(packages), nil
}

// getPackagesToInstall returns the list of packages to install based on OS
func (pi *PackageInstaller) getPackagesToInstall() []string {
	switch pi.osInfo.PackageType {
	case "deb":
		return []string{"mariadb-server", "mariadb-client"}
	case "rpm":
		return []string{"MariaDB-server", "MariaDB-client"}
	default:
		// Default to generic names
		return []string{"mariadb-server", "mariadb-client"}
	}
}

// GetPackageList returns the list of packages that would be installed (for dry-run)
func (pi *PackageInstaller) GetPackageList() []string {
	return pi.getPackagesToInstall()
}
