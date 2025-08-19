package mariadb

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/disk"
	mariadb_utils "sfDBTools/utils/mariadb"
)

// ConfigOptions represents configuration options for MariaDB
// Deprecated: Use ConfigurationOptions from types.go instead
type ConfigOptions struct {
	ServerID                   string
	LogBin                     string
	DataDir                    string
	InnoDBDataHomeDir          string
	InnoDBLogGroupHomeDir      string
	LogDir                     string
	Port                       int
	EnableEncryption           bool
	FileKeyManagementFilename  string
	FileKeyManagementAlgorithm string
	ConfigFilePath             string
	SourceKeyFile              string
}

// ConfigResult represents the result of configuration operation
type ConfigResult struct {
	Success          bool
	ConfigPath       string
	BackupPath       string
	KeyFileCopied    bool
	ServiceRestarted bool
	Message          string
	Error            error
}

// MariaDBStatus represents the status of MariaDB service
type MariaDBStatus struct {
	IsRunning   bool
	ServiceName string
}

// ConfigureMariaDB handles interactive configuration of MariaDB
func ConfigureMariaDB() (*ConfigResult, error) {
	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting MariaDB configuration")

	// Step 3: Get configuration options from user
	options, err := getConfigurationOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration options: %w", err)
	}

	// Step 1: Check MariaDB status
	serviceInfo, err := mariadb_utils.GetServiceInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to check MariaDB status: %w", err)
	}

	// Step 2: Stop MariaDB if running
	if serviceInfo.Running {
		lg.Info("MariaDB is running, stopping service", logger.String("service", serviceInfo.Name))
		if err := mariadb_utils.StopService(); err != nil {
			return nil, fmt.Errorf("failed to stop MariaDB service: %w", err)
		}
		lg.Info("MariaDB service stopped successfully")
	}

	// Apply configuration
	result, err := applyConfiguration(options)
	if err != nil {
		lg.Error("Failed to apply configuration", logger.Error(err))
		return result, err
	}

	lg.Info("MariaDB configuration completed successfully")

	// LANGKAH TAMBAHAN: Eksekusi query CREATE USER dan GRANT
	if err := executeInitialUserGrants(options.Port); err != nil {
		lg.Warn("Failed to execute initial user grants", logger.Error(err))
	} else {
		lg.Info("Initial user grants executed successfully")

		// Setelah membuat user/grant, buat database kosong sesuai pola
		if err := createInitialDatabases(options.Port); err != nil {
			lg.Warn("Failed to create initial databases", logger.Error(err))
		} else {
			lg.Info("Initial databases created successfully")
		}
	}

	return result, nil
}

