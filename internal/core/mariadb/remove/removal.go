package remove

import (
	"fmt"
	"os"
	"os/exec"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
)

// RemovalService handles the actual removal of MariaDB components
type RemovalService struct {
	osInfo *common.OSInfo
}

// NewRemovalService creates a new removal service
func NewRemovalService(osInfo *common.OSInfo) *RemovalService {
	return &RemovalService{
		osInfo: osInfo,
	}
}

// RemovePackages removes MariaDB packages from the system
func (r *RemovalService) RemovePackages(installation *DetectedInstallation) error {
	lg, _ := logger.Get()

	if !installation.IsInstalled {
		lg.Info("No MariaDB packages found to remove")
		return nil
	}

	lg.Info("Removing MariaDB packages",
		logger.String("package", installation.PackageName),
		logger.String("package_type", r.osInfo.PackageType))

	var cmd *exec.Cmd

	switch r.osInfo.PackageType {
	case "rpm":
		// Remove RPM packages - remove all MariaDB related packages
		packages := []string{"MariaDB-server", "MariaDB-client", "MariaDB-common", "MariaDB-compat", "mariadb-server", "mariadb-client"}
		packageArgs := append([]string{"remove", "-y"}, packages...)
		cmd = exec.Command("yum", packageArgs...)

	case "deb":
		// Remove DEB packages - purge to remove config files too
		cmd = exec.Command("apt-get", "purge", "-y", "mariadb-server*", "mariadb-client*", "mariadb-common*")

	default:
		return fmt.Errorf("unsupported package type: %s", r.osInfo.PackageType)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to remove packages",
			logger.String("output", string(output)),
			logger.Error(err))
		return fmt.Errorf("failed to remove MariaDB packages: %w\nOutput: %s", err, string(output))
	}

	lg.Info("Successfully removed MariaDB packages",
		logger.String("output", string(output)))

	return nil
}

// StopAndDisableServices stops and disables MariaDB services
func (r *RemovalService) StopAndDisableServices(installation *DetectedInstallation) error {
	lg, _ := logger.Get()

	if installation.ServiceName == "" {
		lg.Info("No MariaDB service found to stop")
		return nil
	}

	serviceName := installation.ServiceName

	// Stop the service if it's running
	if installation.ServiceActive {
		lg.Info("Stopping MariaDB service", logger.String("service", serviceName))

		cmd := exec.Command("systemctl", "stop", serviceName)
		if output, err := cmd.CombinedOutput(); err != nil {
			lg.Warn("Failed to stop service",
				logger.String("service", serviceName),
				logger.String("output", string(output)),
				logger.Error(err))
		} else {
			lg.Info("Service stopped successfully", logger.String("service", serviceName))
		}
	}

	// Disable the service if it's enabled
	if installation.ServiceEnabled {
		lg.Info("Disabling MariaDB service", logger.String("service", serviceName))

		cmd := exec.Command("systemctl", "disable", serviceName)
		if output, err := cmd.CombinedOutput(); err != nil {
			lg.Warn("Failed to disable service",
				logger.String("service", serviceName),
				logger.String("output", string(output)),
				logger.Error(err))
		} else {
			lg.Info("Service disabled successfully", logger.String("service", serviceName))
		}
	}

	return nil
}

// RemoveDataDirectories removes MariaDB data directories
func (r *RemovalService) RemoveDataDirectories(installation *DetectedInstallation, config *RemovalConfig) error {
	lg, _ := logger.Get()

	if !config.RemoveData {
		lg.Info("Data removal is disabled, skipping")
		return nil
	}

	if !installation.DataDirectoryExists {
		lg.Info("No data directory found to remove")
		return nil
	}

	dataDirs := []string{
		config.DataDirectory,
		"/var/lib/mysql",
		"/var/lib/mariadb",
	}

	for _, dataDir := range dataDirs {
		if stat, err := os.Stat(dataDir); err == nil && stat.IsDir() {
			lg.Info("Removing data directory",
				logger.String("directory", dataDir),
				logger.String("size", r.formatSize(installation.DataDirectorySize)))

			if err := os.RemoveAll(dataDir); err != nil {
				lg.Error("Failed to remove data directory",
					logger.String("directory", dataDir),
					logger.Error(err))
				return fmt.Errorf("failed to remove data directory %s: %w", dataDir, err)
			}

			lg.Info("Data directory removed successfully", logger.String("directory", dataDir))
		}
	}

	return nil
}

// RemoveConfigFiles removes MariaDB configuration files
func (r *RemovalService) RemoveConfigFiles(installation *DetectedInstallation) error {
	lg, _ := logger.Get()

	if len(installation.ConfigFiles) == 0 {
		lg.Info("No configuration files found to remove")
		return nil
	}

	for _, configFile := range installation.ConfigFiles {
		lg.Info("Removing configuration file", logger.String("file", configFile))

		if err := os.RemoveAll(configFile); err != nil {
			lg.Warn("Failed to remove configuration file",
				logger.String("file", configFile),
				logger.Error(err))
		} else {
			lg.Info("Configuration file removed successfully", logger.String("file", configFile))
		}
	}

	return nil
}

