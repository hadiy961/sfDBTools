package remove

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// MariaDBConfig berisi path direktori yang dikonfigurasi dalam MariaDB
type MariaDBConfig struct {
	DataDir    string
	InnoDBDir  string
	BinlogDir  string
	LogDir     string
	ErrorLog   string
	SlowLog    string
	GeneralLog string
	TmpDir     string
	Socket     string
}

// detectCustomDirectories mendeteksi direktori custom dari file konfigurasi MariaDB
func detectCustomDirectories() (*MariaDBConfig, error) {
	lg, _ := logger.Get()
	terminal.SafePrintln("🔍 Mendeteksi direktori custom dari konfigurasi MariaDB...")

	config := &MariaDBConfig{
		// Default values jika tidak ditemukan di config
		DataDir:   "/var/lib/mysql",
		InnoDBDir: "/var/lib/mysql",
		BinlogDir: "/var/lib/mysql",
		LogDir:    "/var/log/mysql",
		TmpDir:    "/tmp",
		Socket:    "/var/run/mysqld/mysqld.sock",
	}

	// Daftar file konfigurasi yang mungkin ada (urutan prioritas)
	configFiles := []string{
		"/etc/mysql/my.cnf",
		"/etc/my.cnf",
		"/etc/my.cnf.d/server.cnf",
		"/etc/mysql/mysql.conf.d/mysqld.cnf",
		"/etc/mysql/mariadb.conf.d/50-server.cnf",
		"/etc/mariadb/my.cnf",
		"/usr/local/mysql/my.cnf",
		"~/.my.cnf",
	}

	// Parse setiap file konfigurasi
	for _, configFile := range configFiles {
		// Expand tilde untuk home directory
		if strings.HasPrefix(configFile, "~/") {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				configFile = filepath.Join(homeDir, configFile[2:])
			}
		}

		if _, err := os.Stat(configFile); err == nil {
			terminal.SafePrintln("   📄 Membaca: " + configFile)
			if err := parseConfigFile(configFile, config); err != nil {
				lg.Warn("Gagal parse config file", logger.String("file", configFile), logger.Error(err))
			}
		}
	}

	// Tampilkan direktori yang terdeteksi
	displayDetectedDirectories(config)

	lg.Info("Deteksi direktori custom selesai",
		logger.String("datadir", config.DataDir),
		logger.String("innodb_dir", config.InnoDBDir),
		logger.String("binlog_dir", config.BinlogDir),
		logger.String("log_dir", config.LogDir))

	return config, nil
}

// parseConfigFile membaca dan parse file konfigurasi MariaDB
func parseConfigFile(filename string, config *MariaDBConfig) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("gagal membuka file %s: %w", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines dan comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Deteksi section header [mysqld], [mariadb], etc
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(strings.Trim(line, "[]"))
			continue
		}

		// Hanya parse jika dalam section yang relevan
		if currentSection != "mysqld" && currentSection != "mariadb" && currentSection != "server" {
			continue
		}

		// Parse key=value pairs
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Remove quotes if present
				value = strings.Trim(value, `"'`)

				parseConfigValue(key, value, config)
			}
		}
	}

	return scanner.Err()
}

// parseConfigValue mengparse nilai konfigurasi dan assign ke struct
func parseConfigValue(key, value string, config *MariaDBConfig) {
	key = strings.ToLower(strings.ReplaceAll(key, "-", "_"))

	switch key {
	case "datadir":
		config.DataDir = value
	case "innodb_data_home_dir":
		config.InnoDBDir = value
	case "log_bin", "log_bin_basename":
		// Extract directory from log-bin path
		if value != "" && value != "1" && value != "ON" {
			config.BinlogDir = filepath.Dir(value)
		}
	case "log_bin_dirname":
		config.BinlogDir = value
	case "log_error":
		config.ErrorLog = value
		config.LogDir = filepath.Dir(value)
	case "slow_query_log_file":
		config.SlowLog = value
		if config.LogDir == "/var/log/mysql" { // Only update if still default
			config.LogDir = filepath.Dir(value)
		}
	case "general_log_file":
		config.GeneralLog = value
		if config.LogDir == "/var/log/mysql" { // Only update if still default
			config.LogDir = filepath.Dir(value)
		}
	case "tmpdir":
		config.TmpDir = value
	case "socket":
		config.Socket = value
	}
}

