package install

import (
	"fmt"

	"sfDBTools/internal/logger"
	defaultsetup "sfDBTools/utils/mariadb/defaultSetup"
	"sfDBTools/utils/mariadb/discovery"
)

// Post-installation setup seperti konfigurasi awal
func postInstallationSetup(deps *defaultsetup.Dependencies, mariadbInstallation *discovery.MariaDBInstallation) error {
	lg, _ := logger.Get()

	lg.Info("Memulai post-installation setup")

	// // Langkah 1 : Jalankan mariadb-secure-installation
	if err := defaultsetup.RunMariaDBSecureInstallation(deps); err != nil {
		return fmt.Errorf("gagal menjalankan secure installation: %w", err)
	}

	// Langkah 2 : Buat user & grants default (hardcoded)
	if err := defaultsetup.CreateDefaultMariaDBUser(); err != nil {
		return fmt.Errorf("gagal membuat default users/grants: %w", err)
	}

	// Langkah 3 : Buat database default (hardcoded)
	if err := defaultsetup.CreateDefaultDatabase(); err != nil {
		return fmt.Errorf("gagal membuat default database: %w", err)
	}

	lg.Info("Post-installation setup selesai")
	return nil
}
