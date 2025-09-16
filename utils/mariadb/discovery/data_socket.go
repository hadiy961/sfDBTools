package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/internal/logger"
)

// detectDataDirAndSocket mendeteksi data directory dan socket path
func detectDataDirAndSocket(installation *MariaDBInstallation) error {
	lg, _ := logger.Get()
	installation.DataDir = "/var/lib/mysql"
	installation.BinlogDir = "/var/lib/mysqlbinlogs"
	installation.SocketPath = "/var/lib/mysql/mysql.sock"
	installation.LogDir = installation.DataDir // default to data dir unless overridden by config
	installation.Port = 3306
	for _, configPath := range installation.ConfigPaths {
		if err := parseConfigFile(configPath, installation); err != nil {
			lg.Debug("Gagal parsing config file", logger.String("path", configPath), logger.Error(err))
			continue
		}
	}
	// Log more details about the discovered installation to help debugging
	lg.Info("Data directory dan socket terdeteksi",
		logger.String("data_dir", installation.DataDir),
		logger.String("socket", installation.SocketPath),
		logger.Int("port", installation.Port),
		logger.String("binlog_dir", installation.BinlogDir),
		logger.String("log_dir", installation.LogDir),
		logger.Strings("config_paths", installation.ConfigPaths),
		logger.String("service_name", installation.ServiceName),
		logger.Bool("is_running", installation.IsRunning),
		logger.String("binary_path", installation.BinaryPath),
	)
	return nil
}

// parseConfigFile parsing file konfigurasi untuk mencari datadir, socket, port
func parseConfigFile(configPath string, installation *MariaDBInstallation) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("gagal membaca file %s: %w", configPath, err)
	}
	lines := strings.Split(string(content), "\n")
	inMysqldSection := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") {
			inMysqldSection = (line == "[mysqld]" || line == "[mariadb]")
			continue
		}
		if !inMysqldSection {
			continue
		}
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			switch key {
			case "datadir":
				installation.DataDir = value
				// If log dir wasn't set explicitly, keep LogDir default in sync with DataDir
				if installation.LogDir == "" {
					installation.LogDir = value
				}
			case "socket":
				installation.SocketPath = value
			case "log_bin":
				if strings.Contains(value, "/") {
					installation.BinlogDir = filepath.Dir(value)
				} else {
					installation.BinlogDir = filepath.Join(installation.DataDir, value)
				}
			case "log_error":
				// log_error may be a filename or full path. If it contains '/', take dirname.
				if strings.Contains(value, "/") {
					installation.LogDir = filepath.Dir(value)
				} else {
					installation.LogDir = filepath.Join(installation.DataDir, value)
				}
			case "general_log_file":
				if strings.Contains(value, "/") {
					installation.LogDir = filepath.Dir(value)
				} else {
					installation.LogDir = filepath.Join(installation.DataDir, value)
				}
			case "slow_query_log_file":
				if strings.Contains(value, "/") {
					installation.LogDir = filepath.Dir(value)
				} else {
					installation.LogDir = filepath.Join(installation.DataDir, value)
				}
			case "port":
				if port, err := parsePort(value); err == nil {
					installation.Port = port
				}
			}
		}
	}
	return nil
}

// parsePort parsing port dari string
func parsePort(portStr string) (int, error) {
	var port int
	_, err := fmt.Sscanf(portStr, "%d", &port)
	return port, err
}