// executeInitialUserGrants membuat user dan grant sesuai template
func executeInitialUserGrants(port int) error {
	db, err := getRootDB(port)
	if err != nil {
		return fmt.Errorf("failed to connect to MariaDB as root: %w", err)
	}
	defer db.Close()

	// Ambil clientCode dari config.yaml
	clientCode := "demo"
	if cfg, err := config.Get(); err == nil && cfg.General.ClientCode != "" {
		clientCode = cfg.General.ClientCode
	}

	queries := []string{
		// Pengguna Administratif
		"CREATE USER IF NOT EXISTS 'papp'@'%' IDENTIFIED BY 'P@ssw0rdpapp!@#';",
		"CREATE USER IF NOT EXISTS 'sysadmin'@'%' IDENTIFIED BY 'P@ssw0rdsys!@#';",
		"CREATE USER IF NOT EXISTS 'dbaDO'@'%' IDENTIFIED BY 'DataOn24!!';",
		// Pengguna Galera SST
		"CREATE USER IF NOT EXISTS 'sst_user'@'%' IDENTIFIED BY 'P@ssw0rdsst!@#';",
		// Backup & Restore
		"CREATE USER IF NOT EXISTS 'backup_user'@'%' IDENTIFIED BY 'P@ssw0rdBackup!@#';",
		"CREATE USER IF NOT EXISTS 'restore_user'@'%' IDENTIFIED BY 'P@ssw0rdRestore!@#';",
		// MaxScale
		"CREATE USER IF NOT EXISTS 'maxscale'@'%' IDENTIFIED BY 'P@ssw0rdMaxscale!@#';",
		// Pengguna aplikasi (admin, user, fin)
		fmt.Sprintf("CREATE USER IF NOT EXISTS 'sfnbc_%s_admin'@'%%' IDENTIFIED BY 'P@ssw0rdadm!@#';", clientCode),
		fmt.Sprintf("CREATE USER IF NOT EXISTS 'sfnbc_%s_user'@'%%' IDENTIFIED BY 'P@ssw0rduser!@#';", clientCode),
		fmt.Sprintf("CREATE USER IF NOT EXISTS 'sfnbc_%s_fin'@'%%' IDENTIFIED BY 'P@ssw0rdfin!@#';", clientCode),

		// GRANT
		"GRANT ALL PRIVILEGES ON *.* TO 'papp'@'%';",
		"GRANT ALL PRIVILEGES ON *.* TO 'sysadmin'@'%';",
		"GRANT ALL PRIVILEGES ON *.* TO 'dbaDO'@'%' WITH GRANT OPTION;",
		"GRANT ALL PRIVILEGES ON *.* TO 'sst_user'@'%';",
		"GRANT SELECT, SHOW VIEW, TRIGGER, LOCK TABLES, EVENT, CREATE ROUTINE, ALTER ROUTINE, RELOAD, PROCESS, REPLICATION CLIENT ON *.* TO 'backup_user'@'%';",
		fmt.Sprintf("GRANT ALL PRIVILEGES ON `dbsf_nbc_%s_secondary_training`.* TO 'restore_user'@'%%';", clientCode),
		fmt.Sprintf("GRANT ALL PRIVILEGES ON `dbsf_nbc_%s_secondary_training_dmart`.* TO 'restore_user'@'%%';", clientCode),
		"GRANT ALL PRIVILEGES ON *.* TO 'maxscale'@'%';",
		fmt.Sprintf("GRANT ALL PRIVILEGES ON `dbsf_nbc_%s`.* TO 'sfnbc_%s_admin'@'%%', 'sfnbc_%s_user'@'%%', 'sfnbc_%s_fin'@'%%';", clientCode, clientCode, clientCode, clientCode),
		fmt.Sprintf("GRANT ALL PRIVILEGES ON `dbsf_nbc_%s_dmart`.* TO 'sfnbc_%s_admin'@'%%', 'sfnbc_%s_user'@'%%', 'sfnbc_%s_fin'@'%%';", clientCode, clientCode, clientCode, clientCode),
		fmt.Sprintf("GRANT ALL PRIVILEGES ON `dbsf_nbc_%s_temp`.* TO 'sfnbc_%s_admin'@'%%', 'sfnbc_%s_user'@'%%', 'sfnbc_%s_fin'@'%%';", clientCode, clientCode, clientCode, clientCode),
		fmt.Sprintf("GRANT ALL PRIVILEGES ON `dbsf_nbc_%s_archive`.* TO 'sfnbc_%s_admin'@'%%', 'sfnbc_%s_user'@'%%', 'sfnbc_%s_fin'@'%%';", clientCode, clientCode, clientCode, clientCode),
		fmt.Sprintf("GRANT ALL PRIVILEGES ON `dbsf_nbc_%s_secondary_training`.* TO 'sfnbc_%s_admin'@'%%', 'sfnbc_%s_user'@'%%', 'sfnbc_%s_fin'@'%%';", clientCode, clientCode, clientCode, clientCode),
		fmt.Sprintf("GRANT ALL PRIVILEGES ON `dbsf_nbc_%s_secondary_training_dmart`.* TO 'sfnbc_%s_admin'@'%%', 'sfnbc_%s_user'@'%%', 'sfnbc_%s_fin'@'%%';", clientCode, clientCode, clientCode, clientCode),
		"FLUSH PRIVILEGES;",
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("failed to execute query: %s, err: %w", q, err)
		}
	}
	return nil
}

