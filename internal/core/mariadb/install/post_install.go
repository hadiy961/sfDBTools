package install

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
)

// Post-installation setup seperti konfigurasi awal
func postInstallationSetup() error {
	lg, _ := logger.Get()

	lg.Info("Memulai post-installation setup")

	// // Langkah 1 : Jalankan mariadb-secure-installation
	// if err := defaultsetup.RunMariaDBSecureInstallation(); err != nil {
	// 	return fmt.Errorf("gagal menjalankan secure installation: %w", err)
	// }

	// // Langkah 2 : Buat user & grants default (hardcoded)
	// if err := defaultsetup.CreateDefaultMariaDBUser(); err != nil {
	// 	return fmt.Errorf("gagal membuat default users/grants: %w", err)
	// }

	// // Langkah 3 : Buat database default (hardcoded)
	// if err := defaultsetup.CreateDefaultDatabase(deps); err != nil {
	// 	return fmt.Errorf("gagal membuat default database: %w", err)
	// }

	// Langkah 4 : Ubah konfigurasi MariaDB sesuai standart perusahaan
	if err := configureMariaDBConf(); err != nil {
		return fmt.Errorf("gagal mengubah konfigurasi MariaDB: %w", err)
	}

	lg.Info("Post-installation setup selesai")
	return nil
}

// mengubah konfigurasi MariaDB sesuai standart perusahaan
func configureMariaDBConf() error {
	lg, _ := logger.Get()

	lg.Info("Mengubah konfigurasi MariaDB sesuai standart perusahaan")

	// check privileges
	lg.Info("Memeriksa hak akses root/sudo untuk mengubah konfigurasi MariaDB")
	if err := system.CheckPrivileges(); err != nil {
		return fmt.Errorf("privilege check failed: %w", err)
	}
	lg.Info("Privilege check passed")

	return nil
}
