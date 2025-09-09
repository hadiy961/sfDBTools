// Package dir menyediakan operasi direktori multi-platform dengan abstraksi filesystem
package dir

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"sfDBTools/internal/logger"

	gopsutildisk "github.com/shirou/gopsutil/v3/disk"
	"github.com/spf13/afero"
)

// Manager handles all directory operations with cross-platform support
type Manager struct {
	fs     afero.Fs
	logger *logger.Logger
}

// NewManager creates a new multi-platform directory manager
func NewManager() *Manager {
	lg, _ := logger.Get()
	return &Manager{
		fs:     afero.NewOsFs(), // Real filesystem
		logger: lg,
	}
}

// NewManagerWithFs creates a manager with custom filesystem (untuk testing)
func NewManagerWithFs(fs afero.Fs) *Manager {
	lg, _ := logger.Get()
	return &Manager{
		fs:     fs,
		logger: lg,
	}
}

// Create membuat direktori jika belum ada dengan validasi lengkap
func (m *Manager) Create(path string) error {
	if err := m.validatePath(path); err != nil {
		return err
	}

	// Normalize path untuk cross-platform compatibility
	normalizedPath := filepath.Clean(path)

	// Cek apakah sudah exists
	exists, err := afero.Exists(m.fs, normalizedPath)
	if err != nil {
		m.logger.Error("Gagal mengecek eksistensi direktori",
			logger.String("path", normalizedPath),
			logger.Error(err))
		return fmt.Errorf("gagal mengecek direktori '%s': %w", normalizedPath, err)
	}

	if !exists {
		// Buat direktori dengan permission sesuai platform
		mode := m.getDefaultDirMode()
		if err := m.fs.MkdirAll(normalizedPath, mode); err != nil {
			m.logger.Error("Gagal membuat direktori",
				logger.String("path", normalizedPath),
				logger.String("mode", mode.String()),
				logger.Error(err))
			return fmt.Errorf("gagal membuat direktori '%s': %w", normalizedPath, err)
		}
		m.logger.Info("Direktori berhasil dibuat",
			logger.String("path", normalizedPath),
			logger.String("mode", mode.String()))
	}

	return m.Validate(normalizedPath)
}

// CreateWithPermissions membuat direktori dengan permission dan ownership khusus
func (m *Manager) CreateWithPermissions(path string, mode os.FileMode, owner, group string) error {
	if err := m.validatePath(path); err != nil {
		return err
	}

	normalizedPath := filepath.Clean(path)

	// Buat direktori dengan mode yang ditentukan
	if err := m.fs.MkdirAll(normalizedPath, mode); err != nil {
		return fmt.Errorf("gagal membuat direktori '%s' dengan permission %s: %w", normalizedPath, mode.String(), err)
	}

	// Set ownership untuk Unix-like systems
	if runtime.GOOS != "windows" && (owner != "" || group != "") {
		if err := m.setUnixOwnership(normalizedPath, owner, group); err != nil {
			m.logger.Warn("Gagal set ownership, direktori tetap dibuat",
				logger.String("path", normalizedPath),
				logger.String("owner", owner),
				logger.String("group", group),
				logger.Error(err))
		}
	}

	m.logger.Info("Direktori dibuat dengan permission khusus",
		logger.String("path", normalizedPath),
		logger.String("mode", mode.String()),
		logger.String("owner", owner),
		logger.String("group", group))

	return nil
}

// Exists mengecek apakah direktori ada
func (m *Manager) Exists(path string) bool {
	if path == "" {
		return false
	}

	normalizedPath := filepath.Clean(path)
	exists, err := afero.DirExists(m.fs, normalizedPath)
	if err != nil {
		m.logger.Debug("Error checking directory existence",
			logger.String("path", normalizedPath),
			logger.Error(err))
		return false
	}

	return exists
}

