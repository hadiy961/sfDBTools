package remove

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
)

// DetectionService handles detection of existing MariaDB installations
type DetectionService struct {
	osInfo *common.OSInfo
}

// NewDetectionService creates a new detection service
func NewDetectionService(osInfo *common.OSInfo) *DetectionService {
	return &DetectionService{
		osInfo: osInfo,
	}
}

// DetectInstallation detects if MariaDB is installed and gathers information
func (d *DetectionService) DetectInstallation() (*DetectedInstallation, error) {
	lg, _ := logger.Get()

	lg.Info("Starting MariaDB installation detection")

	installation := &DetectedInstallation{
		OSInfo: d.osInfo,
	}

	// Detect installed packages
	if err := d.detectPackages(installation); err != nil {
		lg.Warn("Failed to detect packages", logger.Error(err))
	}

	// Detect service status
	if err := d.detectServiceStatus(installation); err != nil {
		lg.Warn("Failed to detect service status", logger.Error(err))
	}

	// Detect data directories and files
	if err := d.detectDataDirectories(installation); err != nil {
		lg.Warn("Failed to detect data directories", logger.Error(err))
	}

	// Detect configuration files
	if err := d.detectConfigFiles(installation); err != nil {
		lg.Warn("Failed to detect config files", logger.Error(err))
	}

	// Detect log files
	if err := d.detectLogFiles(installation); err != nil {
		lg.Warn("Failed to detect log files", logger.Error(err))
	}

	lg.Info("Installation detection completed",
		logger.Bool("is_installed", installation.IsInstalled),
		logger.String("version", installation.Version))

	return installation, nil
}

// detectPackages detects installed MariaDB packages
func (d *DetectionService) detectPackages(installation *DetectedInstallation) error {
	lg, _ := logger.Get()

	var cmd *exec.Cmd
	var packageNames []string

	switch d.osInfo.PackageType {
	case "rpm":
		// Check for RPM packages
		packageNames = []string{"MariaDB-server", "mariadb-server", "mysql-server"}
		cmd = exec.Command("rpm", "-qa")
	case "deb":
		// Check for DEB packages
		packageNames = []string{"mariadb-server", "mariadb-server-10.*", "mysql-server"}
		cmd = exec.Command("dpkg", "-l")
	default:
		return fmt.Errorf("unsupported package type: %s", d.osInfo.PackageType)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}

	outputStr := string(output)

	// Check for MariaDB packages
	for _, packageName := range packageNames {
		if strings.Contains(strings.ToLower(outputStr), strings.ToLower(packageName)) {
			installation.IsInstalled = true
			installation.PackageName = packageName

			// Try to extract version
			if version := d.extractVersionFromPackage(outputStr, packageName); version != "" {
				installation.Version = version
			}

			lg.Info("Found MariaDB package",
				logger.String("package", packageName),
				logger.String("version", installation.Version))
			break
		}
	}

	return nil
}

// extractVersionFromPackage extracts version from package listing output
func (d *DetectionService) extractVersionFromPackage(output, packageName string) string {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.ToLower(line)
		if strings.Contains(line, strings.ToLower(packageName)) {
			// Try to extract version number (look for patterns like 10.11.14)
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.Contains(part, ".") && len(part) > 3 {
					// Basic version pattern check
					if strings.Count(part, ".") >= 1 {
						return part
					}
				}
			}
		}
	}

	return ""
}

// detectServiceStatus detects MariaDB service status
func (d *DetectionService) detectServiceStatus(installation *DetectedInstallation) error {
	serviceNames := []string{"mariadb", "mysql", "mysqld"}

	for _, serviceName := range serviceNames {
		// Check if service exists and its status
		if d.serviceExists(serviceName) {
			installation.ServiceName = serviceName
			installation.ServiceActive = d.isServiceActive(serviceName)
			installation.ServiceEnabled = d.isServiceEnabled(serviceName)
			break
		}
	}

	return nil
}

// serviceExists checks if a systemd service exists
func (d *DetectionService) serviceExists(serviceName string) bool {
	cmd := exec.Command("systemctl", "list-unit-files", serviceName+".service")
	output, err := cmd.CombinedOutput()

	return err == nil && strings.Contains(string(output), serviceName+".service")
}

// isServiceActive checks if a service is currently running
func (d *DetectionService) isServiceActive(serviceName string) bool {
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.CombinedOutput()

	return err == nil && strings.TrimSpace(string(output)) == "active"
}

// isServiceEnabled checks if a service is enabled on boot
func (d *DetectionService) isServiceEnabled(serviceName string) bool {
	cmd := exec.Command("systemctl", "is-enabled", serviceName)
	output, err := cmd.CombinedOutput()

	return err == nil && strings.TrimSpace(string(output)) == "enabled"
}

