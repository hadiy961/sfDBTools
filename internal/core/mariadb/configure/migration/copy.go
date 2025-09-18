package migration

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/internal/logger"
	fsutil "sfDBTools/utils/fs"
)

func PerformSingleMigration(migration DataMigration) error {
	lg, _ := logger.Get()
	lg.Info("Performing migration", logger.String("type", migration.Type))

	if _, err := os.Stat(migration.Source); os.IsNotExist(err) {
		if migration.Critical {
			return fmt.Errorf("source directory does not exist: %s", migration.Source)
		}
		lg.Warn("Source directory does not exist, skipping migration")
		return nil
	}

	if err := os.MkdirAll(migration.Destination, 0750); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Handle logs migration differently - only copy log files
	if migration.Type == "logs" {
		if err := copyLogFilesOnly(migration.Source, migration.Destination); err != nil {
			return fmt.Errorf("failed to copy log files: %w", err)
		}
	} else {
		if err := copyDirectory(migration.Source, migration.Destination); err != nil {
			return fmt.Errorf("failed to copy data: %w", err)
		}
	}

	if migration.Critical && migration.Type == "data" {
		if err := VerifyDataMigration(migration.Source, migration.Destination); err != nil {
			return fmt.Errorf("data verification failed: %w", err)
		}
	}

	lg.Info("Migration completed successfully")
	return nil
}

func copyDirectory(source, destination string) error {
	lg, _ := logger.Get()

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
		if isSpecialFile(info.Mode()) {
			lg.Warn("Skipping unsupported special file during migration", logger.String("path", path))
			return nil
		}

		// Handle directories
		if info.IsDir() {
			return copyDir(destPath, info)
		}

		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			return copySymlink(path, destPath, info)
		}

		// Handle regular files
		return func() error {
			if err := fsutil.CopyFile(path, destPath, info); err != nil {
				return fmt.Errorf("failed to copy file %s to %s: %w", path, destPath, err)
			}
			return nil
		}()
	})
}

func isSpecialFile(mode os.FileMode) bool {
	return mode&(os.ModeSocket|os.ModeDevice|os.ModeNamedPipe) != 0
}

func copyDir(destPath string, info os.FileInfo) error {
	if err := os.MkdirAll(destPath, info.Mode().Perm()); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", destPath, err)
	}
	preserveOwnership(destPath, info)
	return nil
}

func copySymlink(srcPath, destPath string, info os.FileInfo) error {
	target, err := os.Readlink(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read symlink %s: %w", srcPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent for symlink %s: %w", destPath, err)
	}

	_ = os.Remove(destPath) // Remove if exists
	if err := os.Symlink(target, destPath); err != nil {
		return fmt.Errorf("failed to create symlink %s -> %s: %w", destPath, target, err)
	}

	preserveSymlinkOwnership(destPath, info)
	return nil
}

func copyFile(srcPath, destPath string, info os.FileInfo, lg *logger.Logger) error {
	// now delegated to utils/fs.CopyFile which handles parent creation,
	// copying, and preserving permissions/ownership.
	if err := fsutil.CopyFile(srcPath, destPath, info); err != nil {
		return fmt.Errorf("failed to copy file %s to %s: %w", srcPath, destPath, err)
	}
	return nil
}

func preserveOwnership(path string, info os.FileInfo) {
	// delegate to fs helper
	fsutil.PreserveOwnership(path, info)
}

func preserveSymlinkOwnership(path string, info os.FileInfo) {
	// delegate to fs helper
	fsutil.PreserveSymlinkOwnership(path, info)
}

func preserveFilePermissions(path string, info os.FileInfo, lg *logger.Logger) {
	// delegate to fs helper (uses logger internally)
	fsutil.SetPermissionsAndOwnership(path, info.Mode(), info)
}

// copyLogFilesOnly copies only log-related files from source to destination
// This prevents copying entire database files when logs are in the same directory as data
func copyLogFilesOnly(source, destination string) error {
	lg, _ := logger.Get()
	lg.Info("Copying log files only", logger.String("source", source), logger.String("destination", destination))

	return filepath.WalkDir(source, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("error walking source %s: %w", path, walkErr)
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to stat source %s: %w", path, err)
		}

		// Skip special files
		if isSpecialFile(info.Mode()) {
			return nil
		}

		// Handle directories - create them but don't descend into data directories
		if info.IsDir() {
			if isDataDirectory(path, source) {
				lg.Info("Skipping data directory during log migration", logger.String("path", path))
				return filepath.SkipDir
			}
			rel, err := filepath.Rel(source, path)
			if err != nil {
				return fmt.Errorf("failed to compute relative path: %w", err)
			}
			destPath := filepath.Join(destination, rel)
			return copyDir(destPath, info)
		}

		// Only copy files that are log-related
		if !isLogFile(path) {
			return nil
		}

		rel, err := filepath.Rel(source, path)
		if err != nil {
			return fmt.Errorf("failed to compute relative path: %w", err)
		}
		destPath := filepath.Join(destination, rel)

		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			return copySymlink(path, destPath, info)
		}

		// Handle regular files
		lg.Info("Copying log file", logger.String("file", path))
		return copyFile(path, destPath, info, lg)
	})
}

// isLogFile determines if a file is a log-related file
func isLogFile(path string) bool {
	fileName := filepath.Base(path)

	// MariaDB/MySQL log file patterns
	logExtensions := []string{".log", ".err", ".pid"}
	logPrefixes := []string{"mysql-bin.", "mysql-relay-bin.", "slow", "error", "general"}
	logNames := []string{"mysql.log", "mysqld.log", "error.log", "slow.log", "general.log", "relay.log", "mysqld.pid"}

	// Check extensions
	for _, ext := range logExtensions {
		if strings.HasSuffix(fileName, ext) {
			return true
		}
	}

	// Check prefixes
	for _, prefix := range logPrefixes {
		if strings.HasPrefix(fileName, prefix) {
			return true
		}
	}

	// Check exact names
	for _, name := range logNames {
		if fileName == name {
			return true
		}
	}

	return false
}

// isDataDirectory determines if a directory contains database data that should be skipped during log migration
func isDataDirectory(path, sourceRoot string) bool {
	rel, err := filepath.Rel(sourceRoot, path)
	if err != nil {
		return false
	}

	// Skip database directories and performance_schema, information_schema, etc.
	dirName := filepath.Base(path)
	skipDirs := []string{"mysql", "performance_schema", "information_schema", "sys", "test"}

	for _, skipDir := range skipDirs {
		if dirName == skipDir {
			return true
		}
	}

	// Skip any directory that contains .frm, .ibd, .MYD, .MYI files (database files)
	// but allow if it's the root directory
	if rel != "." {
		entries, err := os.ReadDir(path)
		if err != nil {
			return false
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			fileName := entry.Name()
			dbExtensions := []string{".frm", ".ibd", ".MYD", ".MYI", ".opt", ".ARZ", ".ARM"}
			for _, ext := range dbExtensions {
				if strings.HasSuffix(fileName, ext) {
					return true
				}
			}
		}
	}

	return false
}