// createInitialDatabases membuat database kosong sesuai pola yang diminta
func createInitialDatabases(port int) error {
	db, err := getRootDB(port)
	if err != nil {
		return fmt.Errorf("failed to connect to MariaDB as root: %w", err)
	}
	defer db.Close()

	clientCode := "demo"
	if cfg, err := config.Get(); err == nil && cfg.General.ClientCode != "" {
		clientCode = cfg.General.ClientCode
	}

	names := []string{
		"sfDBTools",
		fmt.Sprintf("dbsf_nbc_%s", clientCode),
		fmt.Sprintf("dbsf_nbc_%s_dmart", clientCode),
		fmt.Sprintf("dbsaas_portal_%s", clientCode),
		fmt.Sprintf("dbsf_nbc_%s_temp", clientCode),
		fmt.Sprintf("dbsf_nbc_%s_archive", clientCode),
	}

	for _, name := range names {
		q := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;", name)
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("failed to create database %s: %w", name, err)
		}
	}

	return nil
}

// getRootDB tries to connect as root via TCP then falls back to common unix sockets
func getRootDB(port int) (*sql.DB, error) {
	// Try TCP first
	tcpDSN := fmt.Sprintf("root:@tcp(127.0.0.1:%d)/mysql", port)
	db, err := sql.Open("mysql", tcpDSN)
	if err == nil {
		if pingErr := db.Ping(); pingErr == nil {
			return db, nil
		}
		_ = db.Close()
	}

	// Common socket locations to try
	sockets := []string{"/var/run/mysqld/mysqld.sock", "/var/lib/mysql/mysql.sock", "/tmp/mysql.sock"}
	for _, sock := range sockets {
		dsn := fmt.Sprintf("root:@unix(%s)/mysql", sock)
		db2, err2 := sql.Open("mysql", dsn)
		if err2 != nil {
			continue
		}
		if pingErr := db2.Ping(); pingErr == nil {
			return db2, nil
		}
		_ = db2.Close()
	}

	return nil, fmt.Errorf("unable to connect to MariaDB via TCP or common sockets")
}

// getConfigurationOptions prompts user for configuration options
func getConfigurationOptions() (*ConfigOptions, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n🔧 MariaDB Configuration Setup")
	fmt.Println("===============================")

	options := &ConfigOptions{
		SourceKeyFile: "config/key_maria_nbc.txt",
	}

	// Detect OS and set default config path
	configPath, err := detectMariaDBConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to detect MariaDB config path: %w", err)
	}
	options.ConfigFilePath = configPath

	fmt.Printf("Detected MariaDB config path: %s\n", configPath)

	// Server ID
	fmt.Print("Enter Server ID [PNM-184]: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		options.ServerID = "PNM-184"
	} else {
		options.ServerID = input
	}

	// Log Bin Path
	for {
		fmt.Print("Enter Log Bin directory [/var/lib/mysqlbinlogs]: ")
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			options.LogBin = "/var/lib/mysqlbinlogs/mysql-bin"
			break
		} else if strings.HasPrefix(input, "/home/") {
			fmt.Println("❌ Error: /home/ paths are not allowed for security reasons. Please use system directories.")
			continue
		} else {
			// Add mysql-bin filename to the directory path
			options.LogBin = filepath.Join(input, "mysql-bin")
			break
		}
	}

	// Data Directory
	for {
		fmt.Print("Enter Data directory [/var/lib/mysql]: ")
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			options.DataDir = "/var/lib/mysql"
			break
		} else if strings.HasPrefix(input, "/home/") {
			fmt.Println("❌ Error: /home/ paths are not allowed for security reasons. Please use system directories.")
			continue
		} else {
			options.DataDir = input
			break
		}
	}

	// InnoDB Data Home Directory (automatically same as Data Directory)
	options.InnoDBDataHomeDir = options.DataDir
	fmt.Printf("InnoDB Data Home directory: %s (auto-configured same as Data directory)\n", options.InnoDBDataHomeDir)

	// InnoDB Log Group Home Directory (automatically same as Data Directory)
	options.InnoDBLogGroupHomeDir = options.DataDir
	fmt.Printf("InnoDB Log Group Home directory: %s (auto-configured same as Data directory)\n", options.InnoDBLogGroupHomeDir)

	// Log Directory (automatically same as Data Directory)
	options.LogDir = options.DataDir
	fmt.Printf("Log directory: %s (auto-configured same as Data directory)\n", options.LogDir)

	// Port
	fmt.Print("Enter Port [43306]: ")
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		options.Port = 43306
	} else {
		port, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Invalid port number, using default 43306\n")
			options.Port = 43306
		} else {
			options.Port = port
		}
	}

	// Enable Encryption
	fmt.Print("Enable encryption? (y/N): ")
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	options.EnableEncryption = (input == "y" || input == "yes")

	if options.EnableEncryption {
		// Key file path
		for {
			fmt.Print("Enter encryption key file path [/var/lib/mysql/key_maria_nbc.txt]: ")
			input, _ = reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input == "" {
				options.FileKeyManagementFilename = "/var/lib/mysql/key_maria_nbc.txt"
				break
			} else if strings.HasPrefix(input, "/home/") {
				fmt.Println("❌ Error: /home/ paths are not allowed for security reasons. Please use system directories.")
				continue
			} else {
				options.FileKeyManagementFilename = input
				break
			}
		}

		// Encryption algorithm (automatically set to AES_CTR)
		options.FileKeyManagementAlgorithm = "AES_CTR"
		fmt.Printf("Encryption algorithm: %s (default)\n", options.FileKeyManagementAlgorithm)
	}

	return options, nil
}

