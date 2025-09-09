package dir

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sfDBTools/internal/logger"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/afero"
)

// Scanner handles directory scanning and listing operations
type Scanner struct {
	fs     afero.Fs
	logger *logger.Logger
}

// NewScanner creates a new directory scanner
func NewScanner() *Scanner {
	lg, _ := logger.Get()
	return &Scanner{
		fs:     afero.NewOsFs(),
		logger: lg,
	}
}

// NewScannerWithFs creates a scanner with custom filesystem (untuk testing)
func NewScannerWithFs(fs afero.Fs) *Scanner {
	lg, _ := logger.Get()
	return &Scanner{
		fs:     fs,
		logger: lg,
	}
}

// List mengembalikan list entry dalam direktori dengan opsi filtering
func (s *Scanner) List(path string, options ...ScanOptions) ([]Entry, error) {
	if path == "" {
		return nil, fmt.Errorf("path tidak boleh kosong")
	}

	normalizedPath := filepath.Clean(path)

	// Pastikan direktori ada
	exists, err := afero.DirExists(s.fs, normalizedPath)
	if err != nil {
		return nil, fmt.Errorf("gagal mengecek direktori '%s': %w", normalizedPath, err)
	}
	if !exists {
		return nil, fmt.Errorf("direktori tidak ada: %s", normalizedPath)
	}

	// Parse options
	opts := ScanOptions{
		Recursive:     false,
		IncludeHidden: false,
		MaxDepth:      1,
	}
	if len(options) > 0 {
		opts = options[0]
		if opts.MaxDepth == 0 {
			opts.MaxDepth = -1 // unlimited depth
		}
	}

	var entries []Entry

	if opts.Recursive {
		err = s.walkRecursive(normalizedPath, opts, &entries, 0)
	} else {
		err = s.listSingle(normalizedPath, opts, &entries)
	}

	if err != nil {
		return nil, fmt.Errorf("gagal scan direktori '%s': %w", normalizedPath, err)
	}

	s.logger.Debug("Directory scan completed",
		logger.String("path", normalizedPath),
		logger.Int("entries_found", len(entries)),
		logger.Bool("recursive", opts.Recursive))

	return entries, nil
}

// Find mencari file/direktori berdasarkan pattern glob
func (s *Scanner) Find(path string, pattern string) ([]string, error) {
	if path == "" || pattern == "" {
		return nil, fmt.Errorf("path dan pattern tidak boleh kosong")
	}

	normalizedPath := filepath.Clean(path)

	// Gunakan doublestar untuk advanced glob matching
	fullPattern := filepath.Join(normalizedPath, pattern)
	matches, err := doublestar.Glob(afero.NewIOFS(s.fs), fullPattern)
	if err != nil {
		return nil, fmt.Errorf("gagal mencari dengan pattern '%s': %w", pattern, err)
	}

	// Convert matches menjadi absolute paths
	var results []string
	for _, match := range matches {
		absolutePath := match
		if !filepath.IsAbs(absolutePath) {
			absolutePath = filepath.Join(normalizedPath, match)
		}
		results = append(results, filepath.Clean(absolutePath))
	}

	s.logger.Debug("Pattern search completed",
		logger.String("path", normalizedPath),
		logger.String("pattern", pattern),
		logger.Int("matches", len(results)))

	return results, nil
}

// FindByExtension mencari file berdasarkan ekstensi
func (s *Scanner) FindByExtension(path string, ext string) ([]string, error) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	pattern := "**/*" + ext
	return s.Find(path, pattern)
}

// FindByName mencari file/direktori berdasarkan nama (mendukung wildcard)
func (s *Scanner) FindByName(path string, name string) ([]string, error) {
	pattern := "**/" + name
	return s.Find(path, pattern)
}

// Walk melakukan traversal direktori dengan custom function
func (s *Scanner) Walk(path string, walkFunc WalkFunc) error {
	if path == "" {
		return fmt.Errorf("path tidak boleh kosong")
	}
	if walkFunc == nil {
		return fmt.Errorf("walkFunc tidak boleh nil")
	}

	normalizedPath := filepath.Clean(path)

	return afero.Walk(s.fs, normalizedPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return walkFunc(filePath, Entry{}, err)
		}

		entry := Entry{
			Name:     info.Name(),
			Path:     filePath,
			IsDir:    info.IsDir(),
			Size:     info.Size(),
			Mode:     info.Mode(),
			ModTime:  info.ModTime(),
			IsHidden: isHidden(info.Name()),
		}

		return walkFunc(filePath, entry, nil)
	})
}

