package mariadb_cmd

import (
	"context"

	"sfDBTools/internal/core/mariadb/install"
	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

// InstallCmd installs MariaDB server dengan versi dari flag atau config
var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install MariaDB server dengan versi dari flag atau konfigurasi",
	Long: `Install MariaDB server dengan versi yang ditentukan melalui flag --version
atau menggunakan versi default dari file konfigurasi.

Command ini akan:
1. Melakukan pemeriksaan sistem dan dependency
2. Setup repository resmi MariaDB
3. Menginstall paket MariaDB server dan client
4. Memulai dan mengaktifkan service MariaDB
5. Memverifikasi instalasi

Prioritas versi:
1. Flag --version (tertinggi)
2. Environment variable SFDBTOOLS_MARIADB_VERSION
3. Default dari file config /etc/sfDBTools/config/config.yaml
4. Hardcoded default: 10.6.23 (terendah)

Instalasi memerlukan hak akses root (sudo).

Contoh penggunaan:
  # Instalasi MariaDB dengan versi dari config file
  sudo sfdbtools mariadb install

  # Instalasi MariaDB versi spesifik
  sudo sfdbtools mariadb install --version 11.4
  
  # Instalasi dengan environment variable
  SFDBTOOLS_MARIADB_VERSION=10.11 sudo sfdbtools mariadb install`,
	RunE: executeMariaDBInstall,
}

func init() {
	// Tambah flags untuk konfigurasi instalasi
	InstallCmd.Flags().StringP("version", "v", "", "Versi MariaDB yang akan diinstall (default dari config atau 10.6.23)")

	// Do not print usage automatically on error (we already show an error message / spinner)
	// InstallCmd.SilenceUsage = true
	// Do not let Cobra print the error again; spinner already displayed a user-facing error
	// InstallCmd.SilenceErrors = true

	// Tambah common flags jika tersedia
	// common.AddCommonFlags(InstallCmd) // Uncomment jika ada helper ini
}

// executeMariaDBInstall menjalankan command instalasi MariaDB
func executeMariaDBInstall(cmd *cobra.Command, args []string) error {
	lg, err := logger.Get()
	if err != nil {
		// If logger failed, still print minimal output directly
		terminal.SafePrintln("Gagal inisialisasi logger: " + err.Error())
		return err
	}

	// Clear screen untuk UX yang lebih baik
	terminal.ClearScreen()

	// Resolve konfigurasi dari flags dan environment
	cfg, err := mariadb_config.ResolveMariaDBInstallConfig(cmd)
	if err != nil {
		// Don't duplicate error printing here; return so Cobra prints the error once.
		return err
	}

	cfgPost, err := mariadb_config.ResolveMariaDBConfigureConfig(cmd)
	if err != nil {
		return err
	}

	lg.Info("Konfigurasi instalasi MariaDB",
		logger.String("version", cfg.Version),
		logger.Bool("non_interactive", cfg.NonInteractive))

	// Jalankan instalasi - semua logic di core
	ctx := context.Background()
	if err := install.RunMariaDBInstall(ctx, cfg, cfgPost); err != nil {
		// Spinner already displayed a user-facing error; return the error to Cobra.
		return err
	}

	return nil
}
