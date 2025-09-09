package mariadb_cmd

import (
	"context"

	"sfDBTools/internal/core/mariadb/remove"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/mariadb"
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
	RunE: executeMariaDBRemove,
}

func init() {
	// Tambah flags untuk konfigurasi penghapusan
	RemoveCmd.Flags().Bool("remove-data", false, "Hapus data directory (/var/lib/mysql) - SEMUA DATABASE AKAN HILANG!")
	RemoveCmd.Flags().Bool("remove-config", false, "Hapus file konfigurasi (/etc/mysql, /etc/my.cnf)")
	RemoveCmd.Flags().Bool("remove-repository", false, "Hapus repository MariaDB dari sistem")
	RemoveCmd.Flags().Bool("remove-user", false, "Hapus user sistem 'mysql'")
	RemoveCmd.Flags().Bool("force", false, "Force removal tanpa konfirmasi (BERBAHAYA!)")
	RemoveCmd.Flags().Bool("backup-data", false, "Backup data sebelum dihapus")
	RemoveCmd.Flags().String("backup-path", "/tmp/mariadb_backup", "Path untuk menyimpan backup data")
	RemoveCmd.Flags().Bool("non-interactive", false, "Mode non-interactive, tidak menampilkan output interaktif")
}

// executeMariaDBRemove menjalankan command penghapusan MariaDB
func executeMariaDBRemove(cmd *cobra.Command, args []string) error {
	lg, err := logger.Get()
	if err != nil {
		terminal.SafePrintln("❌ Gagal inisialisasi logger: " + err.Error())
		return err
	}

	// Clear screen untuk UX yang lebih baik
	if !common.GetBoolFlagOrEnv(cmd, "non-interactive", "SFDBTOOLS_NON_INTERACTIVE", false) {
		terminal.ClearScreen()
	}

	terminal.SafePrintln("🗑️  Memulai penghapusan MariaDB...")

	// Resolve konfigurasi dari flags dan environment
	cfg, err := mariadb.ResolveMariaDBRemoveConfig(cmd)
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
