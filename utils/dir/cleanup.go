package dir

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"sfDBTools/internal/logger"

	"github.com/bmatcuk/doublestar/v4"
)

// Cleanup handles directory cleanup operations
type Cleanup struct {
	manager *Manager
	scanner *Scanner
	logger  *logger.Logger
}

// NewCleanup creates a new cleanup manager
func NewCleanup() *Cleanup {
	lg, _ := logger.Get()
	return &Cleanup{
		manager: NewManager(),
		scanner: NewScanner(),
		logger:  lg,
	}
}

// CleanupOldDirectories menghapus direktori lama berdasarkan retention period
func (c *Cleanup) CleanupOldDirectories(path string, retentionDays int, namePattern string) (*CleanupResult, error) {
	if retentionDays <= 0 {
		return &CleanupResult{}, fmt.Errorf("retention days harus lebih dari 0")
	}

	normalizedPath := filepath.Clean(path)
	if !c.manager.Exists(normalizedPath) {
		return &CleanupResult{}, fmt.Errorf("direktori tidak ada: %s", normalizedPath)
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	result := &CleanupResult{
		RemovedDirs:  []string{},
		RemovedFiles: []string{},
		Errors:       []string{},
	}

	// Scan direktori untuk mencari kandidat penghapusan
	entries, err := c.scanner.List(normalizedPath, ScanOptions{
		Recursive:     false,
		IncludeHidden: false,
		Filter:        FilterDirectoriesOnly(),
	})
	if err != nil {
		return result, fmt.Errorf("gagal scan direktori untuk cleanup: %w", err)
	}

	for _, entry := range entries {
		// Skip jika tidak sesuai pattern
		if namePattern != "" {
			matched := c.matchesPattern(entry.Name, namePattern)
			if !matched {
				continue
			}
		}

		// Skip jika belum melewati retention period
		if !entry.ModTime.Before(cutoffTime) {
			c.logger.Debug("Direktori belum melewati retention period, dilewati",
				logger.String("dir", entry.Name),
				logger.Time("mod_time", entry.ModTime),
				logger.Time("cutoff_time", cutoffTime))
			continue
		}

		// Hitung ukuran sebelum dihapus
		size, err := c.manager.GetSize(entry.Path)
		if err != nil {
			c.logger.Warn("Gagal menghitung ukuran direktori",
				logger.String("path", entry.Path),
				logger.Error(err))
		} else {
			result.TotalSizeFreed += size
		}

		// Hapus direktori
		if err := c.manager.RemoveAll(entry.Path); err != nil {
			errorMsg := fmt.Sprintf("Gagal hapus %s: %v", entry.Name, err)
			result.Errors = append(result.Errors, errorMsg)
			c.logger.Error("Gagal menghapus direktori old",
				logger.String("path", entry.Path),
				logger.Error(err))
		} else {
			result.RemovedDirs = append(result.RemovedDirs, entry.Name)
			result.TotalRemoved++
			c.logger.Info("Direktori lama berhasil dihapus",
				logger.String("name", entry.Name),
				logger.String("path", entry.Path),
				logger.Time("mod_time", entry.ModTime))
		}
	}

	c.logger.Info("Cleanup direktori lama selesai",
		logger.String("path", normalizedPath),
		logger.Int("removed_count", result.TotalRemoved),
		logger.String("size_freed", result.GetFormattedSizeFreed()),
		logger.Int("error_count", len(result.Errors)))

	return result, nil
}

// CleanupOldFiles menghapus file lama berdasarkan retention period
func (c *Cleanup) CleanupOldFiles(path string, retentionDays int, namePattern string) (*CleanupResult, error) {
	if retentionDays <= 0 {
		return &CleanupResult{}, fmt.Errorf("retention days harus lebih dari 0")
	}

	normalizedPath := filepath.Clean(path)
	if !c.manager.Exists(normalizedPath) {
		return &CleanupResult{}, fmt.Errorf("direktori tidak ada: %s", normalizedPath)
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	result := &CleanupResult{
		RemovedDirs:  []string{},
		RemovedFiles: []string{},
		Errors:       []string{},
	}

	// Scan untuk file yang perlu dihapus
	entries, err := c.scanner.List(normalizedPath, ScanOptions{
		Recursive:     true,
		IncludeHidden: false,
		Filter:        FilterFilesOnly(),
	})
	if err != nil {
		return result, fmt.Errorf("gagal scan file untuk cleanup: %w", err)
	}

	for _, entry := range entries {
		// Skip jika tidak sesuai pattern
		if namePattern != "" {
			matched := c.matchesPattern(entry.Name, namePattern)
			if !matched {
				continue
			}
		}

		// Skip jika belum melewati retention period
		if !entry.ModTime.Before(cutoffTime) {
			continue
		}

		// Hapus file
		result.TotalSizeFreed += entry.Size

		if err := c.manager.fs.Remove(entry.Path); err != nil {
			errorMsg := fmt.Sprintf("Gagal hapus %s: %v", entry.Name, err)
			result.Errors = append(result.Errors, errorMsg)
			c.logger.Error("Gagal menghapus file lama",
				logger.String("path", entry.Path),
				logger.Error(err))
		} else {
			result.RemovedFiles = append(result.RemovedFiles, entry.Name)
			result.TotalRemoved++
			c.logger.Debug("File lama berhasil dihapus",
				logger.String("name", entry.Name),
				logger.String("path", entry.Path))
		}
	}

	c.logger.Info("Cleanup file lama selesai",
		logger.String("path", normalizedPath),
		logger.Int("removed_count", result.TotalRemoved),
		logger.String("size_freed", result.GetFormattedSizeFreed()),
		logger.Int("error_count", len(result.Errors)))

	return result, nil
}

// CleanupByPattern menghapus file/direktori berdasarkan pattern glob
func (c *Cleanup) CleanupByPattern(path string, pattern string, olderThanDays int) (*CleanupResult, error) {
	normalizedPath := filepath.Clean(path)
	if !c.manager.Exists(normalizedPath) {
		return &CleanupResult{}, fmt.Errorf("direktori tidak ada: %s", normalizedPath)
	}

	result := &CleanupResult{
		RemovedDirs:  []string{},
		RemovedFiles: []string{},
		Errors:       []string{},
	}

	// Cari matches berdasarkan pattern
	matches, err := c.scanner.Find(normalizedPath, pattern)
	if err != nil {
		return result, fmt.Errorf("gagal mencari dengan pattern '%s': %w", pattern, err)
	}

	cutoffTime := time.Time{}
	if olderThanDays > 0 {
		cutoffTime = time.Now().AddDate(0, 0, -olderThanDays)
	}

	for _, matchPath := range matches {
		info, err := c.manager.fs.Stat(matchPath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Gagal stat %s: %v", matchPath, err))
			continue
		}

		// Skip jika belum melewati retention period (jika ditentukan)
		if !cutoffTime.IsZero() && !info.ModTime().Before(cutoffTime) {
			continue
		}

		if info.IsDir() {
			// Hitung ukuran direktori
			size, err := c.manager.GetSize(matchPath)
			if err == nil {
				result.TotalSizeFreed += size
			}

			// Hapus direktori
			if err := c.manager.RemoveAll(matchPath); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Gagal hapus dir %s: %v", matchPath, err))
			} else {
				result.RemovedDirs = append(result.RemovedDirs, filepath.Base(matchPath))
				result.TotalRemoved++
			}
		} else {
			// Hapus file
			result.TotalSizeFreed += info.Size()

			if err := c.manager.fs.Remove(matchPath); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Gagal hapus file %s: %v", matchPath, err))
			} else {
				result.RemovedFiles = append(result.RemovedFiles, filepath.Base(matchPath))
				result.TotalRemoved++
			}
		}
	}

	c.logger.Info("Cleanup berdasarkan pattern selesai",
		logger.String("path", normalizedPath),
		logger.String("pattern", pattern),
		logger.Int("matches", len(matches)),
		logger.Int("removed", result.TotalRemoved))

	return result, nil
}

