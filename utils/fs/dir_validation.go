package fs

import (
	"fmt"
	"os"
	"path/filepath"

	"sfDBTools/internal/logger"
)

// DirectoryValidationOperations provides directory structure validation operations
type DirectoryValidationOperations interface {
	// Structure validation
	ValidateDirectoryStructure(sourceDir, destDir string) error
	VerifyEssentialDirectories(baseDir string, requiredDirs []string) error

	// Directory content validation
	IsDirectoryEmpty(path string) (bool, error)
	GetDirectoryFileCount(path string) (int, error)
	ValidateDirectoryPermissions(path string, expectedPerms os.FileMode) error

	// Batch directory operations
	VerifyDirectoriesExist(basePath string, dirNames []string) map[string]bool
	EnsureDirectoryStructure(basePath string, requiredDirs []string) error
}

type directoryValidationOperations struct {
	logger *logger.Logger
}

func newDirectoryValidationOperations(logger *logger.Logger) DirectoryValidationOperations {
	return &directoryValidationOperations{
		logger: logger,
	}
}

// ValidateDirectoryStructure ensures the basic directory structure is preserved between source and destination
func (d *directoryValidationOperations) ValidateDirectoryStructure(sourceDir, destDir string) error {
	sourceInfo, err := os.Stat(sourceDir)
	if err != nil {
		return fmt.Errorf("cannot stat source directory %s: %w", sourceDir, err)
	}

	destInfo, err := os.Stat(destDir)
	if err != nil {
		return fmt.Errorf("cannot stat destination directory %s: %w", destDir, err)
	}

	if !sourceInfo.IsDir() || !destInfo.IsDir() {
		return fmt.Errorf("source or destination is not a directory")
	}

	// Check that destination directory is writable
	if err := d.checkWritable(destDir); err != nil {
		return fmt.Errorf("destination directory is not writable: %w", err)
	}

	return nil
}

// VerifyEssentialDirectories checks for essential directories in a base directory
func (d *directoryValidationOperations) VerifyEssentialDirectories(baseDir string, requiredDirs []string) error {
	for _, dir := range requiredDirs {
		dirPath := filepath.Join(baseDir, dir)
		if !d.directoryExists(dirPath) {
			return fmt.Errorf("essential directory missing: %s", dir)
		}

		// Check if directory contains some files (not empty)
		isEmpty, err := d.IsDirectoryEmpty(dirPath)
		if err != nil {
			return fmt.Errorf("cannot check if directory %s is empty: %w", dir, err)
		}

		if isEmpty {
			d.logger.Warn("Essential directory is empty", logger.String("directory", dir))
		}
	}

	return nil
}

// IsDirectoryEmpty checks if a directory is empty
func (d *directoryValidationOperations) IsDirectoryEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("cannot read directory %s: %w", path, err)
	}

	return len(entries) == 0, nil
}

// GetDirectoryFileCount returns the number of files (not directories) in a directory
func (d *directoryValidationOperations) GetDirectoryFileCount(path string) (int, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, fmt.Errorf("cannot read directory %s: %w", path, err)
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}

	return count, nil
}

// ValidateDirectoryPermissions checks if a directory has the expected permissions
func (d *directoryValidationOperations) ValidateDirectoryPermissions(path string, expectedPerms os.FileMode) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot stat directory %s: %w", path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	actualPerms := info.Mode().Perm()
	if actualPerms != expectedPerms {
		return fmt.Errorf("directory permissions mismatch for %s: expected %o, got %o",
			path, expectedPerms, actualPerms)
	}

	return nil
}

// VerifyDirectoriesExist checks multiple directories for existence
func (d *directoryValidationOperations) VerifyDirectoriesExist(basePath string, dirNames []string) map[string]bool {
	results := make(map[string]bool)

	for _, dirName := range dirNames {
		dirPath := filepath.Join(basePath, dirName)
		results[dirName] = d.directoryExists(dirPath)
	}

	return results
}

// EnsureDirectoryStructure creates required directories if they don't exist
func (d *directoryValidationOperations) EnsureDirectoryStructure(basePath string, requiredDirs []string) error {
	for _, dir := range requiredDirs {
		dirPath := filepath.Join(basePath, dir)

		if !d.directoryExists(dirPath) {
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				return fmt.Errorf("failed to create required directory %s: %w", dir, err)
			}

			d.logger.Info("Created required directory", logger.String("path", dirPath))
		}
	}

	return nil
}

// Helper functions

// directoryExists checks if a directory exists
func (d *directoryValidationOperations) directoryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// checkWritable checks if a directory is writable
func (d *directoryValidationOperations) checkWritable(path string) error {
	// Try to create a temporary file in the directory
	tempFile := filepath.Join(path, ".write_test_temp")
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("directory is not writable: %w", err)
	}
	file.Close()

	// Clean up the temporary file
	if err := os.Remove(tempFile); err != nil {
		d.logger.Warn("Failed to remove temp file during write test",
			logger.String("file", tempFile),
			logger.Error(err))
	}

	return nil
}
