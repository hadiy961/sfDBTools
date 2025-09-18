package migration

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"sfDBTools/internal/logger"
	fsutil "sfDBTools/utils/fs"
)

// CopyDirectory copies entire directory tree from source to destination
func (m *MigrationManager) CopyDirectory(source, destination string) error {
	return filepath.WalkDir(source, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("error walking source %s: %w", path, walkErr)
		}

		rel, err := filepath.Rel(source, path)
		if err != nil {
			return fmt.Errorf("failed to compute relative path: %w", err)
		}
		destPath := filepath.Join(destination, rel)

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to stat source %s: %w", path, err)
		}

		// Skip special files (sockets, devices, pipes)
		if fsutil.IsSpecialFile(info.Mode()) {
			m.logger.Warn("Skipping unsupported special file during migration", logger.String("path", path))
			return nil
		}

		// Handle directories
		if info.IsDir() {
			return m.copyDir(destPath, info)
		}

		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			return m.copySymlink(path, destPath, info)
		}

		// Handle regular files
		if err := m.fsMgr.File().CopyWithInfo(path, destPath, info); err != nil {
			return fmt.Errorf("failed to copy file %s to %s: %w", path, destPath, err)
		}
		return nil
	})
}

// copyDir creates a directory with proper permissions and ownership
func (m *MigrationManager) copyDir(destPath string, info os.FileInfo) error {
	// Use manager methods instead of deprecated functions
	if err := m.fsMgr.Dir().Create(destPath); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", destPath, err)
	}
	// Set permissions after creation
	if err := m.fsMgr.Perm().SetFilePerms(destPath, info.Mode().Perm(), "", ""); err != nil {
		m.logger.Warn("Failed to set directory permissions", logger.String("path", destPath))
	}
	// Preserve ownership
	if err := m.fsMgr.Perm().PreserveOwnership(destPath, info); err != nil {
		m.logger.Warn("Failed to preserve ownership", logger.String("path", destPath))
	}
	return nil
}

// copySymlink creates a symlink with proper permissions and ownership
func (m *MigrationManager) copySymlink(srcPath, destPath string, info os.FileInfo) error {
	target, err := os.Readlink(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read symlink %s: %w", srcPath, err)
	}

	// Use manager method instead of deprecated function
	if err := m.fsMgr.File().EnsureDir(filepath.Dir(destPath)); err != nil {
		return fmt.Errorf("failed to create parent for symlink %s: %w", destPath, err)
	}

	_ = os.Remove(destPath) // Remove if exists
	if err := os.Symlink(target, destPath); err != nil {
		return fmt.Errorf("failed to create symlink %s -> %s: %w", destPath, target, err)
	}

	// Use manager method instead of deprecated function
	if err := m.fsMgr.Perm().PreserveOwnership(destPath, info); err != nil {
		m.logger.Warn("Failed to preserve symlink ownership", logger.String("path", destPath))
	}
	return nil
}