// detectDataDirectories detects MariaDB data directories from actual configuration
func (d *DetectionService) detectDataDirectories(installation *DetectedInstallation) error {
	lg, _ := logger.Get()

	// First try to get actual directories from MariaDB configuration
	actualDirs, err := d.getActualMariaDBDirectories()
	if err != nil {
		lg.Warn("Failed to get actual directories from MariaDB config, using defaults", logger.Error(err))
		// Fallback to default paths
		dataPaths := []string{
			"/var/lib/mysql",
			"/var/lib/mariadb",
			"/opt/mariadb/data",
			"/data/mysql", // Custom path from configure
		}

		for _, path := range dataPaths {
			if stat, err := os.Stat(path); err == nil && stat.IsDir() {
				installation.DataDirectoryExists = true

				// Calculate directory size
				if size, err := d.calculateDirectorySize(path); err == nil {
					installation.DataDirectorySize = size
				}
				break
			}
		}
	} else {
		// Use actual detected directories
		if actualDirs.DataDir != "" {
			if stat, err := os.Stat(actualDirs.DataDir); err == nil && stat.IsDir() {
				installation.DataDirectoryExists = true
				installation.ActualDataDir = actualDirs.DataDir

				if size, err := d.calculateDirectorySize(actualDirs.DataDir); err == nil {
					installation.DataDirectorySize = size
				}
			}
		}

		if actualDirs.BinlogDir != "" {
			installation.ActualBinlogDir = actualDirs.BinlogDir
		}

		if actualDirs.LogDir != "" {
			installation.ActualLogDir = actualDirs.LogDir
		}

		lg.Info("Detected actual MariaDB directories",
			logger.String("datadir", actualDirs.DataDir),
			logger.String("binlogdir", actualDirs.BinlogDir),
			logger.String("logdir", actualDirs.LogDir))
	}

	return nil
}

// MariaDBDirectories holds actual directory paths from configuration
type MariaDBDirectories struct {
	DataDir   string
	BinlogDir string
	LogDir    string
}

// getActualMariaDBDirectories reads actual directory paths from MariaDB configuration
func (d *DetectionService) getActualMariaDBDirectories() (*MariaDBDirectories, error) {
	dirs := &MariaDBDirectories{}

	// Method 1: Try to get from running MariaDB instance
	if runningDirs, err := d.getDirectoriesFromRunningMariaDB(); err == nil {
		return runningDirs, nil
	}

	// Method 2: Try to parse from configuration files
	if configDirs, err := d.getDirectoriesFromConfigFiles(); err == nil {
		return configDirs, nil
	}

	return dirs, fmt.Errorf("could not determine actual MariaDB directories")
}

// getDirectoriesFromRunningMariaDB gets directories from running MariaDB instance
func (d *DetectionService) getDirectoriesFromRunningMariaDB() (*MariaDBDirectories, error) {
	dirs := &MariaDBDirectories{}

	// Check if MariaDB is running
	if !d.isServiceActive("mariadb") {
		return dirs, fmt.Errorf("MariaDB service is not running")
	}

	// Try to connect and get variables
	queries := map[string]*string{
		"SELECT @@datadir":          &dirs.DataDir,
		"SELECT @@log_bin_basename": &dirs.BinlogDir,
		"SELECT @@log_error":        &dirs.LogDir,
	}

	for query, target := range queries {
		cmd := exec.Command("mariadb", "-e", query, "-s", "-N")
		if output, err := cmd.CombinedOutput(); err == nil {
			result := strings.TrimSpace(string(output))
			if result != "" {
				if target == &dirs.BinlogDir {
					// Extract directory from binlog path
					*target = filepath.Dir(result)
				} else if target == &dirs.LogDir {
					// Extract directory from log file path
					*target = filepath.Dir(result)
				} else {
					*target = result
				}
			}
		}
	}

	if dirs.DataDir == "" {
		return dirs, fmt.Errorf("could not get datadir from running MariaDB")
	}

	return dirs, nil
}

// getDirectoriesFromConfigFiles parses directories from MariaDB config files
func (d *DetectionService) getDirectoriesFromConfigFiles() (*MariaDBDirectories, error) {
	dirs := &MariaDBDirectories{}

	configFiles := []string{
		"/etc/my.cnf.d/server.cnf",
		"/etc/mysql/mariadb.conf.d/50-server.cnf",
		"/etc/my.cnf",
		"/etc/mysql/my.cnf",
	}

	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); err == nil {
			if err := d.parseConfigFile(configFile, dirs); err == nil {
				if dirs.DataDir != "" {
					return dirs, nil
				}
			}
		}
	}

	return dirs, fmt.Errorf("could not parse directories from config files")
}

// parseConfigFile parses a MariaDB configuration file for directory settings
func (d *DetectionService) parseConfigFile(configFile string, dirs *MariaDBDirectories) error {
	content, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		if strings.HasPrefix(line, "datadir") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				dirs.DataDir = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "log_bin") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				binlogPath := strings.TrimSpace(parts[1])
				dirs.BinlogDir = filepath.Dir(binlogPath)
			}
		} else if strings.HasPrefix(line, "log_error") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				logPath := strings.TrimSpace(parts[1])
				dirs.LogDir = filepath.Dir(logPath)
			}
		}
	}

	return nil
}

