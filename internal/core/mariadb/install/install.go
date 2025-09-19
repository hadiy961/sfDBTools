package install

import (
	"context"
	"fmt"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	defaultsetup "sfDBTools/utils/mariadb/defaultSetup"
	"sfDBTools/utils/system"
)

// RunMariaDBInstall menjalankan proses instalasi MariaDB lengkap
func RunMariaDBInstall(ctx context.Context, cfg *mariadb_config.MariaDBInstallConfig, mariadb_config *mariadb_config.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()
	// lg.Info("Memulai instalasi MariaDB",
	// 	logger.String("version", cfg.Version),
	// 	logger.Bool("non_interactive", cfg.NonInteractive))

	// Inisialisasi dependencies
	deps := &defaultsetup.Dependencies{
		PackageManager: system.NewPackageManager(),
		ProcessManager: system.NewProcessManager(),
		ServiceManager: system.NewServiceManager(),
	}

	// // Langkah 1: Pre-installation checks (termasuk OS dan hak akses)
	// terminal.PrintHeader("Pre-Installation Checks")
	// installation, err := preInstallationChecks(cfg, deps)
	// if err != nil {
	// 	return fmt.Errorf("pre-installation checks gagal: %w", err)
	// }
	// terminal.PrintHeader("MariaDB Installation Process")
	// // Langkah 2: Validasi konfigurasi (tidak ada lagi interactive input)
	// if err := validateFinalConfig(cfg); err != nil {
	// 	return fmt.Errorf("validasi konfigurasi gagal: %w", err)
	// }

	// // Langkah 3: Repository setup (selalu dilakukan)
	// if err := setupMariaDBRepository(ctx, cfg, deps); err != nil {
	// 	return fmt.Errorf("setup repository gagal: %w", err)
	// }

	// // Langkah 4: Update system packages
	// if err := updateSystemPackages(deps); err != nil {
	// 	return fmt.Errorf("upgrade paket sistem gagal: %w", err)
	// }

	// // Langkah 4: Update package cache
	// if err := updatePackageCache(deps); err != nil {
	// 	return fmt.Errorf("update package cache gagal: %w", err)
	// }

	// // Langkah 5: Install MariaDB packages
	// if err := installMariaDBPackages(deps); err != nil {
	// 	return fmt.Errorf("instalasi paket MariaDB gagal: %w", err)
	// }

	// // Langkah 6: Start and enable service
	// if err := startMariaDBService(deps); err != nil {
	// 	return fmt.Errorf("start service MariaDB gagal: %w", err)
	// }

	// // Langkah 7: Verification
	// if err := verifyInstallation(cfg, deps); err != nil {
	// 	return fmt.Errorf("verifikasi instalasi gagal: %w", err)
	// }

	// Langkah 8: Post-installation
	if err := postInstallationSetup(deps, mariadb_config); err != nil {
		return fmt.Errorf("post-installation setup gagal: %w", err)
	}

	// // Tampilkan pesan sukses dan instruksi selanjutnya
	displaySuccessMessage(cfg)

	lg.Info("Instalasi MariaDB berhasil diselesaikan", logger.String("version", cfg.Version))
	return nil
}
