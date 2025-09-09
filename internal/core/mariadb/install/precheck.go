package install

import (
	"fmt"
	"os"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/mariadb"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// preInstallationChecks melakukan pemeriksaan sebelum instalasi
func preInstallationChecks(cfg *mariadb.MariaDBInstallConfig, deps *Dependencies) error {
	lg, _ := logger.Get()

	terminal.SafePrintln("🔍 Melakukan pemeriksaan sistem...")

	// Cek OS yang didukung
	if err := system.ValidateOperatingSystem(); err != nil {
		return fmt.Errorf("sistem operasi tidak didukung: %w", err)
	}

	// Cek apakah MariaDB/MySQL sudah terinstall
	if isMariaDBInstalled(deps) {
		installedVersion := getInstalledMariaDBVersion(deps)
		if installedVersion != "" {
			terminal.SafePrintln("⚠️  MariaDB sudah terinstall di sistem")
			terminal.SafePrintln("   Versi terinstall: " + installedVersion)
			terminal.SafePrintln("   Versi yang diminta: " + cfg.Version)

			// Jika versi sama, beri pesan khusus
			if installedVersion == cfg.Version {
				terminal.SafePrintln("   Status: Versi yang diminta sudah terinstall")
			} else {
				terminal.SafePrintln("   Status: Versi berbeda terdeteksi")
			}

			lg.Info("MariaDB sudah terinstall", logger.String("installed_version", installedVersion))
			return fmt.Errorf("MariaDB sudah terinstall (versi: %s). Hapus instalasi existing terlebih dahulu jika ingin menginstall ulang", installedVersion)
		}
	}

	// Cek hak akses root
	if !isRunningAsRoot() {
		return fmt.Errorf("instalasi MariaDB memerlukan hak akses root. Jalankan dengan sudo")
	}

	lg.Info("Pre-installation checks berhasil")
	return nil
}

// validateFinalConfig memvalidasi konfigurasi akhir sebelum instalasi
func validateFinalConfig(cfg *mariadb.MariaDBInstallConfig) error {
	lg, _ := logger.Get()

	// Versi harus sudah ditentukan pada tahap ini
	if cfg.Version == "" {
		return fmt.Errorf("versi MariaDB tidak ditentukan")
	}

	// Tampilkan informasi versi yang akan diinstall
	terminal.SafePrintln(fmt.Sprintf("📋 Versi MariaDB yang akan diinstall: %s", cfg.Version))

	lg.Info("Konfigurasi instalasi valid", logger.String("version", cfg.Version))
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