// calculateDirectorySize calculates the total size of a directory
func (d *DetectionService) calculateDirectorySize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// detectConfigFiles detects MariaDB configuration files
func (d *DetectionService) detectConfigFiles(installation *DetectedInstallation) error {
	lg, _ := logger.Get()

	// First try to get from actual config locations that are being used
	actualConfigFiles := []string{}

	// Add the actual config file being used (if we can detect it)
	if _, err := d.getActualMariaDBDirectories(); err == nil {
		// These are the files we actually parsed, so they exist
		configFiles := []string{
			"/etc/my.cnf.d/server.cnf",
			"/etc/mysql/mariadb.conf.d/50-server.cnf",
			"/etc/my.cnf",
			"/etc/mysql/my.cnf",
		}

		for _, configFile := range configFiles {
			if _, err := os.Stat(configFile); err == nil {
				actualConfigFiles = append(actualConfigFiles, configFile)
			}
		}

		lg.Info("Detected config files from actual configuration",
			logger.Int("count", len(actualConfigFiles)))
	}

	// Fallback to default paths if nothing detected
	if len(actualConfigFiles) == 0 {
		lg.Info("No actual config files detected, using default paths")
		configPaths := []string{
			"/etc/mysql",
			"/etc/my.cnf",
			"/etc/mysql/my.cnf",
			"/etc/mariadb",
			"/usr/local/etc/my.cnf",
			"/etc/my.cnf.d", // Directory with config files
		}

		for _, path := range configPaths {
			if _, err := os.Stat(path); err == nil {
				actualConfigFiles = append(actualConfigFiles, path)
			}
		}
	}

	installation.ConfigFiles = actualConfigFiles
	return nil
}

// detectLogFiles detects MariaDB log files
func (d *DetectionService) detectLogFiles(installation *DetectedInstallation) error {
	lg, _ := logger.Get()

	actualLogFiles := []string{}

	// First try to get actual log directory from detected configuration
	if installation.ActualLogDir != "" {
		if _, err := os.Stat(installation.ActualLogDir); err == nil {
			actualLogFiles = append(actualLogFiles, installation.ActualLogDir)
			lg.Info("Using detected log directory", logger.String("path", installation.ActualLogDir))
		}
	}

	// Try to get log files from running MariaDB
	if d.isServiceActive("mariadb") {
		cmd := exec.Command("mariadb", "-e", "SELECT @@log_error", "-s", "-N")
		if output, err := cmd.CombinedOutput(); err == nil {
			logFile := strings.TrimSpace(string(output))
			if logFile != "" && logFile != "stderr" {
				if _, err := os.Stat(logFile); err == nil {
					actualLogFiles = append(actualLogFiles, logFile)
					lg.Info("Detected actual log file from MariaDB", logger.String("file", logFile))
				}
			}
		}
	}

	// Fallback to common log locations if nothing detected
	if len(actualLogFiles) == 0 {
		lg.Info("No actual log files detected, using default paths")
		logPaths := []string{
			"/var/log/mysql",
			"/var/log/mariadb",
			"/var/log/mysqld.log",
			"/var/log/mysql.log",
			"/var/log/mysql/error.log",
			"/var/log/mariadb/mariadb.log",
		}

		for _, path := range logPaths {
			if _, err := os.Stat(path); err == nil {
				actualLogFiles = append(actualLogFiles, path)
			}
		}
	}

	installation.LogFiles = actualLogFiles
	return nil
}

// GetInstallationSummary returns a human-readable summary of the detected installation
func (d *DetectionService) GetInstallationSummary(installation *DetectedInstallation) string {
	if !installation.IsInstalled {
		return "No MariaDB installation detected"
	}

	var summary strings.Builder

	summary.WriteString("MariaDB Installation Detected:\n")
	summary.WriteString(fmt.Sprintf("  Package: %s\n", installation.PackageName))
	summary.WriteString(fmt.Sprintf("  Version: %s\n", installation.Version))

	if installation.ServiceName != "" {
		summary.WriteString(fmt.Sprintf("  Service: %s", installation.ServiceName))
		if installation.ServiceActive {
			summary.WriteString(" (running)")
		} else {
			summary.WriteString(" (stopped)")
		}
		if installation.ServiceEnabled {
			summary.WriteString(" (enabled on boot)")
		}
		summary.WriteString("\n")
	}

	if installation.DataDirectoryExists {
		summary.WriteString("  Data Directory: exists")
		if installation.DataDirectorySize > 0 {
			summary.WriteString(fmt.Sprintf(" (%s)", d.formatSize(installation.DataDirectorySize)))
		}
		summary.WriteString("\n")
	}

	if len(installation.ConfigFiles) > 0 {
		summary.WriteString(fmt.Sprintf("  Config Files: %d found\n", len(installation.ConfigFiles)))
	}

	if len(installation.LogFiles) > 0 {
		summary.WriteString(fmt.Sprintf("  Log Files: %d found\n", len(installation.LogFiles)))
	}

	return summary.String()
}

// formatSize formats a byte size into human-readable format
func (d *DetectionService) formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
