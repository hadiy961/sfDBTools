package migration

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"sfDBTools/internal/logger"
	fsutil "sfDBTools/utils/fs"
)

// CopyLogFilesOnly copies only log-related files from source to destination
// This prevents copying entire database files when logs are in the same directory as data
func (m *MigrationManager) CopyLogFilesOnly(source, destination string) error {
	m.logger.Info("Copying log files only", logger.String("source", source), logger.String("destination", destination))

	return filepath.WalkDir(source, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("error walking source %s: %w", path, walkErr)
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to stat source %s: %w", path, err)
		}

		// Skip special files
		if fsutil.IsSpecialFile(info.Mode()) {
			return nil
		}

		// Handle directories - create them but don't descend into data directories
		if info.IsDir() {
			if m.IsDataDirectory(path, source) {
				m.logger.Info("Skipping data directory during log migration", logger.String("path", path))
				return filepath.SkipDir
			}
			rel, err := filepath.Rel(source, path)
			if err != nil {
				return fmt.Errorf("failed to compute relative path: %w", err)
			}
			destPath := filepath.Join(destination, rel)
			return m.copyDir(destPath, info)
		}

		// Only copy files that are log-related
		if !m.IsLogFile(path) {
			return nil
		}

		rel, err := filepath.Rel(source, path)
		if err != nil {
			return fmt.Errorf("failed to compute relative path: %w", err)
		}
		destPath := filepath.Join(destination, rel)

		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			return m.copySymlink(path, destPath, info)
		}

		// Handle regular files
		m.logger.Info("Copying log file", logger.String("file", path))
		if err := m.fsMgr.File().CopyWithInfo(path, destPath, info); err != nil {
			return fmt.Errorf("failed to copy file %s to %s: %w", path, destPath, err)
		}
		return nil
	})
}
