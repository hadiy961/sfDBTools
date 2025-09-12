package mariadb

import (
	"github.com/spf13/cobra"
)

// AddMariaDBConfigureFlags menambahkan flags untuk command mariadb configure
func AddMariaDBConfigureFlags(cmd *cobra.Command) {
	// Basic configuration flags
	cmd.Flags().Int("server-id", 0, "Server ID untuk replikasi (1-4294967295)")
	cmd.Flags().Int("port", 0, "Port MariaDB (1024-65535)")

	// Directory configuration flags
	cmd.Flags().String("data-dir", "", "Path direktori data MariaDB (absolute path)")
	cmd.Flags().String("log-dir", "", "Path direktori log MariaDB (absolute path)")
	cmd.Flags().String("binlog-dir", "", "Path direktori binary log MariaDB (absolute path)")

	// Encryption configuration flags
	cmd.Flags().Bool("innodb-encrypt-tables", false, "Aktifkan enkripsi tabel InnoDB")
	cmd.Flags().String("encryption-key-file", "", "Path file kunci enkripsi (absolute path)")

	// Performance tuning flags
	cmd.Flags().String("innodb-buffer-pool-size", "", "Ukuran InnoDB buffer pool (contoh: 1G, 512M)")
	cmd.Flags().Int("innodb-buffer-pool-instances", 0, "Jumlah instance InnoDB buffer pool")

	// Mode configuration flags
	cmd.Flags().Bool("non-interactive", false, "Mode non-interaktif (gunakan default atau nilai flag)")
	cmd.Flags().Bool("auto-tune", true, "Aktifkan auto-tuning berdasarkan resource sistem")

	// Backup and safety flags
	cmd.Flags().Bool("backup-current-config", true, "Backup konfigurasi saat ini sebelum mengubah")
	cmd.Flags().String("backup-dir", "/tmp/sfdbtools_backup", "Direktori untuk backup")

	// Migration flags
	cmd.Flags().Bool("migrate-data", true, "Migrasi data jika direktori berubah")
	cmd.Flags().Bool("verify-migration", true, "Verifikasi integritas data setelah migrasi")
}
