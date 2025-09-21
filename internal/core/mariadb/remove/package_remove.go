package remove

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// removeMariaDBPackages menghapus semua paket MariaDB dari sistem
func removeMariaDBPackages(deps *Dependencies) error {
	lg, _ := logger.Get()

	terminal.PrintSubHeader("🗑️  Menghapus paket MariaDB...")

	// Tentukan nama paket berdasarkan OS
	packages, err := getMariaDBPackageList()
	if err != nil {
		return fmt.Errorf("gagal menentukan daftar paket MariaDB: %w", err)
	}

	// Filter hanya paket yang terinstall
	installedPackages := []string{}
	for _, pkg := range packages {
		if deps.PackageManager.IsInstalled(pkg) {
			installedPackages = append(installedPackages, pkg)
			info("📦 Ditemukan: " + pkg)
		}
	}

	if len(installedPackages) == 0 {
		info("ℹ Tidak ada paket MariaDB yang terinstall")
		return nil
	}

	fmt.Println()
	lg.Info("Menghapus paket MariaDB", logger.Strings("packages", installedPackages))

	// Hapus paket
	if err := deps.PackageManager.Remove(installedPackages); err != nil {
		return fmt.Errorf("gagal menghapus paket MariaDB: %w", err)
	}

	success("Paket MariaDB berhasil dihapus")

	// Purge paket untuk menghapus konfigurasi juga (khusus Debian/Ubuntu)
	if err := purgePackagesIfSupported(deps, installedPackages); err != nil {
		lg.Warn("Gagal purge paket", logger.Error(err))
		// Tidak return error karena paket sudah dihapus
	}

	return nil
}

// getMariaDBPackageList mengembalikan daftar lengkap paket MariaDB yang mungkin terinstall
func getMariaDBPackageList() ([]string, error) {
	osInfo, err := system.DetectOS()
	if err != nil {
		return nil, fmt.Errorf("gagal deteksi OS: %w", err)
	}

	var packages []string

	switch osInfo.PackageType {
	case "deb":
		// Ubuntu/Debian packages
		packages = []string{
			"mariadb-server",
			"mariadb-client",
			"mariadb-common",
			"mariadb-server-core",
			"mariadb-client-core",
			"mysql-common",
			"libmariadb3",
			"libmysqlclient21",
		}
	case "rpm":
		// CentOS/RHEL/Rocky packages
		packages = []string{
			"MariaDB-server",
			"MariaDB-client",
			"MariaDB-common",
			"MariaDB-shared",
			"mariadb-server",
			"mariadb-client",
			"mysql-server",
			"mysql-client",
		}
	default:
		// Default fallback
		packages = []string{
			"mariadb-server",
			"mariadb-client",
			"mysql-server",
			"mysql-client",
		}
	}

	return packages, nil
}

// purgePackagesIfSupported melakukan purge paket jika OS mendukung (Debian/Ubuntu)
func purgePackagesIfSupported(deps *Dependencies, packages []string) error {
	osInfo, err := system.DetectOS()
	if err != nil {
		return err
	}

	// Hanya lakukan purge di Debian/Ubuntu
	if osInfo.PackageType == "deb" {
		info("🧹 Melakukan purge konfigurasi paket...")

		// Gunakan apt-get purge untuk menghapus konfigurasi
		args := append([]string{"purge", "-y"}, packages...)
		if err := deps.ProcessManager.Execute("apt-get", args); err != nil {
			return fmt.Errorf("gagal purge paket: %w", err)
		}

		success("Konfigurasi paket berhasil dihapus")
	}

	return nil
}
