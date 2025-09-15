package template

import (
	"fmt"
	"os"
	"time"
)

func (t *MariaDBConfigTemplate) BackupCurrentConfig(backupDir string) (string, error) {
	if t.CurrentPath == "" {
		return "", fmt.Errorf("no current config path to backup")
	}
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory %s: %w", backupDir, err)
	}
	timestamp := generateTimestamp()
	backupFilename := fmt.Sprintf("mariadb-config-backup-%s.cnf", timestamp)
	backupPath := fmt.Sprintf("%s/%s", backupDir, backupFilename)
	if err := copyFile(t.CurrentPath, backupPath); err != nil {
		return "", fmt.Errorf("failed to backup config file: %w", err)
	}
	return backupPath, nil
}

func generateTimestamp() string {
	return time.Now().Format("20060102-150405")
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer sourceFile.Close()
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer destFile.Close()
	if _, err := destFile.ReadFrom(sourceFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}
	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}
	return nil
}
