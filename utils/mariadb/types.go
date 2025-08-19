package mariadb

import "time"

// UninstallOptions represents the configuration for MariaDB uninstall
type UninstallOptions struct {
	Force       bool   `json:"force"`        // Skip confirmation prompts
	KeepData    bool   `json:"keep_data"`    // Keep data directories
	KeepConfig  bool   `json:"keep_config"`  // Keep configuration files
	BackupFirst bool   `json:"backup_first"` // Create backup before uninstall
	BackupDir   string `json:"backup_dir"`   // Directory for backup files
}

// UninstallResult represents the result of MariaDB uninstall operation
type UninstallResult struct {
	Success             bool          `json:"success"`
	OperatingSystem     string        `json:"operating_system"`
	Distribution        string        `json:"distribution"`
	ServiceStatus       string        `json:"service_status"`
	PackagesRemoved     int           `json:"packages_removed"`
	DirectoriesRemoved  []string      `json:"directories_removed"`
	ConfigFilesRemoved  []string      `json:"config_files_removed"`
	RepositoriesRemoved []string      `json:"repositories_removed"`
	Duration            time.Duration `json:"duration"`
	BackupCreated       bool          `json:"backup_created"`
	BackupLocation      string        `json:"backup_location,omitempty"`
	Warnings            []string      `json:"warnings,omitempty"`
	Errors              []string      `json:"errors,omitempty"`
}

// InstallOptions represents the configuration for MariaDB installation
type InstallOptions struct {
	Version     string `json:"version"`
	Port        int    `json:"port"`
	DataDir     string `json:"data_dir"`
	LogDir      string `json:"log_dir"`
	BinlogDir   string `json:"binlog_dir"`
	KeyFile     string `json:"key_file"`
	Force       bool   `json:"force"`
	CustomPaths bool   `json:"custom_paths"`
}

// InstallResult represents the result of MariaDB installation operation
type InstallResult struct {
	Success         bool          `json:"success"`
	Version         string        `json:"version"`
	Port            int           `json:"port"`
	DataDir         string        `json:"data_dir"`
	LogDir          string        `json:"log_dir"`
	BinlogDir       string        `json:"binlog_dir"`
	OperatingSystem string        `json:"operating_system"`
	Distribution    string        `json:"distribution"`
	ServiceStatus   string        `json:"service_status"`
	Duration        time.Duration `json:"duration"`
}

// OSInfo represents operating system information
type OSInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Codename string `json:"codename,omitempty"`
}

// ServiceInfo represents MariaDB service information
type ServiceInfo struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Enabled bool   `json:"enabled"`
	Active  bool   `json:"active"`
	Running bool   `json:"running"`
}

// PackageInfo represents package information
type PackageInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Status  string `json:"status"`
}
