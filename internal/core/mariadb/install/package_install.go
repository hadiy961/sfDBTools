package install

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/mariadb"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// updatePackageCache mengupdate cache package manager
func updatePackageCache(deps *Dependencies) error {
	lg, _ := logger.Get()

	terminal.SafePrintln("🔄 Mengupdate cache package manager...")

	if err := deps.PackageManager.UpdateCache(); err != nil {
		return fmt.Errorf("gagal mengupdate cache package manager: %w", err)
	}

	lg.Info("Cache package manager berhasil diupdate")
	return nil
}

// installMariaDBPackages menginstall paket MariaDB server dan client
func installMariaDBPackages(cfg *mariadb.MariaDBInstallConfig, deps *Dependencies) error {
	lg, _ := logger.Get()

	terminal.SafePrintln("📥 Menginstall paket MariaDB...")

	// Tentukan nama paket berdasarkan OS
	packages, err := getMariaDBPackageNames()
	if err != nil {
		return fmt.Errorf("gagal menentukan nama paket MariaDB: %w", err)
	}

	lg.Info("Menginstall paket MariaDB", logger.Strings("packages", packages))

	if err := deps.PackageManager.Install(packages); err != nil {
		return fmt.Errorf("gagal menginstall paket MariaDB: %w", err)
	}

	lg.Info("Paket MariaDB berhasil diinstall")
	return nil
}

// getMariaDBPackageNames mengembalikan nama paket yang sesuai untuk OS
func getMariaDBPackageNames() ([]string, error) {
	osInfo, err := system.DetectOS()
	if err != nil {
		return nil, fmt.Errorf("gagal deteksi OS: %w", err)
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
