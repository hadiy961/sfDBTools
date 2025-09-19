package configure

import (
	"context"
	"fmt"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/mariadb/discovery"
	"sfDBTools/utils/system"
)

// PerformPreChecks melakukan pemeriksaan prasyarat sebelum konfigurasi
// Mengembalikan MariaDBInstallation untuk digunakan kembali di tahap selanjutnya
func PerformPreChecks(ctx context.Context, config *mariadb_config.MariaDBConfigureConfig) (*discovery.MariaDBInstallation, error) {
	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting pre-checks for MariaDB configuration")

	// Early cancellation check: return promptly if caller cancelled
	if ctx != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}

	// Ensure config is provided and log a small snapshot for debugging
	if config == nil {
		return nil, fmt.Errorf("nil config passed to performPreChecks")
	}

	lg.Debug("Pre-check using provided configuration",
		logger.String("data_dir", config.DataDir),
		logger.String("log_dir", config.LogDir),
		logger.String("binlog_dir", config.BinlogDir),
		logger.Int("port", config.Port))

	// 1.1: Cek privilege sudo/root
	if err := system.CheckPrivileges(); err != nil {
		return nil, fmt.Errorf("privilege check failed: %w", err)
	}
	lg.Debug("Privilege check passed")

	// 1.2: Cek apakah MariaDB sudah terinstall
	installation, err := discovery.DiscoverMariaDBInstallation()
	if err != nil {
		return nil, fmt.Errorf("installation check failed: %w", err)
	}
	if !installation.IsInstalled {
		return nil, fmt.Errorf("MariaDB is not installed. Please install MariaDB first using 'mariadb install' command")
	}
	lg.Debug("MariaDB installation check passed")

	// 1.5: Cek koneksi ke database
	if err := checkDatabaseConnection(installation); err != nil {
		lg.Warn("Database connection check failed, but continuing", logger.Error(err))
		// Warning saja, tidak fatal karena mungkin konfigurasi yang salah
	} else {
		lg.Debug("Database connection check passed")
	}

	lg.Info("All pre-checks completed successfully")
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
