package mariadb

// Common types and interfaces for MariaDB module

// Config represents common MariaDB configuration
type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Database string `json:"database"`
}

// ServiceConfig represents configuration for MariaDB services
type ServiceConfig struct {
	DataDir    string `json:"data_dir"`
	ConfigFile string `json:"config_file"`
	LogFile    string `json:"log_file"`
	PidFile    string `json:"pid_file"`
}

// Status represents MariaDB service status
type Status struct {
	Running  bool   `json:"running"`
	Version  string `json:"version"`
	Uptime   string `json:"uptime"`
	PID      int    `json:"pid"`
	DataDir  string `json:"data_dir"`
	ErrorLog string `json:"error_log"`
}
