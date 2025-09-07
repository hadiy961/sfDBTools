package remove

import (
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// PackageManager handles package removal operations
type PackageManager struct {
	pkgManager system.PackageManager
}

// NewPackageManager creates a new package manager for removal operations
func NewPackageManager() *PackageManager {
	return &PackageManager{
		pkgManager: system.NewPackageManager(),
	}
}

// RemoveMariaDBPackages removes MariaDB-related packages from the system
func (pm *PackageManager) RemoveMariaDBPackages() error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Removing MariaDB packages...")
	packages := pm.getPackagesToRemove()

	if len(packages) > 0 {
		if err := pm.pkgManager.Remove(packages); err != nil {
			lg.Warn("Failed to remove packages", logger.Error(err))
			terminal.PrintWarning("⚠️  Some packages could not be removed, continuing with cleanup...")
			return err
		} else {
			terminal.PrintSuccess("Package removal completed")
		}
	}
	return nil
}

// getPackagesToRemove determines which packages to remove based on the OS
func (pm *PackageManager) getPackagesToRemove() []string {
	// Use OS detector to determine package type
	osDetector := common.NewOSDetector()
	osInfo, err := osDetector.DetectOS()
	if err != nil {
		// Fallback to generic names
		return []string{"mariadb-server", "mariadb-client", "mariadb"}
	}

	switch osInfo.PackageType {
	case "deb":
		return []string{"^mariadb.*", "^mysql.*"}
	case "rpm":
		return []string{"mariadb-server", "mariadb-client", "mariadb"}
	default:
		return []string{"mariadb-server", "mariadb-client", "mariadb"}
	}
}
