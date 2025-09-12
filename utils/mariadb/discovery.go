package mariadb

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
)

// MariaDBInstallation berisi informasi instalasi MariaDB
type MariaDBInstallation struct {
	Version     string   `json:"version"`
	IsInstalled bool     `json:"is_installed"`
	BinaryPath  string   `json:"binary_path"`
	ConfigPaths []string `json:"config_paths"`
	ServiceName string   `json:"service_name"`
	DataDir     string   `json:"data_dir"`
	IsRunning   bool     `json:"is_running"`
	SocketPath  string   `json:"socket_path"`
	Port        int      `json:"port"`
}

// DiscoverMariaDBInstallation mendeteksi instalasi MariaDB di sistem
func DiscoverMariaDBInstallation() (*MariaDBInstallation, error) {
	lg, _ := logger.Get()
	lg.Info("Memulai discovery instalasi MariaDB")

	installation := &MariaDBInstallation{
		ConfigPaths: []string{},
	}

	// Deteksi binary MariaDB
	if err := detectMariaDBBinary(installation); err != nil {
		lg.Warn("Gagal mendeteksi binary MariaDB", logger.Error(err))
	}

	// Deteksi versi
	if installation.BinaryPath != "" {
		if err := detectMariaDBVersion(installation); err != nil {
			lg.Warn("Gagal mendeteksi versi MariaDB", logger.Error(err))
		}
	}

	// Deteksi file konfigurasi
	if err := detectConfigFiles(installation); err != nil {
		lg.Warn("Gagal mendeteksi file konfigurasi", logger.Error(err))
	}

	// Deteksi service
	if err := detectMariaDBService(installation); err != nil {
		lg.Warn("Gagal mendeteksi service MariaDB", logger.Error(err))
	}

	// Deteksi data directory dan socket
	if err := detectDataDirAndSocket(installation); err != nil {
		lg.Warn("Gagal mendeteksi data directory dan socket", logger.Error(err))
	}

	lg.Info("Discovery MariaDB selesai",
		logger.Bool("is_installed", installation.IsInstalled),
		logger.String("version", installation.Version),
		logger.Bool("is_running", installation.IsRunning))

	return installation, nil
}

// detectMariaDBBinary mendeteksi binary MariaDB/MySQL
func detectMariaDBBinary(installation *MariaDBInstallation) error {
	lg, _ := logger.Get()

	// Coba berbagai nama binary yang mungkin
	binaries := []string{"mariadb", "mysql", "mysqld", "mariadbd"}

	for _, binary := range binaries {
		path, err := exec.LookPath(binary)
		if err == nil {
			lg.Debug("Ditemukan binary", logger.String("binary", binary), logger.String("path", path))
			installation.BinaryPath = path
			installation.IsInstalled = true
			return nil
		}
	}

	// Coba lokasi standar
	standardPaths := []string{
		"/usr/bin/mariadb",
		"/usr/bin/mysql",
		"/usr/sbin/mysqld",
		"/usr/sbin/mariadbd",
		"/usr/local/bin/mariadb",
		"/usr/local/bin/mysql",
	}

	for _, path := range standardPaths {
		if _, err := os.Stat(path); err == nil {
			lg.Debug("Ditemukan binary di path standar", logger.String("path", path))
			installation.BinaryPath = path
			installation.IsInstalled = true
			return nil
		}
	}

	return fmt.Errorf("binary MariaDB/MySQL tidak ditemukan")
}

// detectMariaDBVersion mendeteksi versi MariaDB
func detectMariaDBVersion(installation *MariaDBInstallation) error {
	lg, _ := logger.Get()

	// Coba dengan --version
	cmd := exec.Command(installation.BinaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("gagal menjalankan %s --version: %w", installation.BinaryPath, err)
	}

	version := parseVersionFromOutput(string(output))
	if version != "" {
		installation.Version = version
		lg.Info("Terdeteksi versi MariaDB", logger.String("version", version))
		return nil
	}

	return fmt.Errorf("gagal parsing versi dari output: %s", string(output))
}

// parseVersionFromOutput parsing versi dari output command
func parseVersionFromOutput(output string) string {
	// Regex untuk menangkap versi MariaDB
	patterns := []string{
		`mariadb\s+Ver\s+(\d+\.\d+\.\d+)`,
		`MariaDB\s+(\d+\.\d+\.\d+)`,
		`mysql\s+Ver\s+(\d+\.\d+\.\d+).*MariaDB`,
		`(\d+\.\d+\.\d+)-MariaDB`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		matches := re.FindStringSubmatch(output)
		if len(matches) >= 2 {
			return matches[1]
		}
	}

	return ""
}

// detectConfigFiles mendeteksi file konfigurasi MariaDB
func detectConfigFiles(installation *MariaDBInstallation) error {
	lg, _ := logger.Get()

	// Lokasi standar file konfigurasi MariaDB/MySQL
	configPaths := []string{
		"/etc/my.cnf",
		"/etc/mysql/my.cnf",
		"/etc/my.cnf.d/server.cnf",
		"/etc/my.cnf.d/50-server.cnf",
		"/etc/my.cnf.d/mariadb-server.cnf",
		"/etc/mysql/mariadb.conf.d/50-server.cnf",
		"/etc/mysql/conf.d/mysql.cnf",
		"/usr/local/etc/my.cnf",
	}

	// Tambahkan template sfDBTools
	configPaths = append(configPaths, "/etc/sfDBTools/server.cnf")

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			installation.ConfigPaths = append(installation.ConfigPaths, path)
			lg.Debug("Ditemukan file konfigurasi", logger.String("path", path))
		}
	}

	if len(installation.ConfigPaths) == 0 {
		return fmt.Errorf("tidak ditemukan file konfigurasi MariaDB")
	}

	lg.Info("Ditemukan file konfigurasi",
		logger.Int("count", len(installation.ConfigPaths)),
		logger.Strings("paths", installation.ConfigPaths))

	return nil
}

