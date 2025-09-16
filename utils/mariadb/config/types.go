package mariadb

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
	ConfigDir string `json:"config_dir"`
	SocketDir string `json:"socket_dir"`

	// Encryption configuration
	InnodbEncryptTables bool   `json:"innodb_encrypt_tables"`
	EncryptionKeyFile   string `json:"encryption_key_file"`

	// Performance configuration
	InnodbBufferPoolSize      string `json:"innodb_buffer_pool_size"`
	InnodbBufferPoolInstances int    `json:"innodb_buffer_pool_instances"`

	// Mode configuration
	AutoTune bool `json:"auto_tune"`

	// Backup and safety configuration
	BackupDir string `json:"backup_dir"`

	// Migration configuration
	MigrateData bool `json:"migrate_data"`
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
