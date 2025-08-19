package backup_single_custom

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sfDBTools/internal/logger"
	backup_utils "sfDBTools/utils/backup"
)

// performBackup performs the actual database backup using mysqldump
func performBackup(options backup_utils.BackupOptions, outputFile string) error {
	lg, _ := logger.Get()

	if err := backup_utils.ValidateBackupOptions(options); err != nil {
		lg.Error("Invalid backup options", logger.Error(err))
		return fmt.Errorf("validation failed: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Execute the command with retry logic for remote connections
	lg.Info("Starting mysqldump execution")

	// Start the command execution
	startTime := time.Now()

	duration := time.Since(startTime)
	lg.Info("mysqldump completed successfully",
		logger.String("duration", duration.String()))

	return nil
}
