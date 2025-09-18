package migration

import (
	"fmt"

	"sfDBTools/internal/logger"
)

// PerformSingleMigration performs a single data migration using the migration manager
func PerformSingleMigration(migration DataMigration) error {
	lg, _ := logger.Get()
	lg.Info("Performing migration", logger.String("type", migration.Type))

	// Initialize migration manager
	mgr := NewMigrationManager()

	// Check if source directory exists
	if !mgr.FileSystem().Dir().Exists(migration.Source) {
		if migration.Critical {
			return fmt.Errorf("source directory does not exist: %s", migration.Source)
		}
		lg.Warn("Source directory does not exist, skipping migration")
		return nil
	}

	// Create destination directory
	if err := mgr.FileSystem().Dir().CreateWithPerms(migration.Destination, 0750, "", ""); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Handle logs migration differently - only copy log files
	if migration.Type == "logs" {
		if err := mgr.CopyLogFilesOnly(migration.Source, migration.Destination); err != nil {
			return fmt.Errorf("failed to copy log files: %w", err)
		}
	} else {
		if err := mgr.CopyDirectory(migration.Source, migration.Destination); err != nil {
			return fmt.Errorf("failed to copy data: %w", err)
		}
	}

	if migration.Critical && migration.Type == "data" {
		if err := VerifyDataMigration(migration.Source, migration.Destination); err != nil {
			return fmt.Errorf("data verification failed: %w", err)
		}
	}

	// Cleanup source directory after successful migration
	lg.Info("Cleaning up source directory", logger.String("source", migration.Source))
	// If this migration is for binlogs, remove index/state files from source first
	if migration.Type == "binlogs" {
		lg.Info("Removing binlog index/state files from Destination", logger.String("source", migration.Source))
		if err := mgr.FileSystem().Remove(migration.Destination + "/mysql-bin.index"); err != nil {
			lg.Warn("Failed to remove mysql-bin.index", logger.String("file", migration.Destination+"/mysql-bin.index"), logger.Error(err))
		}
		if err := mgr.FileSystem().Remove(migration.Destination + "/mysql-bin.state"); err != nil {
			lg.Warn("Failed to remove mysql-bin.state", logger.String("file", migration.Destination+"/mysql-bin.state"), logger.Error(err))
		}
	}
	if err := mgr.FileSystem().Dir().Remove(migration.Source); err != nil {
		return fmt.Errorf("failed to remove source directory: %w", err)
	}

	lg.Info("Migration completed successfully")
	return nil
}