// EmptyDirectory mengosongkan direktori tanpa menghapus direktori itu sendiri
func (c *Cleanup) EmptyDirectory(path string) (*CleanupResult, error) {
	normalizedPath := filepath.Clean(path)
	if !c.manager.Exists(normalizedPath) {
		return &CleanupResult{}, fmt.Errorf("direktori tidak ada: %s", normalizedPath)
	}

	result := &CleanupResult{
		RemovedDirs:  []string{},
		RemovedFiles: []string{},
		Errors:       []string{},
	}

	// List semua isi direktori
	entries, err := c.scanner.List(normalizedPath, ScanOptions{
		Recursive:     false,
		IncludeHidden: true, // Include hidden files
	})
	if err != nil {
		return result, fmt.Errorf("gagal list isi direktori: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir {
			// Hitung ukuran direktori
			size, err := c.manager.GetSize(entry.Path)
			if err == nil {
				result.TotalSizeFreed += size
			}

			// Hapus direktori rekursif
			if err := c.manager.RemoveAll(entry.Path); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Gagal hapus dir %s: %v", entry.Name, err))
			} else {
				result.RemovedDirs = append(result.RemovedDirs, entry.Name)
				result.TotalRemoved++
			}
		} else {
			// Hapus file
			result.TotalSizeFreed += entry.Size

			if err := c.manager.fs.Remove(entry.Path); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Gagal hapus file %s: %v", entry.Name, err))
			} else {
				result.RemovedFiles = append(result.RemovedFiles, entry.Name)
				result.TotalRemoved++
			}
		}
	}

	c.logger.Info("Direktori berhasil dikosongkan",
		logger.String("path", normalizedPath),
		logger.Int("removed_count", result.TotalRemoved),
		logger.String("size_freed", result.GetFormattedSizeFreed()))

	return result, nil
}

