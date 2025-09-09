package mariadb_cmd

import (
	"context"

	"sfDBTools/internal/core/mariadb/install"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/mariadb"
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
	InstallCmd.Flags().Bool("non-interactive", false, "Mode non-interactive, tidak menampilkan output interaktif")

	// Tambah common flags jika tersedia
	// common.AddCommonFlags(InstallCmd) // Uncomment jika ada helper ini
}

// executeMariaDBInstall menjalankan command instalasi MariaDB
func executeMariaDBInstall(cmd *cobra.Command, args []string) error {
	lg, err := logger.Get()
	if err != nil {
		terminal.SafePrintln("❌ Gagal inisialisasi logger: " + err.Error())
		return err
	}

	// Clear screen untuk UX yang lebih baik
	if !common.GetBoolFlagOrEnv(cmd, "non-interactive", "SFDBTOOLS_NON_INTERACTIVE", false) {
		terminal.ClearScreen()
	}

	terminal.SafePrintln("🚀 Memulai instalasi MariaDB...")

	// Resolve konfigurasi dari flags dan environment (tanpa interactive)
	cfg, err := mariadb.ResolveMariaDBInstallConfig(cmd)
	if err != nil {
		lg.Error("Gagal resolve konfigurasi", logger.Error(err))
		terminal.SafePrintln("❌ Konfigurasi tidak valid: " + err.Error())
		return err
	}

	lg.Info("Konfigurasi instalasi MariaDB",
		logger.String("version", cfg.Version),
		logger.Bool("non_interactive", cfg.NonInteractive))

	// Jalankan instalasi - semua logic di core
	ctx := context.Background()
	if err := install.RunMariaDBInstall(ctx, cfg); err != nil {
		lg.Error("Instalasi MariaDB gagal", logger.Error(err))
		terminal.SafePrintln("❌ Instalasi gagal: " + err.Error())
		return err
	}

	return nil
}
