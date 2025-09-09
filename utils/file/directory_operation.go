// directory_operations.go
package file

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/logger"
)

// DirectoryManager handles directory operations
type DirectoryManager struct {
	logger *logger.Logger
}

// NewDirectoryManager creates a new DirectoryManager instance
func NewDirectoryManager() *DirectoryManager {
	lg, _ := logger.Get()
	return &DirectoryManager{
		logger: lg,
	}
}

// Create creates a directory if it doesn't exist
func (dm *DirectoryManager) Create(dir string) error {
	return dm.Validate(dir)
}

// Validate ensures the directory exists and is writable
func (dm *DirectoryManager) Validate(dir string) error {
	if err := dm.validatePath(dir); err != nil {
		return err
	}

	if err := dm.ensureExists(dir); err != nil {
		return err
	}

	if err := dm.validateIsDirectory(dir); err != nil {
		return err
	}

	return dm.checkWritePermission(dir)
}

// validatePath checks if the directory path is valid
func (dm *DirectoryManager) validatePath(dir string) error {
	if dir == "" {
		dm.logger.Error("Output directory is required")
		return fmt.Errorf("output directory is required")
	}
	return nil
}

// ensureExists creates the directory if it doesn't exist
func (dm *DirectoryManager) ensureExists(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		dm.logger.Debug("Directory does not exist, attempting to create", logger.String("dir", dir))
		if err := os.MkdirAll(dir, 0755); err != nil {
			dm.logger.Error("Failed to create directory",
				logger.String("dir", dir),
				logger.Error(err))
			return fmt.Errorf("failed to create directory '%s': %w", dir, err)
		}
		dm.logger.Info("Directory created", logger.String("dir", dir))
	} else if err != nil {
		dm.logger.Error("Failed to access output directory",
			logger.String("dir", dir),
			logger.Error(err))
		return fmt.Errorf("failed to access output directory '%s': %w", dir, err)
	}
	return nil
}

// validateIsDirectory checks if the path is actually a directory
func (dm *DirectoryManager) validateIsDirectory(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("failed to stat directory '%s': %w", dir, err)
	}

	if !info.IsDir() {
		dm.logger.Error("Output path is not a directory", logger.String("dir", dir))
		return fmt.Errorf("output path '%s' is not a directory", dir)
	}
	return nil
}

// checkWritePermission verifies that the directory is writable
func (dm *DirectoryManager) checkWritePermission(dir string) error {
	testFile := filepath.Join(dir, ".sfbackup_test")
	dm.logger.Debug("Checking write permission for output directory", logger.String("dir", dir))

	f, err := os.Create(testFile)
	if err != nil {
		dm.logger.Error("Output directory is not writable",
			logger.String("dir", dir),
			logger.Error(err))
		return fmt.Errorf("output directory '%s' is not writable: %w", dir, err)
	}
	f.Close()

	// Remove test file; log warning if removal fails but don't fail the operation
	if err := os.Remove(testFile); err != nil {
		dm.logger.Warn("Failed to remove test file in output directory",
			logger.String("file", testFile),
			logger.Error(err))
	}

	dm.logger.Debug("Output directory is writable", logger.String("dir", dir))
	return nil
}

// CreateDir creates the output directory if it doesn't exist (backward compatibility)
func CreateDir(dir string) error {
	manager := NewDirectoryManager()
	return manager.Create(dir)
}

// ValidateDir ensures outputDir exists and is writable (backward compatibility)
func ValidateDir(dir string) error {
	manager := NewDirectoryManager()
	return manager.Validate(dir)
}
