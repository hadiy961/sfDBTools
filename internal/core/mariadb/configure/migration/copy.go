package migration

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"sfDBTools/internal/logger"
)

func PerformSingleMigration(migration DataMigration) error {
	lg, _ := logger.Get()
	lg.Info("Performing migration")

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

	if err := copyDirectory(migration.Source, migration.Destination); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
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

		mode := info.Mode()

		if mode&os.ModeSocket != 0 || mode&os.ModeDevice != 0 || mode&os.ModeNamedPipe != 0 {
			lg.Warn("Skipping unsupported special file during migration", logger.String("path", path))
			return nil
		}

		if mode.IsDir() {
			if err := os.MkdirAll(destPath, info.Mode().Perm()); err != nil {
				return fmt.Errorf("failed to create destination directory %s: %w", destPath, err)
			}
			if statT, ok := info.Sys().(*syscall.Stat_t); ok {
				_ = os.Chown(destPath, int(statT.Uid), int(statT.Gid))
			}
			return nil
		}

		if mode&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return fmt.Errorf("failed to read symlink %s: %w", path, err)
			}
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent for symlink %s: %w", destPath, err)
			}
			_ = os.Remove(destPath)
			if err := os.Symlink(target, destPath); err != nil {
				return fmt.Errorf("failed to create symlink %s -> %s: %w", destPath, target, err)
			}
			if statT, ok := info.Sys().(*syscall.Stat_t); ok {
				_ = os.Lchown(destPath, int(statT.Uid), int(statT.Gid))
			}
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for file %s: %w", destPath, err)
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open source file %s: %w", path, err)
		}
		defer srcFile.Close()

		dstFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
		if err != nil {
			return fmt.Errorf("failed to create destination file %s: %w", destPath, err)
		}

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			dstFile.Close()
			return fmt.Errorf("failed to copy file %s to %s: %w", path, destPath, err)
		}
		if err := dstFile.Close(); err != nil {
			return fmt.Errorf("failed to close destination file %s: %w", destPath, err)
		}

		if err := os.Chmod(destPath, info.Mode().Perm()); err != nil {
			lg.Warn("Failed to set permissions on destination file", logger.String("file", destPath))
		}

		if statT, ok := info.Sys().(*syscall.Stat_t); ok {
			if err := os.Chown(destPath, int(statT.Uid), int(statT.Gid)); err != nil {
				lg.Warn("Failed to chown destination file", logger.String("file", destPath))
			}
		}

		return nil
	})
}
