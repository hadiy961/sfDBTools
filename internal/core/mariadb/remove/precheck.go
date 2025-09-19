package remove

import (
	"fmt"
	"os"
	"strings"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/terminal"
)

// preRemovalChecks melakukan pemeriksaan sebelum penghapusan
func preRemovalChecks(cfg *mariadb_config.MariaDBRemoveConfig, deps *Dependencies) error {
	lg, _ := logger.Get()

	terminal.SafePrintln("🔍 Melakukan pemeriksaan sistem untuk penghapusan...")

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
		terminal.SafePrintln("📋 MariaDB terdeteksi:")
		terminal.SafePrintln("   Versi: " + installedVersion)
	}

	// Cek status service
	if deps.ServiceManager.IsActive("mariadb") {
		terminal.SafePrintln("⚠️  Service MariaDB sedang berjalan")
		terminal.SafePrintln("   Status: Akan dihentikan sebagai bagian dari proses penghapusan")
	}

	// Cek direktori data jika akan dihapus
	if cfg.RemoveData {
		if err := checkDataDirectories(); err != nil {
			return fmt.Errorf("pemeriksaan direktori data gagal: %w", err)
		}
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

	terminal.SafePrintln("\n⚠️  PERINGATAN: Penghapusan MariaDB")
	terminal.SafePrintln("=====================================")

	// Tampilkan apa yang akan dihapus
	terminal.SafePrintln("Yang akan dihapus:")
	terminal.SafePrintln("✓ Paket MariaDB server dan client")

	if cfg.RemoveData {
		terminal.SafePrintln("✓ Data directory - SEMUA DATABASE AKAN HILANG!")

		// Deteksi dan tampilkan direktori custom yang akan dihapus
		if mariadbConfig, err := detectCustomDirectories(); err == nil {
			customDirs := getAllCustomDirectories(mariadbConfig)
			if len(customDirs) > 0 {
				terminal.SafePrintln("   📂 Direktori yang akan dihapus:")
				for _, dir := range customDirs {
					if _, err := os.Stat(dir); err == nil {
						terminal.SafePrintln("      - " + dir)
					}
				}
			}
		} else {
			terminal.SafePrintln("   📂 Direktori default: /var/lib/mysql")
		}
	}

	if cfg.RemoveConfig {
		terminal.SafePrintln("✓ File konfigurasi (/etc/mysql, /etc/my.cnf)")
	}

	if cfg.RemoveRepository {
		terminal.SafePrintln("✓ Repository MariaDB")
	}

	if cfg.RemoveUser {
		terminal.SafePrintln("✓ User sistem 'mysql'")
	}

	if cfg.BackupData {
		terminal.SafePrintln("\n📋 Backup data akan dibuat di: " + cfg.BackupPath)
	}

	terminal.SafePrintln("\n💥 PERHATIAN:")
	terminal.SafePrintln("   - Proses ini TIDAK DAPAT DIBATALKAN")
	terminal.SafePrintln("   - Pastikan Anda sudah backup data penting")
	terminal.SafePrintln("   - Semua koneksi database akan terputus")

	fmt.Print("\nApakah Anda yakin ingin melanjutkan? Ketik 'HAPUS' untuk konfirmasi: ")

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
	terminal.SafePrintln("⚠️  Direktori data ditemukan: " + dataDir)

	return nil
}

func checkDataDirectories() error {
	// Deteksi direktori custom dari konfigurasi
	mariadbConfig, err := detectCustomDirectories()
	if err != nil {
		terminal.SafePrintln("⚠️  Gagal deteksi direktori custom, akan cek direktori default")
		return checkDataDirectory()
	}

	// Cek semua direktori yang terdeteksi
	customDirs := getAllCustomDirectories(mariadbConfig)

	terminal.SafePrintln("📂 Direktori yang akan dihapus:")
	foundDirs := 0

	for _, dir := range customDirs {
		if _, err := os.Stat(dir); err == nil {
			terminal.SafePrintln("   ✓ " + dir)
			foundDirs++
		}
	}

	if foundDirs == 0 {
		terminal.SafePrintln("   ℹ Tidak ada direktori data yang ditemukan")
	} else {
		terminal.SafePrintln(fmt.Sprintf("   📊 Total: %d direktori akan dihapus", foundDirs))
	}

	return nil
}