// RemoveLogFiles removes MariaDB log files
func (r *RemovalService) RemoveLogFiles(installation *DetectedInstallation) error {
	lg, _ := logger.Get()

	if len(installation.LogFiles) == 0 {
		lg.Info("No log files found to remove")
		return nil
	}

	for _, logFile := range installation.LogFiles {
		lg.Info("Removing log file", logger.String("file", logFile))

		if err := os.RemoveAll(logFile); err != nil {
			lg.Warn("Failed to remove log file",
				logger.String("file", logFile),
				logger.Error(err))
		} else {
			lg.Info("Log file removed successfully", logger.String("file", logFile))
		}
	}

	return nil
}

// RemoveRepositories removes MariaDB repositories
func (r *RemovalService) RemoveRepositories(config *RemovalConfig) error {
	lg, _ := logger.Get()

	if !config.RemoveRepositories {
		lg.Info("Repository removal is disabled, skipping")
		return nil
	}

	switch r.osInfo.PackageType {
	case "rpm":
		return r.removeRPMRepositories()
	case "deb":
		return r.removeDEBRepositories()
	default:
		return fmt.Errorf("unsupported package type: %s", r.osInfo.PackageType)
	}
}

// removeRPMRepositories removes RPM repositories
func (r *RemovalService) removeRPMRepositories() error {
	lg, _ := logger.Get()

	repoFiles := []string{
		"/etc/yum.repos.d/mariadb.repo",
		"/etc/yum.repos.d/MariaDB.repo",
	}

	for _, repoFile := range repoFiles {
		if _, err := os.Stat(repoFile); err == nil {
			lg.Info("Removing repository file", logger.String("file", repoFile))

			if err := os.Remove(repoFile); err != nil {
				lg.Warn("Failed to remove repository file",
					logger.String("file", repoFile),
					logger.Error(err))
			} else {
				lg.Info("Repository file removed successfully", logger.String("file", repoFile))
			}
		}
	}

	// Clean package cache
	cmd := exec.Command("yum", "clean", "all")
	if output, err := cmd.CombinedOutput(); err != nil {
		lg.Warn("Failed to clean package cache",
			logger.String("output", string(output)),
			logger.Error(err))
	}

	return nil
}

// removeDEBRepositories removes DEB repositories
func (r *RemovalService) removeDEBRepositories() error {
	lg, _ := logger.Get()

	// Remove MariaDB apt keys
	cmd := exec.Command("apt-key", "del", "0xF1656F24C74CD1D8")
	if output, err := cmd.CombinedOutput(); err != nil {
		lg.Warn("Failed to remove MariaDB apt key",
			logger.String("output", string(output)),
			logger.Error(err))
	}

	// Remove repository sources
	sourcesFiles := []string{
		"/etc/apt/sources.list.d/mariadb.list",
		"/etc/apt/sources.list.d/MariaDB.list",
	}

	for _, sourceFile := range sourcesFiles {
		if _, err := os.Stat(sourceFile); err == nil {
			lg.Info("Removing repository source file", logger.String("file", sourceFile))

			if err := os.Remove(sourceFile); err != nil {
				lg.Warn("Failed to remove repository source file",
					logger.String("file", sourceFile),
					logger.Error(err))
			} else {
				lg.Info("Repository source file removed successfully", logger.String("file", sourceFile))
			}
		}
	}

	// Update package cache
	cmd = exec.Command("apt-get", "update")
	if output, err := cmd.CombinedOutput(); err != nil {
		lg.Warn("Failed to update package cache",
			logger.String("output", string(output)),
			logger.Error(err))
	}

	return nil
}

// CleanupResidualFiles removes any remaining MariaDB files
func (r *RemovalService) CleanupResidualFiles() error {
	lg, _ := logger.Get()

	// Common residual directories and files
	residualPaths := []string{
		"/etc/systemd/system/mariadb.service.d",
		"/etc/systemd/system/mysql.service.d",
		"/run/mysqld",
		"/tmp/mysql.sock",
		"/var/run/mysqld",
	}

	for _, path := range residualPaths {
		if _, err := os.Stat(path); err == nil {
			lg.Info("Cleaning up residual path", logger.String("path", path))

			if err := os.RemoveAll(path); err != nil {
				lg.Warn("Failed to remove residual path",
					logger.String("path", path),
					logger.Error(err))
			} else {
				lg.Info("Residual path removed successfully", logger.String("path", path))
			}
		}
	}

	// Reload systemd daemon to clear any service references
	cmd := exec.Command("systemctl", "daemon-reload")
	if output, err := cmd.CombinedOutput(); err != nil {
		lg.Warn("Failed to reload systemd daemon",
			logger.String("output", string(output)),
			logger.Error(err))
	}

	return nil
}

// formatSize formats a byte size into human-readable format
func (r *RemovalService) formatSize(bytes int64) string {
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
