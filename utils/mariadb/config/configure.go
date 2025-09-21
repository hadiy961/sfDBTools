package mariadb

import (
	"fmt"

	"sfDBTools/internal/config"

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
	cmd.Flags().Bool("innodb_encrypt_tables", false, "Aktifkan enkripsi tabel InnoDB")
	cmd.Flags().String("encryption-key-file", "", "Path file kunci enkripsi (absolute path)")

	// Performance tuning flags
	cmd.Flags().String("innodb-buffer-pool-size", "", "Ukuran InnoDB buffer pool (contoh: 1G, 512M)")
	cmd.Flags().Int("innodb-buffer-pool-instances", 0, "Jumlah instance InnoDB buffer pool")

	// Mode configuration flags
	cmd.Flags().Bool("auto-tune", false, "Aktifkan auto-tuning berdasarkan resource sistem")

	// Backup and safety flags
	cmd.Flags().String("backup-dir", "", "Direktori untuk backup")

	// Migration flags
	cmd.Flags().Bool("migrate-data", false, "Migrasi data jika direktori berubah")
}

// ResolveMariaDBConfigureConfig menggunakan pola priority: flags > env > config > defaults
func ResolveMariaDBConfigureConfig(cmd *cobra.Command) (*MariaDBConfigureConfig, error) {
	// Load config file untuk default values
	appConfig, err := config.Get()
	if err != nil {
		return nil, fmt.Errorf("gagal memuat konfigurasi dari config.yaml: %w", err)
	}

	// Server ID: only from flag (no hardcoded default)
	serverID := appConfig.MariaDB.ServerID
	if val, err := cmd.Flags().GetInt("server-id"); err == nil && cmd.Flags().Changed("server-id") {
		serverID = val
	}

	// Port: from config.yaml or flag (no hardcoded default)
	port := appConfig.MariaDB.Port
	if val, err := cmd.Flags().GetInt("port"); err == nil && cmd.Flags().Changed("port") {
		port = val
	}

	// Directory configuration - ambil dari config.yaml atau flag
	dataDir := appConfig.MariaDB.DataDir
	if val, err := cmd.Flags().GetString("data-dir"); err == nil && cmd.Flags().Changed("data-dir") {
		dataDir = val
	}

	logDir := appConfig.MariaDB.LogDir
	if val, err := cmd.Flags().GetString("log-dir"); err == nil && cmd.Flags().Changed("log-dir") {
		logDir = val
	}

	binlogDir := appConfig.MariaDB.BinlogDir
	if val, err := cmd.Flags().GetString("binlog-dir"); err == nil && cmd.Flags().Changed("binlog-dir") {
		binlogDir = val
	}

	// Encryption configuration
	innodbEncryptTables := appConfig.MariaDB.InnodbEncryptTables
	if val, err := cmd.Flags().GetBool("innodb_encrypt_tables"); err == nil && cmd.Flags().Changed("innodb_encrypt_tables") {
		innodbEncryptTables = val
	}
	// Gunakan encryption key dari config.yaml (no hardcoded fallback)
	encryptionKeyFile := appConfig.MariaDB.EncryptionKeyFile
	if val, err := cmd.Flags().GetString("encryption-key-file"); err == nil && cmd.Flags().Changed("encryption-key-file") {
		encryptionKeyFile = val
	}

	// Performance configuration (not present in appConfig model) — use flags only
	innodbBufferPoolSize := ""
	if val, err := cmd.Flags().GetString("innodb-buffer-pool-size"); err == nil && cmd.Flags().Changed("innodb-buffer-pool-size") {
		innodbBufferPoolSize = val
	}

	innodbBufferPoolInstances := 0
	if val, err := cmd.Flags().GetInt("innodb-buffer-pool-instances"); err == nil && cmd.Flags().Changed("innodb-buffer-pool-instances") {
		innodbBufferPoolInstances = val
	}

	// Mode configuration (auto-tune) - only from flag
	autoTune := true
	if val, err := cmd.Flags().GetBool("auto-tune"); err == nil && cmd.Flags().Changed("auto-tune") {
		autoTune = val
	}

	// Backup and safety configuration - only from flag
	backupDir := appConfig.Backup.Storage.BaseDirectory
	if val, err := cmd.Flags().GetString("backup-dir"); err == nil && cmd.Flags().Changed("backup-dir") {
		backupDir = val
	}

	// Migration configuration - only from flag
	migrateData := false
	if val, err := cmd.Flags().GetBool("migrate-data"); err == nil && cmd.Flags().Changed("migrate-data") {
		migrateData = val
	}

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
		AutoTune:                  autoTune,
		BackupDir:                 backupDir,
		MigrateData:               migrateData,
	}

	// Validasi input user (penting untuk konfigurasi sistem)
	if err := validateConfigureInput(mariadbCfg); err != nil {
		return nil, fmt.Errorf("validasi konfigurasi gagal: %w", err)
	}

	return mariadbCfg, nil
}
