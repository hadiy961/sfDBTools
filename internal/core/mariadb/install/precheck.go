package install

import (
	"fmt"
	"os"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/mariadb"
	"sfDBTools/utils/system"
)

// preInstallationChecks melakukan pemeriksaan sebelum instalasi
func preInstallationChecks(cfg *mariadb.MariaDBInstallConfig, deps *Dependencies) error {
	lg, _ := logger.Get()

	// Internal diagnostic only; reduce noise on normal runs
	lg.Debug("Melakukan pemeriksaan sistem...")

	// Cek OS yang didukung
	if err := system.ValidateOperatingSystem(); err != nil {
		return fmt.Errorf("sistem operasi tidak didukung: %w", err)
	}

	// Cek apakah MariaDB/MySQL sudah terinstall
	if isMariaDBInstalled(deps) {
		installedVersion := getInstalledMariaDBVersion(deps)
		if installedVersion != "" {
			// Use debug-level logs for internal state to avoid duplicate console output
			lg.Debug("MariaDB sudah terinstall di sistem",
				logger.String("installed_version", installedVersion),
				logger.String("requested_version", cfg.Version))

			// Jika versi sama, beri pesan khusus (debug)
			if installedVersion == cfg.Version {
				lg.Debug("Status: Versi yang diminta sudah terinstall")
			} else {
				lg.Debug("Status: Versi berbeda terdeteksi")
			}

			return fmt.Errorf("MariaDB sudah terinstall (versi: %s). Hapus instalasi existing terlebih dahulu jika ingin menginstall ulang", installedVersion)
		}
	}

	// Cek hak akses root
	if !isRunningAsRoot() {
		return fmt.Errorf("instalasi MariaDB memerlukan hak akses root. Jalankan dengan sudo")
	}

	return nil
}

// validateFinalConfig memvalidasi konfigurasi akhir sebelum instalasi
func validateFinalConfig(cfg *mariadb.MariaDBInstallConfig) error {
	lg, _ := logger.Get()

	// Versi harus sudah ditentukan pada tahap ini
	if cfg.Version == "" {
		return fmt.Errorf("versi MariaDB tidak ditentukan")
	}

	// Internal info only
	lg.Debug(fmt.Sprintf("Versi MariaDB yang akan diinstall: %s", cfg.Version))

	lg.Debug("Konfigurasi instalasi valid", logger.String("version", cfg.Version))
	return nil
}

// Helper functions untuk pre-check

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
