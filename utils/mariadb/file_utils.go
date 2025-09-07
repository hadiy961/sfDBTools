package mariadb

import (
	"os"
	"path/filepath"
)

// FileUtils provides common file operation utilities for MariaDB operations
type FileUtils struct{}

// NewFileUtils creates a new file utilities instance
func NewFileUtils() *FileUtils {
	return &FileUtils{}
}

// Exists checks if a file or directory exists
func (fu *FileUtils) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// SafeRemoveAll removes a directory or file, ignoring errors
func (fu *FileUtils) SafeRemoveAll(path string) {
	_ = os.RemoveAll(path)
}

// CleanPath returns the shortest path name equivalent to path
func (fu *FileUtils) CleanPath(path string) string {
	return filepath.Clean(path)
}

// FindFilesWithName recursively finds files with the specified name
func (fu *FileUtils) FindFilesWithName(rootPath, fileName string) []string {
	var results []string

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info == nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() == fileName {
			results = append(results, path)
		}
		return nil
	})

	return results
}

// DeduplicateStringSlice removes duplicate entries from a string slice
func (fu *FileUtils) DeduplicateStringSlice(slice []string) []string {
	seen := map[string]struct{}{}
	result := []string{}

	for _, item := range slice {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
