package remove

import (
	"fmt"
	"os"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/terminal"
)

// cleanupSystem melakukan cleanup sistem dan user jika diminta
func cleanupSystem(cfg *mariadb_config.MariaDBRemoveConfig, deps *Dependencies) error {
	lg, _ := logger.Get()

	if cfg.RemoveUser {
		if err := removeMySQLUser(deps); err != nil {
			return fmt.Errorf("gagal menghapus user mysql: %w", err)
		}
		lg.Info("User mysql berhasil dihapus")
	}

	// Cleanup log files
	if err := cleanupLogFiles(); err != nil {
		lg.Warn("Gagal cleanup log files", logger.Error(err))
		// Tidak return error karena tidak critical
	}

	// Cleanup tmp files
	if err := cleanupTempFiles(); err != nil {
		lg.Warn("Gagal cleanup temp files", logger.Error(err))
		// Tidak return error karena tidak critical
	}

	lg.Info("System cleanup selesai")
	return nil
}

// removeMySQLUser menghapus user mysql dari sistem
func removeMySQLUser(deps *Dependencies) error {
	terminal.PrintSubHeader("👤 Menghapus user mysql dari sistem...")

	// Cek apakah user mysql ada
	_, err := deps.ProcessManager.ExecuteWithOutput("id", []string{"mysql"})
	if err != nil {
		info("ℹ User mysql tidak ditemukan")
		return nil
	}

	// Hapus user mysql
	if err := deps.ProcessManager.Execute("userdel", []string{"-r", "mysql"}); err != nil {
		// Coba tanpa -r jika gagal
		if err := deps.ProcessManager.Execute("userdel", []string{"mysql"}); err != nil {
			return fmt.Errorf("gagal menghapus user mysql: %w", err)
		}
	}

	success("User mysql berhasil dihapus")
	return nil
}

// cleanupLogFiles menghapus file-file log MariaDB
func cleanupLogFiles() error {
	terminal.PrintSubHeader("🧹 Membersihkan log files...")

	logPaths := []string{
		"/var/log/mysql",
		"/var/log/mariadb",
		"/var/log/mysqld.log",
		"/var/log/mysql.log",
		"/var/log/mysql.err",
		"/var/log/mysql/error.log",
		"/var/log/mariadb/mariadb.log",
	}

	for _, path := range logPaths {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				continue // File/directory tidak ada, skip
			}
			continue // Skip error, tidak critical
		}

		info("🗂️  Menghapus log: " + path)
		if err := os.RemoveAll(path); err != nil {
			warn("Gagal menghapus: " + path)
		} else {
			success("Dihapus: " + path)
		}
	}

	return nil
}

// cleanupTempFiles menghapus file-file temporary MariaDB
func cleanupTempFiles() error {
	terminal.PrintSubHeader("🧹 Membersihkan temp files...")

	tempPaths := []string{
		"/tmp/mysql.sock",
		"/tmp/mysqld.sock",
		"/tmp/mariadb.sock",
		"/run/mysqld",
		"/run/mariadb",
		"/var/run/mysqld",
		"/var/run/mariadb",
	}

	for _, path := range tempPaths {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				continue // File/directory tidak ada, skip
			}
			continue // Skip error, tidak critical
		}

		info("🗂️  Menghapus temp: " + path)
		if err := os.RemoveAll(path); err != nil {
			warn("Gagal menghapus: " + path)
		} else {
			success("Dihapus: " + path)
		}
	}

	return nil
}

// verifyRemoval memverifikasi bahwa penghapusan berhasil
func verifyRemoval(deps *Dependencies) error {
	terminal.PrintSubHeader("✅ Memverifikasi penghapusan...")

	// Cek apakah masih ada paket yang terinstall
	packages, err := getMariaDBPackageList()
	if err != nil {
		return fmt.Errorf("gagal get package list: %w", err)
	}

	stillInstalled := []string{}
	for _, pkg := range packages {
		if deps.PackageManager.IsInstalled(pkg) {
			stillInstalled = append(stillInstalled, pkg)
		}
	}

	if len(stillInstalled) > 0 {
		warn("Masih ada paket yang terinstall:")
		for _, pkg := range stillInstalled {
			info("- " + pkg)
		}
		return fmt.Errorf("penghapusan tidak lengkap, masih ada %d paket terinstall", len(stillInstalled))
	}

	// Cek apakah service masih aktif
	if deps.ServiceManager.IsActive("mariadb") {
		return fmt.Errorf("service MariaDB masih aktif")
	}

	// Cek apakah masih ada proses yang berjalan
	if isMariaDBProcessRunning(deps) {
		warn("Masih ada proses MariaDB yang berjalan")
		return fmt.Errorf("masih ada proses MariaDB yang berjalan")
	}

	success("Tidak ada paket MariaDB yang terinstall")
	success("Tidak ada service MariaDB yang aktif")
	success("Tidak ada proses MariaDB yang berjalan")

	return nil
}

// isMariaDBProcessRunning mengecek apakah masih ada proses MariaDB yang berjalan
func isMariaDBProcessRunning(deps *Dependencies) bool {
	processes := []string{"mysqld", "mariadbd", "mysql"}

	for _, process := range processes {
		// Gunakan pgrep untuk cari proses
		_, err := deps.ProcessManager.ExecuteWithOutput("pgrep", []string{"-f", process})
		if err == nil {
			// Proses ditemukan
			return true
		}
	}

	return false
}
