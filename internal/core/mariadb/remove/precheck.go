package remove

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/terminal"
)

// preRemovalChecks melakukan pemeriksaan sebelum penghapusan
func preRemovalChecks(cfg *mariadb_config.MariaDBRemoveConfig, deps *Dependencies) error {
	lg, _ := logger.Get()

	terminal.PrintSubHeader("Melakukan pemeriksaan sistem untuk penghapusan...")

	// Cek hak akses root
	if !isRunningAsRoot() {
		return fmt.Errorf("penghapusan MariaDB memerlukan hak akses root. Jalankan dengan sudo")
	}

	// Cek apakah MariaDB terinstall
	if !isMariaDBInstalled(deps) {
		return fmt.Errorf("MariaDB tidak terdeteksi di sistem. Tidak ada yang perlu dihapus")
	}

	// Deteksi versi yang terinstall untuk informasi
	installedVersion := getInstalledMariaDBVersion(deps)
	if installedVersion != "" {
		listHeader("📋 MariaDB terdeteksi:")
		infof("Versi: %s", installedVersion)
	}

	// Cek status service
	if deps.ServiceManager.IsActive("mariadb") {
		warn("Service MariaDB sedang berjalan")
		info("Status: Akan dihentikan sebagai bagian dari proses penghapusan")
	}

	lg.Info("Pre-removal checks berhasil")
	return nil
}

// confirmRemoval meminta konfirmasi user untuk penghapusan
func confirmRemoval(cfg *mariadb_config.MariaDBRemoveConfig, deps *Dependencies) error {
	// Skip konfirmasi jika force mode atau non-interactive
	if cfg.Force || cfg.NonInteractive {
		return nil
	}

	// Show concise confirmation block
	terminal.PrintSubHeader("⚠️  Konfirmasi Penghapusan ⚠️")
	listHeader("Yang akan dihapus:")
	info("✓ Paket MariaDB server dan client")

	if cfg.RemoveData {
		info("✓ Data directory - SEMUA DATABASE AKAN HILANG!")

		// Deteksi dan tampilkan direktori custom yang akan dihapus
		if mariadbConfig := getDetectedConfig(deps); mariadbConfig != nil {
			customDirs := getAllCustomDirectories(mariadbConfig)
			if len(customDirs) > 0 {
				info("📂 Direktori yang akan dihapus:")

				// Default placeholder directories that we shouldn't show if they don't exist
				defaults := map[string]bool{
					"/var/lib/mysql": true,
					"/var/log/mysql": true,
				}

				for _, dir := range customDirs {
					// Normalize to absolute path for reliable comparisons
					absDir := dir
					if p, err := filepath.Abs(dir); err == nil {
						absDir = p
					}

					// If this dir is a default placeholder and is absent, skip it to reduce noise
					if defaults[absDir] {
						if _, err := os.Stat(absDir); os.IsNotExist(err) {
							// skip showing default placeholder that doesn't exist
							continue
						}
					}

					// Show directory; if missing, annotate
					if _, err := os.Stat(absDir); err == nil {
						info("- " + absDir)
					} else {
						info("- " + absDir + " (tidak ditemukan)")
					}
				}
			}
		} else {
			info("📂 Direktori default: /var/lib/mysql")
		}
	}

	if cfg.RemoveConfig {
		info("✓ File konfigurasi (/etc/mysql, /etc/my.cnf)")
	}

	if cfg.RemoveRepository {
		info("✓ Repository MariaDB")
	}

	if cfg.RemoveUser {
		info("✓ User sistem 'mysql'")
	}

	if cfg.BackupData {
		infof("Backup data akan dibuat di: %s", cfg.BackupPath)
	}

	warn("PERHATIAN: Proses ini TIDAK DAPAT DIBATALKAN. Ketik 'HAPUS' untuk melanjutkan.")

	fmt.Print("\nKonfirmasi: ")

	var response string
	fmt.Scanln(&response)
	response = strings.TrimSpace(response)

	if response != "HAPUS" {
		return fmt.Errorf("penghapusan dibatalkan oleh user")
	}

	return nil
}

// Helper functions

func isMariaDBInstalled(deps *Dependencies) bool {
	// Cek berbagai kemungkinan nama paket MariaDB
	packages := []string{
		"mariadb-server", "MariaDB-server", // MariaDB packages
		"mysql-server", // MySQL packages
	}

	for _, pkg := range packages {
		if deps.PackageManager.IsInstalled(pkg) {
			return true
		}
	}

	return false
}

func getInstalledMariaDBVersion(deps *Dependencies) string {
	// Coba jalankan mysql --version untuk mendapatkan versi
	output, err := deps.ProcessManager.ExecuteWithOutput("mysql", []string{"--version"})
	if err != nil {
		return ""
	}

	// Parse output untuk mendapatkan versi
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		// Contoh output: mysql  Ver 15.1 Distrib 10.6.23-MariaDB, for Linux (x86_64)
		if strings.Contains(lines[0], "MariaDB") {
			parts := strings.Split(lines[0], " ")
			for _, part := range parts {
				if strings.Contains(part, "MariaDB") {
					// Ekstrak versi dari format "10.6.23-MariaDB,"
					version := strings.Split(part, "-")[0]
					return version
				}
			}
		}
	}

	return ""
}

func isRunningAsRoot() bool {
	return os.Geteuid() == 0
}

func checkDataDirectory() error {
	dataDir := "/var/lib/mysql"

	// Cek apakah direktori ada
	if _, err := os.Stat(dataDir); err != nil {
		if os.IsNotExist(err) {
			// Direktori tidak ada, tidak masalah
			return nil
		}
		return fmt.Errorf("tidak dapat mengakses direktori data %s: %w", dataDir, err)
	}

	// Cek ukuran direktori untuk peringatan
	// Ini adalah implementasi sederhana, bisa diperluas
	warn("Direktori data ditemukan: " + dataDir)

	return nil
}

// func checkDataDirectories() error {
// 	// Deteksi direktori custom dari konfigurasi
// 	mariadbConfig, err := detectCustomDirectories()
// 	if err != nil {
// 		warn("Gagal deteksi direktori custom, akan cek direktori default")
// 		return checkDataDirectory()
// 	}

// 	// Cek semua direktori yang terdeteksi
// 	customDirs := getAllCustomDirectories(mariadbConfig)

// 	foundDirs := 0

// 	for _, dir := range customDirs {
// 		if _, err := os.Stat(dir); err == nil {
// 			success("" + dir)
// 			foundDirs++
// 		}
// 	}

// 	return nil
// }