// applyConfiguration applies the configuration to MariaDB
func applyConfiguration(options *ConfigOptions) (*ConfigResult, error) {
	lg, _ := logger.Get()

	result := &ConfigResult{
		ConfigPath: options.ConfigFilePath,
	}

	// Create backup of existing configuration
	backupPath, err := createConfigBackup(options.ConfigFilePath)
	if err != nil {
		result.Error = fmt.Errorf("failed to create config backup: %w", err)
		return result, result.Error
	}
	result.BackupPath = backupPath

	if backupPath != "" {
		lg.Info("Created configuration backup", logger.String("backup_path", backupPath))
	}

	// Create necessary directories
	if err := createRequiredDirectories(options); err != nil {
		result.Error = fmt.Errorf("failed to create directories: %w", err)
		return result, result.Error
	}

	// Copy encryption key if enabled
	if options.EnableEncryption {
		if err := copyEncryptionKey(options); err != nil {
			result.Error = fmt.Errorf("failed to copy encryption key: %w", err)
			return result, result.Error
		}
		result.KeyFileCopied = true
		lg.Info("Encryption key copied successfully")
	}

	// Step 4: Generate new configuration file (copy server.cnf and replace values)
	if err := generateConfigFile(options); err != nil {
		result.Error = fmt.Errorf("failed to generate config file: %w", err)
		return result, result.Error
	}

	// Migrate existing data if custom data directory is specified and different from default
	if options.DataDir != "/var/lib/mysql" && options.DataDir != "" {
		if err := migrateMariaDBData(options.DataDir); err != nil {
			result.Error = fmt.Errorf("failed to migrate MariaDB data: %w", err)
			return result, result.Error
		}
	}

	// Step 5: Set custom directory permissions
	if err := setCustomDirectoryPermissions(options); err != nil {
		lg.Warn("Failed to set custom directory permissions", logger.Error(err))
	}

	// Set proper permissions for config file
	if err := setConfigPermissions(options.ConfigFilePath); err != nil {
		lg.Warn("Failed to set config permissions", logger.Error(err))
	}

	// Initialize MariaDB database if needed
	if err := initializeMariaDBDatabase(options); err != nil {
		lg.Warn("Failed to initialize MariaDB database", logger.Error(err))
		// Continue anyway, as it might already be initialized
	}

	// Start MariaDB service (using existing utility would require modification)
	if err := startMariaDBService(); err != nil {
		result.Error = fmt.Errorf("failed to start MariaDB service: %w", err)
		return result, result.Error
	}
	result.ServiceRestarted = true

	result.Success = true
	result.Message = "MariaDB configuration applied successfully"

	lg.Info("Configuration applied successfully",
		logger.String("config_path", options.ConfigFilePath),
		logger.Bool("encryption_enabled", options.EnableEncryption))

	// LANGKAH TAMBAHAN: Eksekusi query CREATE USER dan GRANT
	return result, nil
}

// createConfigBackup creates a backup of the existing configuration
func createConfigBackup(configPath string) (string, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, no backup needed
		return "", nil
	}

	backupPath := configPath + ".backup." + fmt.Sprintf("%d", os.Getpid())

	sourceFile, err := os.Open(configPath)
	if err != nil {
		return "", err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return backupPath, err
}