// displayDetectedDirectories menampilkan direktori yang terdeteksi
func displayDetectedDirectories(config *MariaDBConfig) {
	terminal.SafePrintln("   📂 Direktori yang terdeteksi:")
	terminal.SafePrintln("      Data directory: " + config.DataDir)

	if config.InnoDBDir != config.DataDir {
		terminal.SafePrintln("      InnoDB directory: " + config.InnoDBDir)
	}

	if config.BinlogDir != config.DataDir {
		terminal.SafePrintln("      Binary log directory: " + config.BinlogDir)
	}

	if config.LogDir != "/var/log/mysql" {
		terminal.SafePrintln("      Log directory: " + config.LogDir)
	}

	if config.ErrorLog != "" {
		terminal.SafePrintln("      Error log: " + config.ErrorLog)
	}

	if config.SlowLog != "" {
		terminal.SafePrintln("      Slow query log: " + config.SlowLog)
	}

	if config.GeneralLog != "" {
		terminal.SafePrintln("      General log: " + config.GeneralLog)
	}

	if config.TmpDir != "/tmp" {
		terminal.SafePrintln("      Temp directory: " + config.TmpDir)
	}

	if config.Socket != "/var/run/mysqld/mysqld.sock" {
		terminal.SafePrintln("      Socket: " + config.Socket)
	}
}

// getAllCustomDirectories mengembalikan semua direktori custom yang perlu dihapus
func getAllCustomDirectories(config *MariaDBConfig) []string {
	dirs := make(map[string]bool) // Use map to avoid duplicates

	// Tambahkan direktori utama
	dirs[config.DataDir] = true

	if config.InnoDBDir != "" && config.InnoDBDir != config.DataDir {
		dirs[config.InnoDBDir] = true
	}

	if config.BinlogDir != "" && config.BinlogDir != config.DataDir {
		dirs[config.BinlogDir] = true
	}

	if config.LogDir != "" {
		dirs[config.LogDir] = true
	}

	if config.TmpDir != "" && config.TmpDir != "/tmp" {
		dirs[config.TmpDir] = true
	}

	// Tambahkan directory dari file log individual
	if config.ErrorLog != "" {
		dirs[filepath.Dir(config.ErrorLog)] = true
	}

	if config.SlowLog != "" {
		dirs[filepath.Dir(config.SlowLog)] = true
	}

	if config.GeneralLog != "" {
		dirs[filepath.Dir(config.GeneralLog)] = true
	}

	// Convert map to slice
	result := make([]string, 0, len(dirs))
	for dir := range dirs {
		// Resolve relative paths dan validate
		if absDir, err := filepath.Abs(dir); err == nil {
			result = append(result, absDir)
		} else {
			result = append(result, dir)
		}
	}

	return result
}

// getAllCustomFiles mengembalikan semua file custom yang perlu dihapus
func getAllCustomFiles(config *MariaDBConfig) []string {
	files := []string{}

	if config.ErrorLog != "" {
		files = append(files, config.ErrorLog)
	}

	if config.SlowLog != "" {
		files = append(files, config.SlowLog)
	}

	if config.GeneralLog != "" {
		files = append(files, config.GeneralLog)
	}

	if config.Socket != "" {
		files = append(files, config.Socket)
	}

	// Tambahkan file-file binlog (biasanya ada multiple files)
	if config.BinlogDir != "" {
		binlogPattern := filepath.Join(config.BinlogDir, "*-bin.*")
		if matches, err := filepath.Glob(binlogPattern); err == nil {
			files = append(files, matches...)
		}
	}

	return files
}

// validateDirectoryForRemoval memvalidasi apakah directory aman untuk dihapus
func validateDirectoryForRemoval(dir string) error {
	// Daftar direktori yang TIDAK BOLEH dihapus
	protectedDirs := []string{
		"/",
		"/bin",
		"/boot",
		"/dev",
		"/etc",
		"/home",
		"/lib",
		"/lib64",
		"/media",
		"/mnt",
		"/opt",
		"/proc",
		"/root",
		"/run",
		"/sbin",
		"/srv",
		"/sys",
		"/tmp",
		"/usr",
		"/var",
	}

	// Resolve absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("gagal resolve path %s: %w", dir, err)
	}

	// Cek apakah directory dalam daftar yang dilindungi
	for _, protected := range protectedDirs {
		if absDir == protected {
			return fmt.Errorf("direktori %s adalah direktori sistem yang dilindungi", absDir)
		}
	}

	// Cek apakah parent directory adalah direktori yang dilindungi (kecuali /var/lib, /var/log)
	parent := filepath.Dir(absDir)
	allowedParents := regexp.MustCompile(`^/(var/(lib|log)|opt|usr/local)`)

	if !allowedParents.MatchString(parent) {
		for _, protected := range protectedDirs {
			if parent == protected {
				return fmt.Errorf("direktori %s berada di bawah direktori sistem yang dilindungi %s", absDir, protected)
			}
		}
	}

	return nil
}
