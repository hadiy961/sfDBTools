package install

import (
	"fmt"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	defaultsetup "sfDBTools/utils/mariadb/defaultSetup"
)

// startMariaDBService memulai dan mengaktifkan service MariaDB
func startMariaDBService(deps *defaultsetup.Dependencies) error {
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
func verifyInstallation(cfg *mariadb_config.MariaDBInstallConfig, deps *defaultsetup.Dependencies) error {
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