// CountEntries menghitung jumlah file dan direktori
func (s *Scanner) CountEntries(path string, recursive bool) (int, int, error) {
	fileCount := 0
	dirCount := 0

	options := ScanOptions{
		Recursive:     recursive,
		IncludeHidden: false,
	}

	entries, err := s.List(path, options)
	if err != nil {
		return 0, 0, err
	}

	for _, entry := range entries {
		if entry.IsDir {
			dirCount++
		} else {
			fileCount++
		}
	}

	return fileCount, dirCount, nil
}

// GetLargestFiles mencari file terbesar dalam direktori
func (s *Scanner) GetLargestFiles(path string, count int, recursive bool) ([]Entry, error) {
	options := ScanOptions{
		Recursive:     recursive,
		IncludeHidden: false,
		Filter:        FilterFilesOnly(), // Hanya file
	}

	entries, err := s.List(path, options)
	if err != nil {
		return nil, err
	}

	// Sort berdasarkan size (descending)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Size > entries[j].Size
	})

	// Return top N entries
	if count > len(entries) {
		count = len(entries)
	}

	return entries[:count], nil
}

// GetOldestFiles mencari file terlama dalam direktori
func (s *Scanner) GetOldestFiles(path string, count int, recursive bool) ([]Entry, error) {
	options := ScanOptions{
		Recursive:     recursive,
		IncludeHidden: false,
		Filter:        FilterFilesOnly(), // Hanya file
	}

	entries, err := s.List(path, options)
	if err != nil {
		return nil, err
	}

	// Sort berdasarkan ModTime (ascending - oldest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ModTime.Before(entries[j].ModTime)
	})

	// Return top N entries
	if count > len(entries) {
		count = len(entries)
	}

	return entries[:count], nil
}

// listSingle melakukan list direktori tunggal (non-recursive)
func (s *Scanner) listSingle(path string, opts ScanOptions, entries *[]Entry) error {
	files, err := afero.ReadDir(s.fs, path)
	if err != nil {
		return err
	}

	for _, file := range files {
		entry := Entry{
			Name:     file.Name(),
			Path:     filepath.Join(path, file.Name()),
			IsDir:    file.IsDir(),
			Size:     file.Size(),
			Mode:     file.Mode(),
			ModTime:  file.ModTime(),
			IsHidden: isHidden(file.Name()),
		}

		// Skip hidden files jika tidak diminta
		if !opts.IncludeHidden && entry.IsHidden {
			continue
		}

		// Apply filter jika ada
		if opts.Filter != nil && !opts.Filter(entry) {
			continue
		}

		*entries = append(*entries, entry)
	}

	return nil
}

// walkRecursive melakukan recursive directory traversal
func (s *Scanner) walkRecursive(path string, opts ScanOptions, entries *[]Entry, currentDepth int) error {
	// Cek depth limit
	if opts.MaxDepth > 0 && currentDepth >= opts.MaxDepth {
		return nil
	}

	files, err := afero.ReadDir(s.fs, path)
	if err != nil {
		return err
	}

	for _, file := range files {
		entry := Entry{
			Name:     file.Name(),
			Path:     filepath.Join(path, file.Name()),
			IsDir:    file.IsDir(),
			Size:     file.Size(),
			Mode:     file.Mode(),
			ModTime:  file.ModTime(),
			IsHidden: isHidden(file.Name()),
		}

		// Skip hidden files jika tidak diminta
		if !opts.IncludeHidden && entry.IsHidden {
			continue
		}

		// Apply filter jika ada
		if opts.Filter != nil && !opts.Filter(entry) {
			continue
		}

		*entries = append(*entries, entry)

		// Recursive ke subdirectory
		if entry.IsDir {
			err := s.walkRecursive(entry.Path, opts, entries, currentDepth+1)
			if err != nil {
				s.logger.Warn("Error walking subdirectory",
					logger.String("path", entry.Path),
					logger.Error(err))
				// Continue dengan direktori lain, tidak gagalkan keseluruhan
			}
		}
	}

	return nil
}