// detectMariaDBService mendeteksi service MariaDB
func detectMariaDBService(installation *MariaDBInstallation) error {
	lg, _ := logger.Get()

	// Nama service yang mungkin
	serviceNames := []string{"mariadb", "mysql", "mysqld"}

	for _, serviceName := range serviceNames {
		// Cek dengan systemctl
		cmd := exec.Command("systemctl", "is-active", serviceName)
		output, err := cmd.Output()
		if err == nil {
			status := strings.TrimSpace(string(output))
			installation.ServiceName = serviceName
			installation.IsRunning = (status == "active")
			lg.Info("Ditemukan service",
				logger.String("service", serviceName),
				logger.Bool("is_running", installation.IsRunning))
			return nil
		}
	}

	// Jika systemctl tidak tersedia, coba dengan ps
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err == nil {
		outputStr := string(output)
		for _, serviceName := range serviceNames {
			if strings.Contains(outputStr, serviceName) {
				installation.ServiceName = serviceName
				installation.IsRunning = true
				lg.Info("Ditemukan process service", logger.String("service", serviceName))
				return nil
			}
		}
	}

	return fmt.Errorf("service MariaDB tidak ditemukan")
}

// detectDataDirAndSocket mendeteksi data directory dan socket path
func detectDataDirAndSocket(installation *MariaDBInstallation) error {
	lg, _ := logger.Get()

	// Default values
	installation.DataDir = "/var/lib/mysql"
	installation.SocketPath = "/var/lib/mysql/mysql.sock"
	installation.Port = 3306

	// Coba baca dari file konfigurasi
	for _, configPath := range installation.ConfigPaths {
		if err := parseConfigFile(configPath, installation); err != nil {
			lg.Debug("Gagal parsing config file",
				logger.String("path", configPath),
				logger.Error(err))
			continue
		}
	}

	lg.Info("Data directory dan socket terdeteksi",
		logger.String("data_dir", installation.DataDir),
		logger.String("socket", installation.SocketPath),
		logger.Int("port", installation.Port))

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

		// Cek section
		if strings.HasPrefix(line, "[") {
			inMysqldSection = (line == "[mysqld]" || line == "[mariadb]")
			continue
		}

		// Skip jika tidak di section yang benar
		if !inMysqldSection {
			continue
		}

		// Skip comments
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Parse key = value
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
			case "socket":
				installation.SocketPath = value
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
	// Simple parsing untuk port
	var port int
	_, err := fmt.Sscanf(portStr, "%d", &port)
	return port, err
}

// GetTemplateConfigPath mencari template konfigurasi sfDBTools
func GetTemplateConfigPath() (string, error) {
	templatePath := "/etc/sfDBTools/server.cnf"

	// Cek apakah template ada
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return "", fmt.Errorf("template konfigurasi tidak ditemukan di: %s", templatePath)
	}

	return templatePath, nil
}

// GetMainConfigPath mencari file konfigurasi utama MariaDB
func GetMainConfigPath(installation *MariaDBInstallation) (string, error) {
	// Prioritas file konfigurasi (sesuai urutan discovery)
	priorityPaths := []string{
		"/etc/my.cnf.d/server.cnf",                // Yang paling umum di CentOS/RHEL
		"/etc/my.cnf.d/50-server.cnf",             // Alternative
		"/etc/my.cnf.d/mariadb-server.cnf",        // MariaDB specific
		"/etc/mysql/mariadb.conf.d/50-server.cnf", // Ubuntu/Debian
		"/etc/my.cnf",                             // Global config
		"/etc/mysql/my.cnf",                       // Ubuntu global
	}

	for _, path := range priorityPaths {
		for _, configPath := range installation.ConfigPaths {
			if configPath == path {
				return path, nil
			}
		}
	}

	// Jika tidak ditemukan, gunakan yang pertama dari detected configs
	if len(installation.ConfigPaths) > 0 {
		return installation.ConfigPaths[0], nil
	}

	return "", fmt.Errorf("tidak ditemukan file konfigurasi utama MariaDB")
}

// String mengembalikan representasi string dari MariaDBInstallation
func (mi *MariaDBInstallation) String() string {
	status := "Not Installed"
	if mi.IsInstalled {
		if mi.IsRunning {
			status = "Installed and Running"
		} else {
			status = "Installed but Not Running"
		}
	}

	return fmt.Sprintf("MariaDB %s - %s (Service: %s, Data: %s, Socket: %s, Port: %d)",
		mi.Version, status, mi.ServiceName, mi.DataDir, mi.SocketPath, mi.Port)
}

// CreateDatabaseConfigFromInstallation membuat database config dari installation info
func CreateDatabaseConfigFromInstallation(installation *MariaDBInstallation) *database.Config {
	if installation == nil {
		return nil
	}

	// Buat basic config dengan informasi yang tersedia
	config := &database.Config{
		Host:     "localhost",
		Port:     installation.Port,
		User:     "root", // Default user, bisa di-override
		Password: "",     // Empty password, perlu di-set sesuai kebutuhan
		DBName:   "",     // Empty database name untuk koneksi sistem
	}

	return config
}
