package fs

import (
	"fmt"
	"path/filepath"
	"time"

	"sfDBTools/internal/logger"
)

// CleanupResult merepresentasikan hasil operasi cleanup
type CleanupResult struct {
	RemovedDirs    []string `json:"removed_dirs"`
	RemovedFiles   []string `json:"removed_files"`
	TotalRemoved   int      `json:"total_removed"`
	TotalSizeFreed int64    `json:"total_size_freed"`
	Errors         []string `json:"errors"`
}

// GetFormattedSizeFreed mengembalikan ukuran yang dibebaskan dalam format yang mudah dibaca
func (c *CleanupResult) GetFormattedSizeFreed() string {
	return FormatSize(c.TotalSizeFreed)
}

// Cleaner handles cleanup operations
type Cleaner struct {
	manager *Manager
	scanner *Scanner
	logger  *logger.Logger
}

// NewCleaner membuat cleaner baru
func NewCleaner() *Cleaner {
	lg, _ := logger.Get()
	return &Cleaner{
		manager: NewManager(),
		scanner: NewScanner(),
		logger:  lg,
	}
}

// CleanupOldDirectories menghapus direktori lama berdasarkan retention period
func (c *Cleaner) CleanupOldDirectories(path string, retentionDays int, namePattern string) (*CleanupResult, error) {
	if retentionDays <= 0 {
		return &CleanupResult{}, fmt.Errorf("retention days harus lebih dari 0")
	}

	normalizedPath := NormalizePath(path)
	if !c.manager.Dir().Exists(normalizedPath) {
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
			matched, err := filepath.Match(namePattern, entry.Name)
			if err != nil || !matched {
				continue
			}
		}

		// Skip jika belum melewati retention period
		if !entry.ModTime.Before(cutoffTime) {
			c.logger.Debug("Direktori belum melewati retention period",
				logger.String("dir", entry.Name),
				logger.Time("mod_time", entry.ModTime))
			continue
		}

		// Hitung ukuran sebelum dihapus
		size, err := c.manager.Dir().GetSize(entry.Path)
		if err != nil {
			c.logger.Warn("Gagal hitung ukuran direktori",
				logger.String("path", entry.Path),
				logger.Error(err))
		} else {
			result.TotalSizeFreed += size
		}

		// Hapus direktori
		if err := c.removeDirectory(entry.Path); err != nil {
			errorMsg := fmt.Sprintf("Gagal hapus %s: %v", entry.Name, err)
			result.Errors = append(result.Errors, errorMsg)
			c.logger.Error("Gagal hapus direktori",
				logger.String("path", entry.Path),
				logger.Error(err))
		} else {
			result.RemovedDirs = append(result.RemovedDirs, entry.Name)
			result.TotalRemoved++
			c.logger.Info("Direktori lama berhasil dihapus",
				logger.String("name", entry.Name),
				logger.String("path", entry.Path))
		}
	}

	return result, nil
}

// CleanupOldFiles menghapus file lama berdasarkan retention period
func (c *Cleaner) CleanupOldFiles(path string, retentionDays int, filePattern string) (*CleanupResult, error) {
	if retentionDays <= 0 {
		return &CleanupResult{}, fmt.Errorf("retention days harus lebih dari 0")
	}

	normalizedPath := NormalizePath(path)
	if !c.manager.Dir().Exists(normalizedPath) {
		return &CleanupResult{}, fmt.Errorf("direktori tidak ada: %s", normalizedPath)
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	result := &CleanupResult{
		RemovedDirs:  []string{},
		RemovedFiles: []string{},
		Errors:       []string{},
	}

	// Scan file untuk mencari kandidat penghapusan
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
		if filePattern != "" {
			matched, err := filepath.Match(filePattern, entry.Name)
			if err != nil || !matched {
				continue
			}
		}

		// Skip jika belum melewati retention period
		if !entry.ModTime.Before(cutoffTime) {
			continue
		}

		// Tambahkan ukuran file ke total yang akan dibebaskan
		result.TotalSizeFreed += entry.Size

		// Hapus file
		if err := c.removeFile(entry.Path); err != nil {
			errorMsg := fmt.Sprintf("Gagal hapus file %s: %v", entry.Name, err)
			result.Errors = append(result.Errors, errorMsg)
			c.logger.Error("Gagal hapus file",
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

	return result, nil
}

// removeDirectory menghapus direktori secara rekursif
func (c *Cleaner) removeDirectory(path string) error {
	return c.manager.fs.RemoveAll(path)
}

// removeFile menghapus file individual
func (c *Cleaner) removeFile(path string) error {
	return c.manager.fs.Remove(path)
}
