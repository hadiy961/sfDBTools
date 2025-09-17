package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sfDBTools/internal/logger"
)

// detectDataDirAndSocket mendeteksi data directory dan socket path
func detectDataDirAndSocket(installation *MariaDBInstallation) error {
	lg, _ := logger.Get()
	for _, configPath := range installation.ConfigPaths {
		if err := parseConfigFile(configPath, installation); err != nil {
			lg.Debug("Gagal parsing config file", logger.String("path", configPath), logger.Error(err))
			continue
		}
	}

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
	inServerSection := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") {
			// track both [mysqld]/[mariadb] and [server] sections
			inMysqldSection = (line == "[mysqld]" || line == "[mariadb]")
			inServerSection = (line == "[server]")
			continue
		}
		if !(inMysqldSection || inServerSection) {
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
			case "server_id":
				if v, err := strconv.Atoi(value); err == nil {
					installation.ServerID = v
				}
			case "innodb_encrypt_tables":
				installation.InnodbEncryptTables = (strings.ToUpper(value) == "ON")
			case "file_key_management_filename":
				installation.EncryptionKeyFile = value
			case "innodb_buffer_pool_size":
				installation.InnodbBufferPoolSize = value
			case "innodb_buffer_pool_instances":
				if v, err := strconv.Atoi(value); err == nil {
					installation.InnodbBufferPoolInstances = v
				}
			case "datadir":
				installation.DataDir = value
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
