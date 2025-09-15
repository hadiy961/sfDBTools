package install

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// updatePackageCache mengupdate cache package manager
func updatePackageCache(deps *Dependencies) error {
	lg, _ := logger.Get()

	// Show spinner while updating package cache so user sees progress
	spinner := terminal.NewInstallSpinner("Mengupdate cache package manager...")
	spinner.Start()

	lg.Info("[Package Manager] Mengupdate cache")

	if err := deps.PackageManager.UpdateCache(); err != nil {
		spinner.StopWithError("Gagal mengupdate cache package manager")
		return fmt.Errorf("gagal mengupdate cache package manager: %w", err)
	}

	spinner.StopWithSuccess("Cache package manager berhasil diupdate")
	lg.Info("[Package Manager] Cache berhasil diupdate")

	return nil
}

// installMariaDBPackages menginstall paket MariaDB server dan client
func installMariaDBPackages(deps *Dependencies) error {
	lg, _ := logger.Get()
	// Use spinner to provide user feedback while determining and installing packages
	spinner := terminal.NewInstallSpinner("Menentukan dan menginstall paket MariaDB...")
	spinner.Start()

	// Tentukan nama paket berdasarkan OS
	lg.Info("Menentukan nama paket MariaDB yang sesuai untuk OS")
	osInfo, err := system.DetectOS()
	if err != nil {
		spinner.StopWithError("Gagal mendeteksi OS untuk penentuan paket")
		return fmt.Errorf("gagal deteksi OS untuk penentuan paket MariaDB: %w", err)
	}

	packages, err := getMariaDBPackageNames(osInfo)
	if err != nil {
		spinner.StopWithError("Gagal menentukan nama paket MariaDB")
		return fmt.Errorf("gagal menentukan nama paket MariaDB: %w", err)
	}

	lg.Info("Menginstall paket MariaDB", logger.Strings("packages", packages))

	if err := deps.PackageManager.Install(packages); err != nil {
		spinner.StopWithError("Gagal menginstall paket MariaDB")
		return fmt.Errorf("gagal menginstall paket MariaDB: %w", err)
	}

	spinner.StopWithSuccess("Paket MariaDB berhasil diinstall")
	lg.Info("Paket MariaDB berhasil diinstall")
	return nil
}

// getMariaDBPackageNames mengembalikan nama paket yang sesuai untuk OS
func getMariaDBPackageNames(osInfo *system.OSInfo) ([]string, error) {
	if osInfo == nil {
		return nil, fmt.Errorf("osInfo tidak boleh nil")
	}

	switch osInfo.PackageType {
	case "deb":
		// Ubuntu/Debian menggunakan huruf kecil
		return []string{"mariadb-server", "mariadb-client"}, nil
	case "rpm":
		// CentOS/RHEL/Rocky menggunakan huruf kapital dari repo MariaDB
		return []string{"MariaDB-server", "MariaDB-client"}, nil
	default:
		// Default fallback ke nama standar
		return []string{"mariadb-server", "mariadb-client"}, nil
	}
}
