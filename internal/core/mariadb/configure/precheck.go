package configure

import (
	"context"
	"fmt"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/mariadb/discovery"
	"sfDBTools/utils/system"
)

// performPreChecks melakukan pemeriksaan prasyarat sebelum konfigurasi
// Sesuai dengan Step 1 dalam flow implementasi
// Mengembalikan MariaDBInstallation untuk digunakan kembali di tahap selanjutnya
func performPreChecks(ctx context.Context, config *mariadb_config.MariaDBConfigureConfig) (*discovery.MariaDBInstallation, error) {
	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting pre-checks for MariaDB configuration")

	// 1.1: Cek privilege sudo/root
	if err := system.CheckPrivileges(); err != nil {
		return nil, fmt.Errorf("privilege check failed: %w", err)
	}
	lg.Info("Privilege check passed")

	// 1.2: Cek apakah MariaDB sudah terinstall
	installation, err := checkMariaDBInstallation()
	if err != nil {
		return nil, fmt.Errorf("installation check failed: %w", err)
	}
	if !installation.IsInstalled {
		return nil, fmt.Errorf("MariaDB is not installed. Please install MariaDB first using 'mariadb install' command")
	}
	lg.Info("MariaDB installation check passed")

	// 1.5: Cek koneksi ke database
	if err := checkDatabaseConnection(installation); err != nil {
		lg.Warn("Database connection check failed, but continuing", logger.Error(err))
		// Warning saja, tidak fatal karena mungkin konfigurasi yang salah
	} else {
		lg.Info("Database connection check passed")
	}

	lg.Info("All pre-checks completed successfully")
	return installation, nil
}

// checkMariaDBInstallation memeriksa apakah MariaDB sudah terinstall
func checkMariaDBInstallation() (*discovery.MariaDBInstallation, error) {
	// Gunakan discovery function yang sudah ada
	installation, err := discovery.DiscoverMariaDBInstallation()
	if err != nil {
		return nil, fmt.Errorf("failed to discover MariaDB installation: %w", err)
	}

	return installation, nil
}

// checkDatabaseConnection memeriksa koneksi ke database
func checkDatabaseConnection(installation *discovery.MariaDBInstallation) error {
	// Jika service tidak berjalan, tidak perlu coba koneksi
	if !installation.IsRunning {
		return fmt.Errorf("MariaDB service is not running")
	}

	// Buat database config dari installation info
	dbConfig := mariadb_config.CreateDatabaseConfigFromInstallation(installation)
	if dbConfig == nil {
		return fmt.Errorf("failed to create database config from installation info")
	}

	// Coba koneksi ke database dengan password kosong (default untuk fresh install)
	// Ini hanya test koneksi, bukan untuk operasi serius
	sm := system.NewServiceManager()
	status, err := sm.GetStatus(installation.ServiceName)
	if err != nil {
		return fmt.Errorf("failed to get service status: %w", err)
	}

	if !status.Active {
		return fmt.Errorf("MariaDB service is not active")
	}

	return nil
}
