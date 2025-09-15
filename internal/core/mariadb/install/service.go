package install

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/mariadb"
	"sfDBTools/utils/terminal"
)

// startMariaDBService memulai dan mengaktifkan service MariaDB
func startMariaDBService(deps *Dependencies) error {
	lg, _ := logger.Get()

	serviceName := "mariadb"

	// Start service
	if err := deps.ServiceManager.Start(serviceName); err != nil {
		return fmt.Errorf("gagal memulai service MariaDB: %w", err)
	}

	// Enable service untuk auto-start
	if err := deps.ServiceManager.Enable(serviceName); err != nil {
		return fmt.Errorf("gagal mengaktifkan auto-start MariaDB: %w", err)
	}

	lg.Info("Service MariaDB berhasil dimulai dan diaktifkan")
	return nil
}

// verifyInstallation memverifikasi bahwa instalasi berhasil
func verifyInstallation(cfg *mariadb.MariaDBInstallConfig, deps *Dependencies) error {
	lg, _ := logger.Get()

	// Cek apakah service berjalan
	if !deps.ServiceManager.IsActive("mariadb") {
		return fmt.Errorf("service MariaDB tidak berjalan")
	}

	// Cek versi yang terinstall
	installedVersion := getInstalledMariaDBVersion(deps)
	if installedVersion == "" {
		return fmt.Errorf("tidak dapat mendeteksi versi MariaDB yang terinstall")
	}

	lg.Info("Verifikasi instalasi berhasil",
		logger.String("installed_version", installedVersion),
		logger.String("requested_version", cfg.Version))

	return nil
}

// displaySuccessMessage menampilkan pesan sukses dan instruksi
func displaySuccessMessage(cfg *mariadb.MariaDBInstallConfig) {
	terminal.SafePrintln("\n🎉 Instalasi MariaDB berhasil!")
	terminal.SafePrintln(fmt.Sprintf("   Versi: %s", cfg.Version))
	terminal.SafePrintln("\n📝 Langkah selanjutnya:")
	terminal.SafePrintln("   1. Jalankan secure installation (opsional):")
	terminal.SafePrintln("      sudo mysql_secure_installation")
	terminal.SafePrintln("   2. Login ke MariaDB:")
	terminal.SafePrintln("      sudo mysql -u root")
	terminal.SafePrintln("   3. Cek status service:")
	terminal.SafePrintln("      systemctl status mariadb")
	terminal.SafePrintln("")
}
