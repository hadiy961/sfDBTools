package dir

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Entry merepresentasikan entry dalam direktori
type Entry struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	IsDir    bool        `json:"is_dir"`
	Size     int64       `json:"size"`
	Mode     os.FileMode `json:"mode"`
	ModTime  time.Time   `json:"mod_time"`
	IsHidden bool        `json:"is_hidden"`
}

// GetFormattedSize mengembalikan ukuran dalam format yang mudah dibaca
func (e *Entry) GetFormattedSize() string {
	if e.IsDir {
		return "<DIR>"
	}
	return formatSize(e.Size)
}

// IsOlderThan mengecek apakah entry lebih tua dari durasi yang ditentukan
func (e *Entry) IsOlderThan(duration time.Duration) bool {
	return time.Since(e.ModTime) > duration
}

// DiskUsage merepresentasikan informasi penggunaan disk
type DiskUsage struct {
	Path        string  `json:"path"`
	CheckedPath string  `json:"checked_path"`
	Total       int64   `json:"total"`
	Used        int64   `json:"used"`
	Free        int64   `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

// GetFormattedTotal mengembalikan total size dalam format yang mudah dibaca
func (d *DiskUsage) GetFormattedTotal() string {
	return formatSize(d.Total)
}

// GetFormattedUsed mengembalikan used size dalam format yang mudah dibaca
func (d *DiskUsage) GetFormattedUsed() string {
	return formatSize(d.Used)
}

// GetFormattedFree mengembalikan free size dalam format yang mudah dibaca
func (d *DiskUsage) GetFormattedFree() string {
	return formatSize(d.Free)
}

// WalkFunc adalah function type untuk directory walking
type WalkFunc func(path string, entry Entry, err error) error

// FilterFunc adalah function type untuk filtering entries
type FilterFunc func(entry Entry) bool

// CreateOptions berisi opsi untuk pembuatan direktori
type CreateOptions struct {
	Mode  os.FileMode `json:"mode"`
	Owner string      `json:"owner"`
	Group string      `json:"group"`
}

// ScanOptions berisi opsi untuk scanning direktori
type ScanOptions struct {
	Recursive     bool       `json:"recursive"`
	IncludeHidden bool       `json:"include_hidden"`
	Filter        FilterFunc `json:"-"`         // Tidak bisa di-serialize
	MaxDepth      int        `json:"max_depth"` // 0 = unlimited
}

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
	return formatSize(c.TotalSizeFreed)
}

// FilterByExtension membuat filter berdasarkan ekstensi file
func FilterByExtension(ext string) FilterFunc {
	return func(entry Entry) bool {
		if entry.IsDir {
			return true // Selalu include direktori untuk traversal
		}
		return strings.HasSuffix(strings.ToLower(entry.Name), strings.ToLower(ext))
	}
}

// FilterByPattern membuat filter berdasarkan pattern (menggunakan filepath.Match)
func FilterByPattern(pattern string) FilterFunc {
	return func(entry Entry) bool {
		matched, err := filepath.Match(pattern, entry.Name)
		return err == nil && matched
	}
}

// FilterByModTime membuat filter berdasarkan modification time
func FilterByModTime(before, after time.Time) FilterFunc {
	return func(entry Entry) bool {
		if !before.IsZero() && entry.ModTime.After(before) {
			return false
		}
		if !after.IsZero() && entry.ModTime.Before(after) {
			return false
		}
		return true
	}
}

// FilterBySize membuat filter berdasarkan ukuran file
func FilterBySize(minSize, maxSize int64) FilterFunc {
	return func(entry Entry) bool {
		if entry.IsDir {
			return true
		}
		if minSize > 0 && entry.Size < minSize {
			return false
		}
		if maxSize > 0 && entry.Size > maxSize {
			return false
		}
		return true
	}
}

// FilterHiddenFiles membuat filter untuk exclude hidden files
func FilterHiddenFiles() FilterFunc {
	return func(entry Entry) bool {
		return !entry.IsHidden
	}
}

// FilterDirectoriesOnly membuat filter hanya untuk direktori
func FilterDirectoriesOnly() FilterFunc {
	return func(entry Entry) bool {
		return entry.IsDir
	}
}

// FilterFilesOnly membuat filter hanya untuk file
func FilterFilesOnly() FilterFunc {
	return func(entry Entry) bool {
		return !entry.IsDir
	}
}

// CombineFilters menggabungkan multiple filter dengan AND logic
func CombineFilters(filters ...FilterFunc) FilterFunc {
	return func(entry Entry) bool {
		for _, filter := range filters {
			if !filter(entry) {
				return false
			}
		}
		return true
	}
}

// formatSize memformat ukuran byte menjadi string yang mudah dibaca
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// isHidden mengecek apakah file/direktori hidden berdasarkan platform
func isHidden(name string) bool {
	if runtime.GOOS == "windows" {
		// Di Windows, bisa menggunakan syscall untuk cek attribute
		// Untuk sekarang, anggap file yang dimulai dengan '.' sebagai hidden
		return strings.HasPrefix(name, ".")
	}
	// Di Unix-like systems, file hidden dimulai dengan '.'
	return strings.HasPrefix(name, ".")
}
