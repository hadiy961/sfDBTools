package discovery

// MariaDBInstallation berisi informasi instalasi MariaDB
type MariaDBInstallation struct {
	Version     string   `json:"version"`
	IsInstalled bool     `json:"is_installed"`
	BinaryPath  string   `json:"binary_path"`
	ConfigPaths []string `json:"config_paths"`
	ServiceName string   `json:"service_name"`
	DataDir     string   `json:"data_dir"`
	LogDir      string   `json:"log_dir"`
	BinlogDir   string   `json:"binlog_dir"`
	IsRunning   bool     `json:"is_running"`
	SocketPath  string   `json:"socket_path"`
	Port        int      `json:"port"`
	ServerID    int      `json:"server_id"`
	// Encryption
	InnodbEncryptTables       bool   `json:"innodb_encrypt_tables"`
	EncryptionKeyFile         string `json:"encryption_key_file"`
	InnodbBufferPoolSize      string `json:"innodb_buffer_pool_size"`
	InnodbBufferPoolInstances int    `json:"innodb_buffer_pool_instances"`
	// Backup
	BackupDir string `json:"backup_dir"`
}
