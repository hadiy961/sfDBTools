package install

import (
	"context"
	"fmt"
	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	defaultsetup "sfDBTools/utils/mariadb/defaultSetup"
	"sfDBTools/utils/mariadb/discovery"
	"sfDBTools/utils/terminal"
)

// Post-installation setup seperti konfigurasi awal
func postInstallationSetup(deps *defaultsetup.Dependencies, mariadb_config *mariadb_config.MariaDBConfigureConfig, installation *discovery.MariaDBInstallation) error {
	lg, _ := logger.Get()
	terminal.Clear()
	lg.Info("Memulai post-installation setup")

	// Langkah 1 : Jalankan mariadb-secure-installation
	// terminal.PrintHeader("MariaDB Secure Installation Process")
	// if err := defaultsetup.RunMariaDBSecureInstallation(deps); err != nil {
	// 	return fmt.Errorf("gagal menjalankan secure installation: %w", err)
	// }

	// Langkah 3 : Buat database default (hardcoded)
	terminal.PrintHeader("Creating Default Database")
	if err := defaultsetup.CreateDefaultDatabase(); err != nil {
		return fmt.Errorf("gagal membuat default database: %w", err)
	}

	// Langkah 2 : Buat user & grants default (hardcoded)
	terminal.PrintHeader("Creating Default Users and Grants")
	if err := defaultsetup.CreateDefaultMariaDBUser(); err != nil {
		return fmt.Errorf("gagal membuat default users/grants: %w", err)
	}

	// Langkah 4 : Konfigurasi standart perusahaan
	ctx := context.Background()
	if err := defaultsetup.RunStandardConfiguration(ctx, mariadb_config, installation); err != nil {
		return fmt.Errorf("gagal menjalankan konfigurasi standart perusahaan: %w", err)
	}

	// Selesai
	lg.Info("Post-installation setup selesai")
	return nil
}
