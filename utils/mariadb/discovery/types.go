package mariadb

// MariaDBInstallation berisi informasi instalasi MariaDB
type MariaDBInstallation struct {
	Version     string   `json:"version"`
	IsInstalled bool     `json:"is_installed"`
	BinaryPath  string   `json:"binary_path"`
	ConfigPaths []string `json:"config_paths"`
	ServiceName string   `json:"service_name"`
	DataDir     string   `json:"data_dir"`
	BinlogDir   string   `json:"binlog_dir"`
	IsRunning   bool     `json:"is_running"`
	SocketPath  string   `json:"socket_path"`
	Port        int      `json:"port"`
}