// createRequiredDirectories creates all required directories using existing utilities
func createRequiredDirectories(options *ConfigOptions) error {
	lg, _ := logger.Get()

	directories := []string{
		options.DataDir,
		options.LogDir,
		options.InnoDBDataHomeDir,
		options.InnoDBLogGroupHomeDir,
		filepath.Dir(options.LogBin),
		filepath.Dir(options.ConfigFilePath),
	}

	for _, dir := range directories {
		if dir == "" {
			continue
		}

		// Check if directory already exists
		exists := false
		if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
			exists = true
			lg.Debug("Directory already exists", logger.String("directory", dir))
		}

		// Use existing disk utility to create directory
		if err := disk.CreateOutputDirectory(dir); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// Only set permissions for newly created directories or non-system directories
		// Skip setting permissions for standard system directories that may already have proper ownership
		if !exists || !isSystemDirectory(dir) {
			if err := setDirectoryOwnershipSafely(dir, "mysql", "mysql"); err != nil {
				lg.Warn("Failed to set MySQL ownership", logger.String("directory", dir), logger.Error(err))
			} else {
				lg.Debug("Set ownership for directory", logger.String("directory", dir))
			}
		}

		lg.Info("Created directory with permissions", logger.String("path", dir))
	}

	return nil
}

// isSystemDirectory checks if a directory is a standard system directory that may already have proper ownership
func isSystemDirectory(dir string) bool {
	systemDirs := []string{
		"/var/lib/mysql",
		"/etc/mysql",
		"/etc/my.cnf.d",
	}

	for _, sysDir := range systemDirs {
		if dir == sysDir {
			return true
		}
	}
	return false
}

// setDirectoryOwnershipSafely sets directory ownership without recursive operations on system directories
func setDirectoryOwnershipSafely(dir, owner, group string) error {
	lg, _ := logger.Get()

	// For system directories, only set ownership on the directory itself, not recursively
	// and ensure proper read permissions
	if isSystemDirectory(dir) {
		// Set proper permissions first
		if err := os.Chmod(dir, 0755); err != nil {
			lg.Debug("Failed to set permissions on system directory", logger.String("directory", dir), logger.Error(err))
		}

		// Then set ownership (non-recursive)
		cmd := exec.Command("chown", owner+":"+group, dir)
		if err := cmd.Run(); err != nil {
			lg.Debug("Failed to set ownership on system directory", logger.String("directory", dir), logger.Error(err))
			return err
		}
		lg.Debug("Set ownership on system directory", logger.String("directory", dir))
		return nil
	}

	// For non-system directories, use the existing common utility
	return common.SetDirectoryPermissions(dir, 0755, owner, group)
}

// copyEncryptionKey copies the encryption key file
func copyEncryptionKey(options *ConfigOptions) error {
	lg, _ := logger.Get()

	sourceFile := options.SourceKeyFile
	destFile := options.FileKeyManagementFilename

	// Check if source file exists
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		return fmt.Errorf("source key file does not exist: %s", sourceFile)
	}

	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(destFile)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create key file directory: %w", err)
	}

	// Copy the file
	source, err := os.Open(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to open source key file: %w", err)
	}
	defer source.Close()

	dest, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("failed to create destination key file: %w", err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, source); err != nil {
		return fmt.Errorf("failed to copy key file: %w", err)
	}

	// Set proper permissions using common utility
	if err := common.SetFilePermissions(destFile, 0600, "mysql", "mysql"); err != nil {
		lg.Warn("Failed to set key file permissions", logger.Error(err))
	}

	lg.Info("Encryption key copied",
		logger.String("from", sourceFile),
		logger.String("to", destFile))

	return nil
}

