package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"sfDBTools/internal/logger"

	gopsutildisk "github.com/shirou/gopsutil/v3/disk"
	"github.com/spf13/afero"
)

// directoryOperations mengimplementasikan DirectoryOperations interface
type directoryOperations struct {
	fs     afero.Fs
	logger *logger.Logger
}

// newDirectoryOperations membuat instance directory operations baru
func newDirectoryOperations(fs afero.Fs, logger *logger.Logger) DirectoryOperations {
	return &directoryOperations{
		fs:     fs,
		logger: logger,
	}
}

// Create membuat direktori jika belum ada
func (d *directoryOperations) Create(path string) error {
	if path == "" {
		return fmt.Errorf("path tidak boleh kosong")
	}

	normalizedPath := filepath.Clean(path)

	exists, err := afero.Exists(d.fs, normalizedPath)
	if err != nil {
		return fmt.Errorf("gagal cek eksistensi dir %s: %w", normalizedPath, err)
	}

	if !exists {
		mode := d.getDefaultDirMode()
		if err := d.fs.MkdirAll(normalizedPath, mode); err != nil {
			return fmt.Errorf("gagal buat dir %s: %w", normalizedPath, err)
		}
		d.logger.Debug("Direktori berhasil dibuat",
			logger.String("path", normalizedPath),
			logger.String("mode", mode.String()))
	}

	return d.validateDirectory(normalizedPath)
}

// CreateWithPerms membuat direktori dengan permission dan ownership khusus
func (d *directoryOperations) CreateWithPerms(path string, mode os.FileMode, owner, group string) error {
	if path == "" {
		return fmt.Errorf("path tidak boleh kosong")
	}

	normalizedPath := filepath.Clean(path)

	if err := d.fs.MkdirAll(normalizedPath, mode); err != nil {
		return fmt.Errorf("gagal buat dir dengan permission khusus %s: %w", normalizedPath, err)
	}

	// Set ownership untuk Unix-like systems
	if runtime.GOOS != "windows" && (owner != "" || group != "") {
		if err := d.setUnixOwnership(normalizedPath, owner, group); err != nil {
			d.logger.Warn("Gagal set ownership",
				logger.String("path", normalizedPath),
				logger.Error(err))
		}
	}

	d.logger.Debug("Direktori dibuat dengan permission khusus",
		logger.String("path", normalizedPath),
		logger.String("mode", mode.String()))

	return nil
}

// Exists mengecek apakah direktori ada
func (d *directoryOperations) Exists(path string) bool {
	if path == "" {
		return false
	}

	normalizedPath := filepath.Clean(path)
	exists, err := afero.DirExists(d.fs, normalizedPath)
	if err != nil {
		d.logger.Debug("Error checking directory existence",
			logger.String("path", normalizedPath),
			logger.Error(err))
		return false
	}

	return exists
}

// IsWritable mengecek apakah direktori dapat ditulis
func (d *directoryOperations) IsWritable(path string) error {
	if !d.Exists(path) {
		return fmt.Errorf("direktori tidak ada: %s", path)
	}

	normalizedPath := filepath.Clean(path)
	testFileName := d.generateTestFileName()
	testPath := filepath.Join(normalizedPath, testFileName)

	testFile, err := d.fs.Create(testPath)
	if err != nil {
		return fmt.Errorf("direktori '%s' tidak dapat ditulis: %w", normalizedPath, err)
	}
	testFile.Close()

	if err := d.fs.Remove(testPath); err != nil {
		d.logger.Warn("Gagal hapus test file", logger.String("file", testPath), logger.Error(err))
	}

	return nil
}

// GetSize menghitung total ukuran direktori
func (d *directoryOperations) GetSize(path string) (int64, error) {
	if !d.Exists(path) {
		return 0, fmt.Errorf("direktori tidak ada: %s", path)
	}

	var totalSize int64
	err := afero.Walk(d.fs, path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	return totalSize, err
}

// GetDiskUsage mengembalikan informasi penggunaan disk untuk path
func (d *directoryOperations) GetDiskUsage(path string) (*DiskUsage, error) {
	existingPath := d.findExistingPath(path)

	usage, err := gopsutildisk.Usage(existingPath)
	if err != nil {
		return nil, fmt.Errorf("gagal dapatkan disk usage untuk %s: %w", existingPath, err)
	}

	result := &DiskUsage{
		Path:        path,
		Total:       int64(usage.Total),
		Used:        int64(usage.Used),
		Free:        int64(usage.Free),
		UsedPercent: usage.UsedPercent,
	}

	d.logger.Debug("Disk usage berhasil diperoleh",
		logger.String("path", path),
		logger.String("existing_path", existingPath))

	return result, nil
}

// Helper methods
func (d *directoryOperations) validateDirectory(path string) error {
	if !d.Exists(path) {
		return fmt.Errorf("direktori tidak ada: %s", path)
	}

	info, err := d.fs.Stat(path)
	if err != nil {
		return fmt.Errorf("gagal stat direktori %s: %w", path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path '%s' bukan direktori", path)
	}

	return d.IsWritable(path)
}

func (d *directoryOperations) getDefaultDirMode() os.FileMode {
	if runtime.GOOS == "windows" {
		return 0755
	}
	return 0755
}

func (d *directoryOperations) generateTestFileName() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf(".sfdbtools_write_test_%d", timestamp)
}

func (d *directoryOperations) findExistingPath(path string) string {
	checkPath := path

	for {
		if exists, _ := afero.Exists(d.fs, checkPath); exists {
			return checkPath
		}

		parent := filepath.Dir(checkPath)
		if parent == checkPath || parent == "." {
			if runtime.GOOS == "windows" {
				return filepath.VolumeName(checkPath) + string(os.PathSeparator)
			}
			return string(os.PathSeparator)
		}
		checkPath = parent
	}
}

// setUnixOwnership untuk Unix-like systems (implementasi sederhana)
func (d *directoryOperations) setUnixOwnership(path, owner, group string) error {
	// Implementasi sederhana - bisa diperluas sesuai kebutuhan
	// Untuk sekarang hanya log warning karena kompleksitas implementasi
	d.logger.Warn("Unix ownership setting belum diimplementasi secara lengkap",
		logger.String("path", path),
		logger.String("owner", owner),
		logger.String("group", group))
	return nil
}
