package mariadb_cmd

import (
	"context"

	"sfDBTools/internal/core/mariadb/remove"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	mariadb_config "sfDBTools/utils/mariadb/config"

	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

// RemoveCmd menghapus instalasi MariaDB secara menyeluruh
var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Hapus instalasi MariaDB secara menyeluruh sampai akar-akarnya",
	Long: `Hapus instalasi MariaDB secara menyeluruh dari sistem.

Command ini akan:
1. Menghentikan semua service MariaDB yang berjalan
2. Menghapus semua paket MariaDB (server, client, libraries)
3. Menghapus data directory (/var/lib/mysql) jika diminta
4. Menghapus file konfigurasi (/etc/mysql, /etc/my.cnf) jika diminta
5. Menghapus repository MariaDB jika diminta
6. Menghapus user sistem 'mysql' jika diminta
7. Membersihkan log files dan temp files
8. Memverifikasi penghapusan lengkap

PERINGATAN: 
- Proses ini TIDAK DAPAT DIBATALKAN
- Pastikan backup data penting sebelum menjalankan
- Semua database akan HILANG jika --remove-data digunakan

Penghapusan memerlukan hak akses root (sudo).

Contoh penggunaan:
  # Hapus paket saja (data dan config dipertahankan)
  sudo sfdbtools mariadb remove

  # Hapus lengkap termasuk data (BERBAHAYA!)
  sudo sfdbtools mariadb remove --remove-data --remove-config

  # Hapus lengkap dengan backup data terlebih dahulu
  sudo sfdbtools mariadb remove --remove-data --backup-data --backup-path /backup

  # Hapus force tanpa konfirmasi (untuk automation)
  sudo sfdbtools mariadb remove --remove-data --force

  # Hapus semua termasuk repository dan user sistem
  sudo sfdbtools mariadb remove --remove-data --remove-config --remove-repository --remove-user`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeMariaDBRemove(cmd, args, Lg); err != nil {
			terminal.PrintError("Instalasi MariaDB gagal")
			terminal.WaitForEnterWithMessage("Tekan Enter untuk melanjutkan...")
			// Jangan panggil os.Exit di sini; biarkan Cobra menangani exit code
		} else {
			terminal.PrintSuccess("Instalasi MariaDB selesai")
			terminal.WaitForEnterWithMessage("Tekan Enter untuk melanjutkan...")
			return
		}
	},
}

// executeMariaDBRemove menjalankan command penghapusan MariaDB
func executeMariaDBRemove(cmd *cobra.Command, args []string, lg *logger.Logger) error {

	// Clear screen untuk UX yang lebih baik
	if !common.GetBoolFlagOrEnv(cmd, "non-interactive", "SFDBTOOLS_NON_INTERACTIVE", false) {
		terminal.ClearScreen()
	}

	// Resolve konfigurasi dari flags dan environment
	cfg, err := mariadb_config.ResolveMariaDBRemoveConfig(cmd)
	if err != nil {
		lg.Error("Gagal resolve konfigurasi", logger.Error(err))
		terminal.SafePrintln("❌ Konfigurasi tidak valid: " + err.Error())
		return err
	}

	lg.Info("Konfigurasi penghapusan MariaDB",
		logger.Bool("remove_data", cfg.RemoveData),
		logger.Bool("remove_config", cfg.RemoveConfig),
		logger.Bool("remove_repository", cfg.RemoveRepository),
		logger.Bool("remove_user", cfg.RemoveUser),
		logger.Bool("force", cfg.Force),
		logger.Bool("backup_data", cfg.BackupData))

	// Jalankan penghapusan - semua logic di core
	ctx := context.Background()
	if err := remove.RunMariaDBRemove(ctx, cfg); err != nil {
		lg.Error("Penghapusan MariaDB gagal", logger.Error(err))
		terminal.SafePrintln("❌ Penghapusan gagal: " + err.Error())
		return err
	}

	return nil
}