// CleanupTemporaryFiles menghapus file temporary berdasarkan pattern umum
func (c *Cleanup) CleanupTemporaryFiles(path string, maxAge time.Duration) (*CleanupResult, error) {
	// Pattern umum untuk temporary files
	tempPatterns := []string{
		"*.tmp",
		"*.temp",
		"*~",
		".DS_Store",
		"Thumbs.db",
		"*.swp",
		"*.swo",
		"*.bak",
		"*.backup",
		"*sfdbtools_write_test*",
		"*.sfbackup_test*",
	}

	result := &CleanupResult{
		RemovedDirs:  []string{},
		RemovedFiles: []string{},
		Errors:       []string{},
	}

	cutoffTime := time.Now().Add(-maxAge)

	for _, pattern := range tempPatterns {
		matches, err := c.scanner.Find(path, pattern)
		if err != nil {
			c.logger.Warn("Gagal mencari temporary files dengan pattern",
				logger.String("pattern", pattern),
				logger.Error(err))
			continue
		}

		for _, matchPath := range matches {
			info, err := c.manager.fs.Stat(matchPath)
			if err != nil {
				continue
			}

			// Skip jika file masih fresh
			if info.ModTime().After(cutoffTime) {
				continue
			}

			// Hapus file temporary
			result.TotalSizeFreed += info.Size()

			if err := c.manager.fs.Remove(matchPath); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Gagal hapus temp file %s: %v", matchPath, err))
			} else {
				result.RemovedFiles = append(result.RemovedFiles, filepath.Base(matchPath))
				result.TotalRemoved++
			}
		}
	}

	c.logger.Info("Cleanup temporary files selesai",
		logger.String("path", path),
		logger.Int("removed_count", result.TotalRemoved),
		logger.String("size_freed", result.GetFormattedSizeFreed()))

	return result, nil
}

// matchesPattern mengecek apakah nama cocok dengan pattern
func (c *Cleanup) matchesPattern(name, pattern string) bool {
	if pattern == "" {
		return true
	}

	// Support simple wildcard matching dan contains
	if strings.Contains(pattern, "*") {
		matched, err := doublestar.Match(pattern, name)
		return err == nil && matched
	}

	// Simple contains check
	return strings.Contains(name, pattern)
}

// Convenience functions untuk backward compatibility
func CleanupOldBackups(outputDir string, retentionDays int) ([]string, error) {
	cleanup := NewCleanup()
	result, err := cleanup.CleanupOldDirectories(outputDir, retentionDays, "")
	return result.RemovedDirs, err
}

func CleanupOldFiles(path string, retentionDays int, pattern string) (int, error) {
	cleanup := NewCleanup()
	result, err := cleanup.CleanupOldFiles(path, retentionDays, pattern)
	if err != nil {
		return 0, err
	}
	return result.TotalRemoved, nil
}
