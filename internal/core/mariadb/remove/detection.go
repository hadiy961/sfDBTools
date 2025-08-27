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

// detectDataDirectories detects MariaDB data directories
func (d *DetectionService) detectDataDirectories(installation *DetectedInstallation) error {
	dataPaths := []string{
		"/var/lib/mysql",
		"/var/lib/mariadb",
		"/opt/mariadb/data",
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
	configPaths := []string{
		"/etc/mysql",
		"/etc/my.cnf",
		"/etc/mysql/my.cnf",
		"/etc/mariadb",
		"/usr/local/etc/my.cnf",
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			installation.ConfigFiles = append(installation.ConfigFiles, path)
		}
	}

	return nil
}

// detectLogFiles detects MariaDB log files
func (d *DetectionService) detectLogFiles(installation *DetectedInstallation) error {
	logPaths := []string{
		"/var/log/mysql",
		"/var/log/mariadb",
		"/var/log/mysqld.log",
		"/var/log/mysql.log",
	}

	for _, path := range logPaths {
		if _, err := os.Stat(path); err == nil {
			installation.LogFiles = append(installation.LogFiles, path)
		}
	}

	return nil
}

// GetInstallationSummary returns a human-readable summary of the detected installation
func (d *DetectionService) GetInstallationSummary(installation *DetectedInstallation) string {
	if !installation.IsInstalled {
		return "No MariaDB installation detected"
	}

	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("MariaDB Installation Detected:\n"))
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
		summary.WriteString(fmt.Sprintf("  Data Directory: exists"))
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
