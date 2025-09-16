package mariadb

import (
	"fmt"

	"sfDBTools/internal/config"
	"sfDBTools/utils/common"

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
	cmd.Flags().String("backup-dir", "config/backups", "Direktori untuk backup")

	// Migration flags
	cmd.Flags().Bool("migrate-data", true, "Migrasi data jika direktori berubah")
	cmd.Flags().Bool("verify-migration", true, "Verifikasi integritas data setelah migrasi")
}

// ResolveMariaDBConfigureConfig menggunakan pola priority: flags > env > config > defaults
func ResolveMariaDBConfigureConfig(cmd *cobra.Command) (*MariaDBConfigureConfig, error) {
	// Load config file untuk default values
	appConfig, err := config.Get()
	if err != nil {
		// Jika gagal load config, lanjutkan dengan default hardcoded
		defaultCfg := &MariaDBConfigureConfig{
			DataDir:   "/var/lib/mysql",
			LogDir:    "/var/lib/mysql",
			BinlogDir: "/var/lib/mysqlbinlogs",
			Port:      3306,
		}
		fmt.Println("Peringatan: Gagal memuat konfigurasi, menggunakan default hardcoded")
		return defaultCfg, nil
	}

	// Resolve dari flags -> env -> config.yaml -> defaults
	serverID := common.GetIntFlagOrEnv(cmd, "server-id", "SFDBTOOLS_MARIADB_SERVER_ID", 1)

	// Port: prioritas flags > env > config.yaml > default
	port := appConfig.MariaDB.Port
	if port == 0 {
		port = 3306 // default jika tidak ada di config
	}
	port = common.GetIntFlagOrEnv(cmd, "port", "SFDBTOOLS_MARIADB_PORT", port)

	// Directory configuration - ambil dari config.yaml
	dataDir := common.GetStringFlagOrEnv(cmd, "data-dir", "SFDBTOOLS_MARIADB_DATA_DIR", appConfig.MariaDB.DataDir)
	logDir := common.GetStringFlagOrEnv(cmd, "log-dir", "SFDBTOOLS_MARIADB_LOG_DIR", appConfig.MariaDB.LogDir)
	binlogDir := common.GetStringFlagOrEnv(cmd, "binlog-dir", "SFDBTOOLS_MARIADB_BINLOG_DIR", appConfig.MariaDB.BinlogDir)

	// Encryption configuration
	innodbEncryptTables := common.GetBoolFlagOrEnv(cmd, "innodb-encrypt-tables", "SFDBTOOLS_MARIADB_ENCRYPT_TABLES", false)
	// Gunakan encryption key dari config.yaml sebagai default
	defaultEncryptionKey := appConfig.ConfigDir.MariaDBKey
	if defaultEncryptionKey == "" {
		defaultEncryptionKey = "/var/lib/mysql/encryption/keyfile" // fallback jika tidak ada di config
	}
	encryptionKeyFile := common.GetStringFlagOrEnv(cmd, "encryption-key-file", "SFDBTOOLS_MARIADB_ENCRYPTION_KEY_FILE", defaultEncryptionKey)

	// Performance configuration
	innodbBufferPoolSize := common.GetStringFlagOrEnv(cmd, "innodb-buffer-pool-size", "SFDBTOOLS_MARIADB_BUFFER_POOL_SIZE", "128M")
	innodbBufferPoolInstances := common.GetIntFlagOrEnv(cmd, "innodb-buffer-pool-instances", "SFDBTOOLS_MARIADB_BUFFER_POOL_INSTANCES", 8)

	// Mode configuration
	nonInteractive := common.GetBoolFlagOrEnv(cmd, "non-interactive", "SFDBTOOLS_NON_INTERACTIVE", false)
	autoTune := common.GetBoolFlagOrEnv(cmd, "auto-tune", "SFDBTOOLS_MARIADB_AUTO_TUNE", true)

	// Backup and safety configuration
	backupCurrentConfig := common.GetBoolFlagOrEnv(cmd, "backup-current-config", "SFDBTOOLS_MARIADB_BACKUP_CONFIG", true)
	backupDir := common.GetStringFlagOrEnv(cmd, "backup-dir", "SFDBTOOLS_MARIADB_BACKUP_DIR", "config/backups")

	// Migration configuration
	migrateData := common.GetBoolFlagOrEnv(cmd, "migrate-data", "SFDBTOOLS_MARIADB_MIGRATE_DATA", true)
	verifyMigration := common.GetBoolFlagOrEnv(cmd, "verify-migration", "SFDBTOOLS_MARIADB_VERIFY_MIGRATION", true)

	mariadbCfg := &MariaDBConfigureConfig{
		ServerID:                  serverID,
		Port:                      port,
		DataDir:                   dataDir,
		LogDir:                    logDir,
		BinlogDir:                 binlogDir,
		InnodbEncryptTables:       innodbEncryptTables,
		EncryptionKeyFile:         encryptionKeyFile,
		InnodbBufferPoolSize:      innodbBufferPoolSize,
		InnodbBufferPoolInstances: innodbBufferPoolInstances,
		NonInteractive:            nonInteractive,
		AutoTune:                  autoTune,
		BackupCurrentConfig:       backupCurrentConfig,
		BackupDir:                 backupDir,
		MigrateData:               migrateData,
		VerifyMigration:           verifyMigration,
	}

	// Validasi input user (penting untuk konfigurasi sistem)
	if err := validateConfigureInput(mariadbCfg); err != nil {
		return nil, fmt.Errorf("validasi konfigurasi gagal: %w", err)
	}

	return mariadbCfg, nil
}
