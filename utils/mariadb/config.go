package mariadb

import (
	"fmt"
	"path/filepath"
	"strings"

	"sfDBTools/internal/config"
	"sfDBTools/utils/common"

	"github.com/spf13/cobra"
)

// MariaDBInstallConfig berisi konfigurasi untuk instalasi MariaDB
type MariaDBInstallConfig struct {
	Version        string // Versi MariaDB yang akan diinstall
	NonInteractive bool   // Mode non-interactive
}

// MariaDBConfigureConfig berisi konfigurasi untuk setup MariaDB custom
type MariaDBConfigureConfig struct {
	// Basic configuration
	ServerID int `json:"server_id"`
	Port     int `json:"port"`

	// Directory configuration
	DataDir   string `json:"data_dir"`
	LogDir    string `json:"log_dir"`
	BinlogDir string `json:"binlog_dir"`

	// Encryption configuration
	InnodbEncryptTables bool   `json:"innodb_encrypt_tables"`
	EncryptionKeyFile   string `json:"encryption_key_file"`

	// Performance configuration
	InnodbBufferPoolSize      string `json:"innodb_buffer_pool_size"`
	InnodbBufferPoolInstances int    `json:"innodb_buffer_pool_instances"`

	// Mode configuration
	NonInteractive bool `json:"non_interactive"`
	AutoTune       bool `json:"auto_tune"`

	// Backup and safety configuration
	BackupCurrentConfig bool   `json:"backup_current_config"`
	BackupDir           string `json:"backup_dir"`

	// Migration configuration
	MigrateData     bool `json:"migrate_data"`
	VerifyMigration bool `json:"verify_migration"`
}

// MariaDBRemoveConfig berisi konfigurasi untuk penghapusan MariaDB
type MariaDBRemoveConfig struct {
	RemoveData       bool   // Hapus data directory (/var/lib/mysql)
	RemoveConfig     bool   // Hapus file konfigurasi (/etc/mysql, /etc/my.cnf)
	RemoveRepository bool   // Hapus repository MariaDB
	RemoveUser       bool   // Hapus user mysql dari sistem
	Force            bool   // Force removal tanpa konfirmasi
	BackupData       bool   // Backup data sebelum dihapus
	BackupPath       string // Path untuk backup data
	NonInteractive   bool   // Mode non-interactive
}

// ResolveMariaDBInstallConfig membaca flags/env dan default dari config file
func ResolveMariaDBInstallConfig(cmd *cobra.Command) (*MariaDBInstallConfig, error) {
	// Baca konfigurasi dari flags dan environment variables
	version := common.GetStringFlagOrEnv(cmd, "version", "SFDBTOOLS_MARIADB_VERSION", "")
	nonInteractive := common.GetBoolFlagOrEnv(cmd, "non-interactive", "SFDBTOOLS_NON_INTERACTIVE", false)

	// Jika versi tidak ditentukan melalui flag/env, ambil dari config file
	if version == "" {
		cfg, err := config.Get()
		if err != nil {
			// Jika config tidak dapat dimuat, gunakan default hardcoded
			version = "10.6.23"
		} else {
			// Ambil dari config file
			if cfg.MariaDB.Version != "" {
				version = cfg.MariaDB.Version
			} else {
				// Fallback ke default hardcoded jika config kosong
				version = "10.6.23"
			}
		}
	}

	cfg := &MariaDBInstallConfig{
		Version:        version,
		NonInteractive: nonInteractive,
	}

	// Validasi konfigurasi basic (format saja)
	if err := validateVersionFormat(cfg.Version); err != nil {
		return nil, fmt.Errorf("format versi tidak valid: %w", err)
	}

	return cfg, nil
}