// IsDirectory mengecek apakah path adalah direktori yang valid
func (m *Manager) IsDirectory(path string) bool {
	if path == "" {
		return false
	}

	normalizedPath := filepath.Clean(path)
	info, err := m.fs.Stat(normalizedPath)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// IsWritable mengecek apakah direktori dapat ditulis
func (m *Manager) IsWritable(path string) error {
	if !m.Exists(path) {
		return fmt.Errorf("direktori tidak ada: %s", path)
	}

	normalizedPath := filepath.Clean(path)

	// Buat test file untuk cek write permission
	testFileName := m.generateTestFileName()
	testPath := filepath.Join(normalizedPath, testFileName)

	// Coba buat file test
	testFile, err := m.fs.Create(testPath)
	if err != nil {
		m.logger.Debug("Direktori tidak dapat ditulis",
			logger.String("path", normalizedPath),
			logger.Error(err))
		return fmt.Errorf("direktori '%s' tidak dapat ditulis: %w", normalizedPath, err)
	}

	testFile.Close()

	// Hapus test file
	if err := m.fs.Remove(testPath); err != nil {
		m.logger.Warn("Gagal menghapus test file",
			logger.String("test_file", testPath),
			logger.Error(err))
	}

	m.logger.Debug("Direktori dapat ditulis", logger.String("path", normalizedPath))
	return nil
}

// Validate memastikan direktori ada dan dapat ditulis
func (m *Manager) Validate(path string) error {
	if err := m.validatePath(path); err != nil {
		return err
	}

	normalizedPath := filepath.Clean(path)

	// Pastikan direktori ada
	if !m.Exists(normalizedPath) {
		return fmt.Errorf("direktori tidak ada: %s", normalizedPath)
	}

	// Pastikan ini adalah direktori, bukan file
	if !m.IsDirectory(normalizedPath) {
		return fmt.Errorf("path '%s' bukan direktori", normalizedPath)
	}

	// Cek write permission
	if err := m.IsWritable(normalizedPath); err != nil {
		return err
	}

	return nil
}

// Remove menghapus direktori kosong
func (m *Manager) Remove(path string) error {
	if path == "" {
		return fmt.Errorf("path tidak boleh kosong")
	}

	normalizedPath := filepath.Clean(path)

	if !m.Exists(normalizedPath) {
		m.logger.Debug("Direktori tidak ada, tidak perlu dihapus",
			logger.String("path", normalizedPath))
		return nil
	}

	if err := m.fs.Remove(normalizedPath); err != nil {
		m.logger.Error("Gagal menghapus direktori",
			logger.String("path", normalizedPath),
			logger.Error(err))
		return fmt.Errorf("gagal menghapus direktori '%s': %w", normalizedPath, err)
	}

	m.logger.Info("Direktori berhasil dihapus", logger.String("path", normalizedPath))
	return nil
}

// RemoveAll menghapus direktori beserta isinya secara rekursif
func (m *Manager) RemoveAll(path string) error {
	if path == "" {
		return fmt.Errorf("path tidak boleh kosong")
	}

	normalizedPath := filepath.Clean(path)

	if !m.Exists(normalizedPath) {
		m.logger.Debug("Direktori tidak ada, tidak perlu dihapus",
			logger.String("path", normalizedPath))
		return nil
	}

	if err := m.fs.RemoveAll(normalizedPath); err != nil {
		m.logger.Error("Gagal menghapus direktori rekursif",
			logger.String("path", normalizedPath),
			logger.Error(err))
		return fmt.Errorf("gagal menghapus direktori '%s' secara rekursif: %w", normalizedPath, err)
	}

	m.logger.Info("Direktori dan isinya berhasil dihapus", logger.String("path", normalizedPath))
	return nil
}

// GetSize menghitung total ukuran direktori
func (m *Manager) GetSize(path string) (int64, error) {
	if !m.Exists(path) {
		return 0, fmt.Errorf("direktori tidak ada: %s", path)
	}

	normalizedPath := filepath.Clean(path)
	var totalSize int64

	err := afero.Walk(m.fs, normalizedPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("gagal menghitung ukuran direktori '%s': %w", normalizedPath, err)
	}

	return totalSize, nil
}

// GetDiskUsage mengembalikan informasi penggunaan disk untuk path
func (m *Manager) GetDiskUsage(path string) (*DiskUsage, error) {
	normalizedPath := filepath.Clean(path)

	// Cari path yang ada untuk dicek
	checkPath := m.findExistingPath(normalizedPath)

	usage, err := gopsutildisk.Usage(checkPath)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan info disk usage untuk '%s': %w", checkPath, err)
	}

	diskUsage := &DiskUsage{
		Path:        normalizedPath,
		CheckedPath: checkPath,
		Total:       int64(usage.Total),
		Used:        int64(usage.Used),
		Free:        int64(usage.Free),
		UsedPercent: usage.UsedPercent,
	}

	return diskUsage, nil
}

// CheckDiskSpace mengecek apakah tersedia ruang disk minimal (dalam MB)
func (m *Manager) CheckDiskSpace(path string, minFreeMB int64) error {
	diskUsage, err := m.GetDiskUsage(path)
	if err != nil {
		return err
	}

	requiredBytes := minFreeMB * 1024 * 1024 // Convert MB to bytes

	if diskUsage.Free < requiredBytes {
		return fmt.Errorf("ruang disk tidak mencukupi: tersedia %d MB, dibutuhkan %d MB",
			diskUsage.Free/1024/1024, minFreeMB)
	}

	m.logger.Debug("Pengecekan ruang disk berhasil",
		logger.String("path", path),
		logger.Int64("available_mb", diskUsage.Free/1024/1024),
		logger.Int64("required_mb", minFreeMB))

	return nil
}

// validatePath melakukan validasi dasar path
func (m *Manager) validatePath(path string) error {
	if path == "" {
		m.logger.Error("Path direktori tidak boleh kosong")
		return fmt.Errorf("path direktori tidak boleh kosong")
	}

	// Cek karakter yang tidak valid di Windows
	if runtime.GOOS == "windows" {
		invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
		for _, char := range invalidChars {
			if strings.Contains(path, char) {
				return fmt.Errorf("path mengandung karakter tidak valid untuk Windows: %s", char)
			}
		}
	}

	return nil
}

// getDefaultDirMode mengembalikan permission default sesuai platform
func (m *Manager) getDefaultDirMode() os.FileMode {
	if runtime.GOOS == "windows" {
		return 0755 // Windows tidak terlalu peduli dengan file mode
	}
	return 0755 // Unix default: owner rwx, group rx, other rx
}

// generateTestFileName membuat nama file test yang unik
func (m *Manager) generateTestFileName() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf(".sfdbtools_write_test_%d", timestamp)
}

// findExistingPath mencari parent path yang ada untuk pengecekan disk usage
func (m *Manager) findExistingPath(path string) string {
	checkPath := path

	for {
		if exists, _ := afero.Exists(m.fs, checkPath); exists {
			return checkPath
		}

		parent := filepath.Dir(checkPath)
		if parent == checkPath || parent == "." {
			// Reached root, return system root
			if runtime.GOOS == "windows" {
				return filepath.VolumeName(checkPath) + string(os.PathSeparator)
			}
			return string(os.PathSeparator)
		}
		checkPath = parent
	}
}

// Backward compatibility functions
func Create(path string) error {
	manager := NewManager()
	return manager.Create(path)
}

func Validate(path string) error {
	manager := NewManager()
	return manager.Validate(path)
}

func Exists(path string) bool {
	manager := NewManager()
	return manager.Exists(path)
}
