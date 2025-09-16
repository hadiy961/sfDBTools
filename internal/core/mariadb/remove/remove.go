package remove

import (
	"context"
	"fmt"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/system"
)

// Dependencies berisi semua dependency yang dibutuhkan untuk penghapusan
type Dependencies struct {
	PackageManager system.PackageManager
	ProcessManager system.ProcessManager
	ServiceManager system.ServiceManager
}

// RunMariaDBRemove menjalankan proses penghapusan MariaDB secara menyeluruh
func RunMariaDBRemove(ctx context.Context, cfg *mariadb_config.MariaDBRemoveConfig) error {
	lg, _ := logger.Get()
	lg.Info("Memulai penghapusan MariaDB",
		logger.Bool("remove_data", cfg.RemoveData),
		logger.Bool("remove_config", cfg.RemoveConfig),
		logger.Bool("force", cfg.Force))

	// Inisialisasi dependencies
	deps := &Dependencies{
		PackageManager: system.NewPackageManager(),
		ProcessManager: system.NewProcessManager(),
		ServiceManager: system.NewServiceManager(),
	}

	// Langkah 1: Pre-removal checks dan validasi
	if err := preRemovalChecks(cfg, deps); err != nil {
		return fmt.Errorf("pre-removal checks gagal: %w", err)
	}

	// Langkah 2: Konfirmasi penghapusan (jika tidak force mode)
	if err := confirmRemoval(cfg, deps); err != nil {
		return fmt.Errorf("konfirmasi penghapusan gagal: %w", err)
	}

	// Langkah 3: Stop dan disable service MariaDB
	if err := stopMariaDBService(deps); err != nil {
		return fmt.Errorf("stop service MariaDB gagal: %w", err)
	}

	// Langkah 4: Backup data sebelum dihapus (jika diminta)
	if err := handleDataBackup(cfg, deps); err != nil {
		return fmt.Errorf("backup data gagal: %w", err)
	}

	// Langkah 5: Hapus paket MariaDB
	if err := removeMariaDBPackages(deps); err != nil {
		return fmt.Errorf("penghapusan paket MariaDB gagal: %w", err)
	}

	// Langkah 6: Hapus data dan konfigurasi (jika diminta)
	if err := removeDataAndConfig(cfg, deps); err != nil {
		return fmt.Errorf("penghapusan data/config gagal: %w", err)
	}

	// Langkah 7: Hapus repository MariaDB (jika diminta)
	if err := removeMariaDBRepository(cfg, deps); err != nil {
		return fmt.Errorf("penghapusan repository gagal: %w", err)
	}

	// Langkah 8: Cleanup sistem dan user
	if err := cleanupSystem(cfg, deps); err != nil {
		return fmt.Errorf("cleanup sistem gagal: %w", err)
	}

	// Langkah 9: Verifikasi penghapusan
	if err := verifyRemoval(deps); err != nil {
		return fmt.Errorf("verifikasi penghapusan gagal: %w", err)
	}

	// Tampilkan pesan sukses
	displayRemovalSuccess(cfg)

	lg.Info("Penghapusan MariaDB berhasil diselesaikan")
	return nil
}
