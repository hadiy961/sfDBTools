package remove

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
	"time"
)

// BackupStep creates a backup of MariaDB data before removal
type BackupStep struct {
	deps Dependencies
}

// Name returns the step name
func (s *BackupStep) Name() string {
	return "Data Backup"
}

// Validate validates the step preconditions
func (s *BackupStep) Validate(state *State) error {
	if state.Installation == nil {
		return fmt.Errorf("installation detection is required before backup")
	}
	return nil
}

// Execute creates backup of MariaDB data
func (s *BackupStep) Execute(ctx context.Context, state *State) error {
	lg, _ := logger.Get()

	installation := state.Installation
	config := state.Config

	// Skip backup if not requested
	if !config.BackupData {
		terminal.PrintInfo("Skipping data backup (--backup-data not specified)")
		return nil
	}

	// Skip if no data directory or it doesn't exist
	if installation.ActualDataDir == "" || !s.deps.FileSystem.Exists(installation.ActualDataDir) {
		terminal.PrintInfo("No data directory to backup")
		return nil
	}

	lg.Info("Starting data backup")

	// Determine backup path
	backupPath := config.BackupPath
	if backupPath == "" {
		homeDir, _ := os.UserHomeDir()
		timestamp := time.Now().Format("20060102_150405")
		backupPath = filepath.Join(homeDir, "mariadb_backups", fmt.Sprintf("mariadb_backup_%s", timestamp))
	}

	// Ensure backup directory exists
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory %s: %w", backupPath, err)
	}

	// Store backup path for other steps and rollback
	if state.RollbackData == nil {
		state.RollbackData = make(map[string]interface{})
	}
	state.RollbackData["backupPath"] = backupPath
	state.BackupPath = backupPath

	terminal.PrintInfo(fmt.Sprintf("Creating backup at: %s", backupPath))

	spinner := terminal.NewProgressSpinner("Backing up MariaDB data...")
	spinner.Start()
	defer spinner.Stop()

	// Create backup of data directory
	dataBackupPath := filepath.Join(backupPath, "data.tar.gz")
	if err := s.createTarBackup(installation.ActualDataDir, dataBackupPath); err != nil {
		return fmt.Errorf("failed to backup data directory: %w", err)
	}

	// Backup configuration files
	configBackupDir := filepath.Join(backupPath, "config")
	if err := os.MkdirAll(configBackupDir, 0755); err != nil {
		lg.Warn("Failed to create config backup directory", logger.Error(err))
	} else {
		for _, configFile := range installation.ConfigFiles {
			if s.deps.FileSystem.Exists(configFile) {
				configFileName := filepath.Base(configFile)
				configBackupPath := filepath.Join(configBackupDir, configFileName)
				if err := s.deps.FileSystem.CreateBackup(configFile, configBackupPath); err != nil {
					lg.Warn("Failed to backup config file", logger.String("file", configFile), logger.Error(err))
				}
			}
		}
	}

	// Calculate and display backup size
	backupSize, err := s.deps.FileSystem.CalculateSize(backupPath)
	if err == nil {
		terminal.PrintSuccess(fmt.Sprintf("Backup completed (%.2f MB) at: %s",
			float64(backupSize)/(1024*1024), backupPath))
	} else {
		terminal.PrintSuccess(fmt.Sprintf("Backup completed at: %s", backupPath))
	}

	return nil
}

// Rollback removes the backup if it was created
func (s *BackupStep) Rollback(ctx context.Context, state *State) error {
	lg, _ := logger.Get()

	backupPath, exists := state.RollbackData["backupPath"].(string)
	if !exists || backupPath == "" {
		return nil
	}

	lg.Info("Removing backup created during failed removal", logger.String("path", backupPath))

	if s.deps.FileSystem.Exists(backupPath) {
		if err := os.RemoveAll(backupPath); err != nil {
			lg.Error("Failed to remove backup during rollback", logger.String("path", backupPath), logger.Error(err))
			return err
		}
		terminal.PrintInfo(fmt.Sprintf("Removed backup: %s", backupPath))
	}

	return nil
}

// createTarBackup creates a tar.gz backup of a directory
func (s *BackupStep) createTarBackup(sourceDir, targetFile string) error {
	// Use tar command for creating compressed backup
	args := []string{"-czf", targetFile, "-C", filepath.Dir(sourceDir), filepath.Base(sourceDir)}

	cmd := system.NewProcessManager()
	if err := cmd.Execute("tar", args); err != nil {
		return fmt.Errorf("tar backup failed: %w", err)
	}

	return nil
}