// generateConfigFile generates the MariaDB configuration file
func generateConfigFile(options *ConfigOptions) error {
	lg, _ := logger.Get()

	// Read the template server.cnf
	templatePath := "config/server.cnf"
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read config template: %w", err)
	}

	configContent := string(content)

	// Replace configuration values
	configContent = replaceConfigValue(configContent, "server_id", options.ServerID)
	configContent = replaceConfigValue(configContent, "log_bin", options.LogBin)
	configContent = replaceConfigValue(configContent, "datadir", options.DataDir)
	configContent = replaceConfigValue(configContent, "innodb_data_home_dir", options.InnoDBDataHomeDir)
	configContent = replaceConfigValue(configContent, "innodb_log_group_home_dir", options.InnoDBLogGroupHomeDir)
	configContent = replaceConfigValue(configContent, "port", fmt.Sprintf("%d", options.Port))

	// Update log file paths
	logErrorPath := filepath.Join(options.LogDir, "mysql_error.log")
	slowLogPath := filepath.Join(options.LogDir, "mysql_slow.log")
	configContent = replaceConfigValue(configContent, "log_error", logErrorPath)
	configContent = replaceConfigValue(configContent, "slow_query_log_file", slowLogPath)

	// Handle encryption settings
	if options.EnableEncryption {
		configContent = replaceConfigValue(configContent, "file_key_management_filename", options.FileKeyManagementFilename)
		configContent = replaceConfigValue(configContent, "file_key_management_encryption_algorithm", options.FileKeyManagementAlgorithm)
		configContent = replaceConfigValue(configContent, "innodb-encrypt-tables", "ON")
	} else {
		// Remove or comment out encryption-related lines
		configContent = commentOutConfig(configContent, "plugin-load-add")
		configContent = commentOutConfig(configContent, "file_key_management_encryption_algorithm")
		configContent = commentOutConfig(configContent, "file_key_management_filename")
		configContent = commentOutConfig(configContent, "innodb-encrypt-tables")
	}

	// Write the new configuration file
	if err := os.WriteFile(options.ConfigFilePath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	lg.Info("Configuration file generated", logger.String("path", options.ConfigFilePath))
	return nil
}

// replaceConfigValue replaces a configuration value in the config content
func replaceConfigValue(content, key, value string) string {
	// Pattern to match key = value or key = "value"
	pattern := fmt.Sprintf(`(?m)^(\s*%s\s*=\s*)[^\n]*`, regexp.QuoteMeta(key))
	replacement := fmt.Sprintf("${1}%s", value)

	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(content, replacement)
}

// commentOutConfig comments out a configuration line
func commentOutConfig(content, key string) string {
	pattern := fmt.Sprintf(`(?m)^(\s*)(%s.*?)$`, regexp.QuoteMeta(key))
	replacement := "${1}#${2}"

	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(content, replacement)
}

// setConfigPermissions sets proper permissions for the config file using common utility
func setConfigPermissions(configPath string) error {
	return common.SetFilePermissions(configPath, 0644, "root", "root")
}

// DisplayConfigResult displays the configuration result to the user
func DisplayConfigResult(result *ConfigResult) {
	if result.Success {
		fmt.Println("\n✅ MariaDB Configuration Completed Successfully!")
		fmt.Println("================================================")
		fmt.Printf("📁 Config file: %s\n", result.ConfigPath)

		if result.BackupPath != "" {
			fmt.Printf("💾 Backup created: %s\n", result.BackupPath)
		}

		if result.KeyFileCopied {
			fmt.Println("🔐 Encryption key copied successfully")
		}

		if result.ServiceRestarted {
			fmt.Println("🔄 MariaDB service restarted")
		}

		fmt.Printf("✨ %s\n", result.Message)
	} else {
		fmt.Println("\n❌ MariaDB Configuration Failed!")
		fmt.Println("=================================")

		if result.Error != nil {
			fmt.Printf("Error: %v\n", result.Error)
		}

		if result.BackupPath != "" {
			fmt.Printf("💾 Backup available: %s\n", result.BackupPath)
			fmt.Println("You can restore the backup if needed")
		}
	}
}

// detectMariaDBConfigPath detects the MariaDB config path based on OS
func detectMariaDBConfigPath() (string, error) {
	lg, _ := logger.Get()

	lg.Info("Using OS-based config path detection")
	configPath := getOSBasedConfigPath()
	lg.Info("Detected config path", logger.String("path", configPath))

	return configPath, nil
}

