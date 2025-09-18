package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sfDBTools/internal/logger"

	"github.com/spf13/afero"
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
	return FormatSize(e.Size)
}

// IsOlderThan mengecek apakah entry lebih tua dari durasi yang ditentukan
func (e *Entry) IsOlderThan(duration time.Duration) bool {
	return time.Since(e.ModTime) > duration
}

// FilterFunc adalah function type untuk filtering entries
type FilterFunc func(entry Entry) bool

// ScanOptions berisi opsi untuk scanning direktori
type ScanOptions struct {
	Recursive     bool       `json:"recursive"`
	IncludeHidden bool       `json:"include_hidden"`
	Filter        FilterFunc `json:"-"`
	MaxDepth      int        `json:"max_depth"` // 0 = unlimited
}

// Scanner handles directory scanning operations
type Scanner struct {
	fs     afero.Fs
	logger *logger.Logger
}

// NewScanner membuat scanner baru
func NewScanner() *Scanner {
	lg, _ := logger.Get()
	return &Scanner{
		fs:     afero.NewOsFs(),
		logger: lg,
	}
}

// NewScannerWithFs membuat scanner dengan filesystem custom
func NewScannerWithFs(fs afero.Fs) *Scanner {
	lg, _ := logger.Get()
	return &Scanner{
		fs:     fs,
		logger: lg,
	}
}

// List mengembalikan daftar entry dalam direktori
func (s *Scanner) List(path string, options ...ScanOptions) ([]Entry, error) {
	var opts ScanOptions
	if len(options) > 0 {
		opts = options[0]
	}

	normalizedPath := NormalizePath(path)

	// Pastikan direktori ada
	exists, err := afero.DirExists(s.fs, normalizedPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("direktori tidak ada: %s", normalizedPath)
	}

	var entries []Entry

	if opts.Recursive {
		err = s.walkRecursive(normalizedPath, &entries, opts, 0)
	} else {
		err = s.listSingle(normalizedPath, &entries, opts)
	}

	return entries, err
}

// listSingle melakukan list pada satu level direktori
func (s *Scanner) listSingle(path string, entries *[]Entry, opts ScanOptions) error {
	dirEntries, err := afero.ReadDir(s.fs, path)
	if err != nil {
		return err
	}

	for _, entry := range dirEntries {
		fsEntry := Entry{
			Name:     entry.Name(),
			Path:     filepath.Join(path, entry.Name()),
			IsDir:    entry.IsDir(),
			Size:     entry.Size(),
			Mode:     entry.Mode(),
			ModTime:  entry.ModTime(),
			IsHidden: IsHidden(entry.Name()),
		}

		// Apply filters
		if !opts.IncludeHidden && fsEntry.IsHidden {
			continue
		}

		if opts.Filter != nil && !opts.Filter(fsEntry) {
			continue
		}

		*entries = append(*entries, fsEntry)
	}

	return nil
}

// walkRecursive melakukan recursive walk
func (s *Scanner) walkRecursive(path string, entries *[]Entry, opts ScanOptions, currentDepth int) error {
	// Check depth limit
	if opts.MaxDepth > 0 && currentDepth >= opts.MaxDepth {
		return nil
	}

	// List current directory
	if err := s.listSingle(path, entries, opts); err != nil {
		return err
	}

	// Recurse into subdirectories
	dirEntries, err := afero.ReadDir(s.fs, path)
	if err != nil {
		return err
	}

	for _, entry := range dirEntries {
		if entry.IsDir() {
			if !opts.IncludeHidden && IsHidden(entry.Name()) {
				continue
			}

			subPath := filepath.Join(path, entry.Name())
			if err := s.walkRecursive(subPath, entries, opts, currentDepth+1); err != nil {
				s.logger.Warn("Error walking subdirectory",
					logger.String("path", subPath),
					logger.Error(err))
			}
		}
	}

	return nil
}

// Filter helper functions
func FilterByExtension(ext string) FilterFunc {
	return func(entry Entry) bool {
		if entry.IsDir {
			return true
		}
		return filepath.Ext(entry.Name) == ext
	}
}

func FilterByPattern(pattern string) FilterFunc {
	return func(entry Entry) bool {
		matched, err := filepath.Match(pattern, entry.Name)
		return err == nil && matched
	}
}

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

func FilterDirectoriesOnly() FilterFunc {
	return func(entry Entry) bool {
		return entry.IsDir
	}
}

func FilterFilesOnly() FilterFunc {
	return func(entry Entry) bool {
		return !entry.IsDir
	}
}

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
