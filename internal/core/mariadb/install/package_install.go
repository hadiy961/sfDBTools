package install

import (
	"fmt"

	"sfDBTools/internal/logger"
	defaultsetup "sfDBTools/utils/mariadb/defaultSetup"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// updatePackageCache mengupdate cache package manager
func updatePackageCache(deps *defaultsetup.Dependencies) error {
	lg, _ := logger.Get()

	terminal.PrintSubHeader("[Package Manager] Update Cache")

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

// updateSystemPackages menjalankan upgrade/update paket sistem (mis. apt upgrade / yum update)
// dan menampilkan spinner selama proses berlangsung.
func updateSystemPackages(deps *defaultsetup.Dependencies) error {
	lg, _ := logger.Get()

	terminal.PrintSubHeader("[Package Manager] Upgrade Paket Sistem")
	// Show spinner while upgrading packages so user sees progress
	spinner := terminal.NewInstallSpinner("Mengupgrade paket sistem...")
	spinner.Start()

	lg.Info("[Package Manager] Mengupgrade paket sistem")

	if err := deps.PackageManager.Upgrade(); err != nil {
		spinner.StopWithError("Gagal mengupgrade paket sistem")
		lg.Error("upgrade paket sistem gagal", logger.Error(err))
		return fmt.Errorf("gagal mengupgrade paket sistem: %w", err)
	}

	spinner.StopWithSuccess("Paket sistem berhasil diupgrade")
	lg.Info("[Package Manager] Paket sistem berhasil diupgrade")
	return nil
}

// installMariaDBPackages menginstall paket MariaDB server dan client satu per satu dengan progress
func installMariaDBPackages(deps *defaultsetup.Dependencies) error {
	lg, _ := logger.Get()
	terminal.PrintSubHeader("[Package Manager] Install Paket MariaDB")
	spinner := terminal.NewInstallSpinner("Menentukan dan menginstall paket MariaDB...")
	spinner.Start()

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

	spinner.StopWithSuccess("Daftar paket MariaDB berhasil didapatkan")

	total := len(packages)
	for i, pkg := range packages {
		// Use spinner per package to avoid flooding logs with many lines
		pkgSpinner := terminal.NewInstallSpinner(fmt.Sprintf("[%d/%d] Menginstall paket: %s", i+1, total, pkg))
		pkgSpinner.Start()

		lg.Info("installing package", logger.Int("index", i+1), logger.Int("total", total), logger.String("package", pkg))

		// Install satu package
		if err := deps.PackageManager.Install([]string{pkg}); err != nil {
			pkgSpinner.StopWithError("Gagal menginstall paket")
			errorMsg := fmt.Sprintf("Gagal menginstall paket %s", pkg)
			// Log error with stack-like message but avoid printing full output to stdout
			lg.Error(errorMsg, logger.Error(err))
			return fmt.Errorf("gagal menginstall paket %s: %w", pkg, err)
		}

		pkgSpinner.StopWithSuccess("Berhasil")
		lg.Info("package installed", logger.String("package", pkg))
	}

	fmt.Println("Semua paket MariaDB berhasil diinstall")
	lg.Info("Semua paket MariaDB berhasil diinstall")
	return nil
}

// getMariaDBPackageNames mengembalikan nama paket yang sesuai untuk OS
func getMariaDBPackageNames(osInfo *system.OSInfo) ([]string, error) {
	if osInfo == nil {
		return nil, fmt.Errorf("osInfo tidak boleh nil")
	}

	var packages []string

	switch osInfo.PackageType {
	case "deb":
		// Ubuntu/Debian packages
		packages = []string{
			// Core MariaDB
			"mariadb-server",
			"mariadb-client",
			"mariadb-backup",

			// Performance & Monitoring
			"mytop",
			"mariadb-plugin-connect",
			"mariadb-plugin-spider",
			"mariadb-plugin-oqgraph",

			// Development & Utilities
			"libmariadb-dev",
			"mariadb-test",
			"percona-toolkit",

			// Security & SSL
			"ssl-cert",

			// System utilities
			"htop",
			"iotop",
			"sysstat",
		}

	case "rpm":
		// CentOS/RHEL/Rocky packages
		packages = []string{
			// EPEL repository (harus diinstall pertama)
			"epel-release",

			// Core MariaDB
			"MariaDB-server",
			"MariaDB-client",
			"MariaDB-backup",
			"MariaDB-common",
			"MariaDB-shared",

			// Performance & Monitoring
			"mytop",
			"nmon",

			// System utilities & monitoring
			"htop",
			"iotop",
			"sysstat",
			"rsync",
			"lsof",
			"strace",

			// Compression utilities for backups
			"pigz",
			"pv",
		}

	default:
		// Default fallback
		packages = []string{
			"mariadb-server",
			"mariadb-client",
			"MariaBackup",
			"mariadb-common",
			"mariadb-shared",
			"mytop",
			"nmon",
			"htop",
			"sysstat",
		}
	}

	return packages, nil
}