// getOSBasedConfigPath returns config path based on OS distribution
func getOSBasedConfigPath() string {
	// Check for Ubuntu/Debian - prioritize mariadb.conf.d
	if _, err := os.Stat("/etc/mysql/mariadb.conf.d/50-server.cnf"); err == nil {
		return "/etc/mysql/mariadb.conf.d/50-server.cnf"
	}

	// Check for RHEL/CentOS/Fedora - prioritize my.cnf.d
	if _, err := os.Stat("/etc/my.cnf.d/50-server.cnf"); err == nil {
		return "/etc/my.cnf.d/50-server.cnf"
	}
	if _, err := os.Stat("/etc/my.cnf.d/server.cnf"); err == nil {
		return "/etc/my.cnf.d/server.cnf"
	}

	// Check for main config files as fallback
	if _, err := os.Stat("/etc/my.cnf"); err == nil {
		return "/etc/my.cnf"
	}
	if _, err := os.Stat("/etc/mysql/my.cnf"); err == nil {
		return "/etc/mysql/my.cnf"
	}

	// Default fallback for Debian/Ubuntu
	return "/etc/mysql/mariadb.conf.d/50-server.cnf"
}

// migrateMariaDBData migrates existing MariaDB data to new directory using rsync
func migrateMariaDBData(customDataDir string) error {
	lg, _ := logger.Get()

	defaultDataDir := "/var/lib/mysql"

	// Check if default data directory exists and has data
	if _, err := os.Stat(defaultDataDir); os.IsNotExist(err) {
		lg.Info("No existing data directory to migrate")
		return nil
	}

	// Check if custom directory already has data
	if _, err := os.Stat(filepath.Join(customDataDir, "mysql")); err == nil {
		lg.Info("Custom data directory already has data, skipping migration")
		return nil
	}

	lg.Info("Migrating MariaDB data using rsync",
		logger.String("from", defaultDataDir),
		logger.String("to", customDataDir))

	// Use rsync to copy data
	cmd := exec.Command("rsync", "-av", defaultDataDir+"/", customDataDir+"/")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to migrate data with rsync: %w\nOutput: %s", err, string(output))
	}

	lg.Info("Data migration completed successfully", logger.String("output", string(output)))
	return nil
}

// setCustomDirectoryPermissions sets proper permissions for custom directories using common utility
func setCustomDirectoryPermissions(options *ConfigOptions) error {
	lg, _ := logger.Get()

	directories := []string{
		options.DataDir,
		options.LogDir,
		options.InnoDBDataHomeDir,
		options.InnoDBLogGroupHomeDir,
		filepath.Dir(options.LogBin),
	}

	// Remove duplicates
	seen := make(map[string]bool)
	uniqueDirs := []string{}
	for _, dir := range directories {
		if dir == "" || seen[dir] {
			continue
		}
		seen[dir] = true
		uniqueDirs = append(uniqueDirs, dir)
	}

	for _, dir := range uniqueDirs {
		// Check if directory exists first
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			lg.Debug("Directory does not exist, skipping permission setting", logger.String("directory", dir))
			continue
		}

		// Use safe ownership setting for all directories
		if err := setDirectoryOwnershipSafely(dir, "mysql", "mysql"); err != nil {
			lg.Warn("Failed to set directory permissions with safe method", logger.String("directory", dir), logger.Error(err))

			// Only try aggressive methods for non-system directories
			if !isSystemDirectory(dir) {
				cmd := exec.Command("chown", "-R", "mysql:mysql", dir)
				if chownErr := cmd.Run(); chownErr != nil {
					lg.Warn("Failed to set ownership", logger.String("directory", dir), logger.Error(chownErr))
				} else {
					lg.Info("Set ownership successfully", logger.String("directory", dir))
				}

				cmd = exec.Command("chmod", "-R", "755", dir)
				if chmodErr := cmd.Run(); chmodErr != nil {
					lg.Warn("Failed to set permissions", logger.String("directory", dir), logger.Error(chmodErr))
				} else {
					lg.Info("Set permissions successfully", logger.String("directory", dir))
				}
			}
		} else {
			lg.Info("Set permissions for custom directory", logger.String("path", dir))
		}

		// For data directory, ensure database files have proper permissions (only for non-system directories)
		if dir == options.DataDir && !isSystemDirectory(dir) {
			lg.Info("Setting proper permissions for database files", logger.String("data_dir", dir))

			// Set file permissions for database files
			cmd := exec.Command("find", dir, "-type", "f", "-exec", "chmod", "660", "{}", "+")
			if err := cmd.Run(); err != nil {
				lg.Warn("Failed to set file permissions", logger.String("directory", dir), logger.Error(err))
			} else {
				lg.Info("Set file permissions successfully using find", logger.String("directory", dir))
			}

			// Set directory permissions for subdirectories
			cmd = exec.Command("find", dir, "-type", "d", "-exec", "chmod", "770", "{}", "+")
			if err := cmd.Run(); err != nil {
				lg.Warn("Failed to set subdirectory permissions", logger.String("directory", dir), logger.Error(err))
			} else {
				lg.Info("Set subdirectory permissions successfully", logger.String("directory", dir))
			}
		} else if dir == options.DataDir && isSystemDirectory(dir) {
			lg.Info("Skipping recursive permission setting for system data directory", logger.String("data_dir", dir))
		}
	}

	return nil
}