// ResolveMariaDBRemoveConfig membaca flags/env untuk konfigurasi penghapusan
func ResolveMariaDBRemoveConfig(cmd *cobra.Command) (*MariaDBRemoveConfig, error) {
	// Baca konfigurasi dari flags dan environment variables
	removeData := common.GetBoolFlagOrEnv(cmd, "remove-data", "SFDBTOOLS_REMOVE_DATA", false)
	removeConfig := common.GetBoolFlagOrEnv(cmd, "remove-config", "SFDBTOOLS_REMOVE_CONFIG", false)
	removeRepository := common.GetBoolFlagOrEnv(cmd, "remove-repository", "SFDBTOOLS_REMOVE_REPOSITORY", false)
	removeUser := common.GetBoolFlagOrEnv(cmd, "remove-user", "SFDBTOOLS_REMOVE_USER", false)
	force := common.GetBoolFlagOrEnv(cmd, "force", "SFDBTOOLS_FORCE", false)
	backupData := common.GetBoolFlagOrEnv(cmd, "backup-data", "SFDBTOOLS_BACKUP_DATA", false)
	backupPath := common.GetStringFlagOrEnv(cmd, "backup-path", "SFDBTOOLS_BACKUP_PATH", "/tmp/mariadb_backup")
	nonInteractive := common.GetBoolFlagOrEnv(cmd, "non-interactive", "SFDBTOOLS_NON_INTERACTIVE", false)

	cfg := &MariaDBRemoveConfig{
		RemoveData:       removeData,
		RemoveConfig:     removeConfig,
		RemoveRepository: removeRepository,
		RemoveUser:       removeUser,
		Force:            force,
		BackupData:       backupData,
		BackupPath:       backupPath,
		NonInteractive:   nonInteractive,
	}

	return cfg, nil
}

// validateVersionFormat melakukan validasi sederhana format versi
func validateVersionFormat(version string) error {
	// Versi harus berupa angka dan titik, misalnya: 10.6, 10.6.23, 11.4
	if len(version) == 0 {
		return fmt.Errorf("versi tidak boleh kosong")
	}

	// Cek apakah mengandung karakter yang valid
	for _, char := range version {
		if char != '.' && (char < '0' || char > '9') {
			return fmt.Errorf("karakter tidak valid dalam versi: %c", char)
		}
	}

	// Minimal harus ada satu titik untuk major.minor
	if !strings.Contains(version, ".") {
		return fmt.Errorf("format versi harus berupa major.minor (contoh: 10.6)")
	}

	return nil
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

// validateConfigureInput melakukan validasi input untuk MariaDB configure
func validateConfigureInput(cfg *MariaDBConfigureConfig) error {
	// Server ID validation
	if cfg.ServerID <= 0 || cfg.ServerID > 4294967295 {
		return fmt.Errorf("server_id harus antara 1 dan 4294967295, diberikan: %d", cfg.ServerID)
	}

	// Port validation
	if cfg.Port < 1024 || cfg.Port > 65535 {
		return fmt.Errorf("port harus antara 1024 dan 65535, diberikan: %d", cfg.Port)
	}

	// Directory validation - harus absolute path
	dirs := map[string]string{
		"data-dir":   cfg.DataDir,
		"log-dir":    cfg.LogDir,
		"binlog-dir": cfg.BinlogDir,
	}

	for name, dir := range dirs {
		if !filepath.IsAbs(dir) {
			return fmt.Errorf("direktori %s harus absolute path: %s", name, dir)
		}
	}

	// Directories must be different

	if cfg.DataDir == cfg.BinlogDir {
		return fmt.Errorf("data-dir dan binlog-dir tidak boleh sama: %s", cfg.DataDir)
	}
	if cfg.LogDir == cfg.BinlogDir {
		return fmt.Errorf("log-dir dan binlog-dir tidak boleh sama: %s", cfg.LogDir)
	}

	// Encryption key file validation (jika encryption enabled)
	if cfg.InnodbEncryptTables && cfg.EncryptionKeyFile != "" {
		if !filepath.IsAbs(cfg.EncryptionKeyFile) {
			return fmt.Errorf("encryption-key-file harus absolute path: %s", cfg.EncryptionKeyFile)
		}
	}

	return nil
}
