package remove

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
)

// BackupService handles data backup before removal
type BackupService struct {
	osInfo *common.OSInfo
}

// NewBackupService creates a new backup service
func NewBackupService(osInfo *common.OSInfo) *BackupService {
	return &BackupService{
		osInfo: osInfo,
	}
}

// BackupData creates a backup of MariaDB data before removal
func (b *BackupService) BackupData(installation *DetectedInstallation, backupPath string) error {
	lg, _ := logger.Get()

	if !installation.DataDirectoryExists {
		lg.Info("No data directory found, skipping backup")
		return nil
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(backupPath, fmt.Sprintf("mariadb_backup_%s.tar.gz", timestamp))

	lg.Info("Creating data backup",
		logger.String("backup_file", backupFile),
		logger.String("data_size", b.formatSize(installation.DataDirectorySize)))

	// Stop MariaDB service before backup to ensure consistency
	if installation.ServiceActive {
		if err := b.stopService(installation.ServiceName); err != nil {
			lg.Warn("Failed to stop service before backup", logger.Error(err))
		}
	}

	// Create compressed backup of data directory
	dataDir := "/var/lib/mysql" // Default, could be configurable
	cmd := exec.Command("tar", "-czf", backupFile, "-C", filepath.Dir(dataDir), filepath.Base(dataDir))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create backup: %w\nOutput: %s", err, string(output))
	}

	// Verify backup file was created
	if stat, err := os.Stat(backupFile); err != nil {
		return fmt.Errorf("backup file was not created: %w", err)
	} else {
		lg.Info("Backup created successfully",
			logger.String("backup_file", backupFile),
			logger.String("backup_size", b.formatSize(stat.Size())))
	}

	return nil
}

// stopService stops a systemd service
func (b *BackupService) stopService(serviceName string) error {
	cmd := exec.Command("systemctl", "stop", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop service %s: %w\nOutput: %s", serviceName, err, string(output))
	}

	return nil
}

// formatSize formats a byte size into human-readable format
func (b *BackupService) formatSize(bytes int64) string {
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