// initializeMariaDBDatabase initializes MariaDB database if not already initialized
func initializeMariaDBDatabase(options *ConfigOptions) error {
	lg, _ := logger.Get()

	// Check if data directory is already initialized
	mysqlSystemDir := filepath.Join(options.DataDir, "mysql")
	if _, err := os.Stat(mysqlSystemDir); err == nil {
		lg.Info("Database already initialized", logger.String("data_dir", options.DataDir))
		return nil
	}

	lg.Info("Initializing MariaDB database", logger.String("data_dir", options.DataDir))

	// Try mysql_install_db first (older versions)
	if _, err := exec.LookPath("mysql_install_db"); err == nil {
		cmd := exec.Command("mysql_install_db",
			"--datadir="+options.DataDir,
			"--user=mysql")
		output, err := cmd.CombinedOutput()
		if err != nil {
			lg.Warn("mysql_install_db failed, trying mariadb-install-db",
				logger.Error(err),
				logger.String("output", string(output)))
		} else {
			lg.Info("Database initialized successfully with mysql_install_db")
			return nil
		}
	}

	// Try mariadb-install-db (newer versions)
	if _, err := exec.LookPath("mariadb-install-db"); err == nil {
		cmd := exec.Command("mariadb-install-db",
			"--datadir="+options.DataDir,
			"--user=mysql")
		output, err := cmd.CombinedOutput()
		if err != nil {
			lg.Warn("mariadb-install-db failed",
				logger.Error(err),
				logger.String("output", string(output)))
			return fmt.Errorf("failed to initialize database: %w", err)
		} else {
			lg.Info("Database initialized successfully with mariadb-install-db")
			return nil
		}
	}

	// If neither command is available
	return fmt.Errorf("neither mysql_install_db nor mariadb-install-db found")
}

// startMariaDBService starts the MariaDB service
func startMariaDBService() error {
	lg, _ := logger.Get()

	lg.Info("Starting MariaDB service")

	serviceNames := []string{"mariadb", "mysql", "mysqld"}
	var lastError error

	for _, serviceName := range serviceNames {
		// Check if service exists
		cmd := exec.Command("systemctl", "is-enabled", serviceName)
		if err := cmd.Run(); err != nil {
			lg.Debug("Service not found or not enabled", logger.String("service", serviceName))
			continue // Service doesn't exist
		}

		lg.Info("Found service, attempting to start", logger.String("service", serviceName))

		// Start the service
		cmd = exec.Command("systemctl", "start", serviceName)
		output, err := cmd.CombinedOutput()
		if err != nil {
			lastError = err
			lg.Warn("Failed to start service",
				logger.String("service", serviceName),
				logger.Error(err),
				logger.String("output", string(output)))

			// Check service status for more details
			statusCmd := exec.Command("systemctl", "status", serviceName)
			statusOutput, _ := statusCmd.CombinedOutput()
			lg.Warn("Service status",
				logger.String("service", serviceName),
				logger.String("status_output", string(statusOutput)))
			continue
		}

		// Enable the service to start on boot
		cmd = exec.Command("systemctl", "enable", serviceName)
		if err := cmd.Run(); err != nil {
			lg.Warn("Failed to enable service", logger.String("service", serviceName), logger.Error(err))
		}

		lg.Info("MariaDB service started successfully", logger.String("service", serviceName))
		return nil
	}

	if lastError != nil {
		return fmt.Errorf("failed to start MariaDB service - last error: %w", lastError)
	}
	return fmt.Errorf("failed to start MariaDB service - no valid service found")
}
