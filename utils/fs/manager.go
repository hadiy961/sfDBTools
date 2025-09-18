// Package fs menyediakan implementasi filesystem yang terpusat
package fs

import (
	"os"
	"path/filepath"

	"sfDBTools/internal/logger"

	"github.com/spf13/afero"
)

// Manager menyediakan implementasi terpusat untuk semua operasi filesystem
type Manager struct {
	fs          afero.Fs
	logger      *logger.Logger
	fileOps     FileOperations
	dirOps      DirectoryOperations
	permMgr     PermissionManager
	checksumOps ChecksumOperations
	verifyOps   FileVerificationOperations
	dirValidOps DirectoryValidationOperations
	patternOps  PatternMatchingOperations
}

// NewManager membuat manager filesystem baru dengan real OS filesystem
func NewManager() *Manager {
	lg, _ := logger.Get()
	fs := afero.NewOsFs()

	m := &Manager{
		fs:     fs,
		logger: lg,
	}

	// Initialize sub-managers
	m.fileOps = newFileOperations(fs, lg)
	m.dirOps = newDirectoryOperations(fs, lg)
	m.permMgr = newPermissionManager(fs, lg)
	m.checksumOps = newChecksumOperations(lg)
	m.verifyOps = newFileVerificationOperations(lg)
	m.dirValidOps = newDirectoryValidationOperations(lg)
	m.patternOps = newPatternMatchingOperations(lg)

	return m
}

// NewManagerWithFs membuat manager dengan filesystem custom (untuk testing)
func NewManagerWithFs(fs afero.Fs) *Manager {
	lg, _ := logger.Get()

	m := &Manager{
		fs:     fs,
		logger: lg,
	}

	// Initialize sub-managers
	m.fileOps = newFileOperations(fs, lg)
	m.dirOps = newDirectoryOperations(fs, lg)
	m.permMgr = newPermissionManager(fs, lg)
	m.checksumOps = newChecksumOperations(lg)
	m.verifyOps = newFileVerificationOperations(lg)
	m.dirValidOps = newDirectoryValidationOperations(lg)
	m.patternOps = newPatternMatchingOperations(lg)

	return m
}

// File returns file operations interface
func (m *Manager) File() FileOperations {
	return m.fileOps
}

// Dir returns directory operations interface
func (m *Manager) Dir() DirectoryOperations {
	return m.dirOps
}

// Perm returns permission manager interface
func (m *Manager) Perm() PermissionManager {
	return m.permMgr
}

// Checksum returns checksum operations interface
func (m *Manager) Checksum() ChecksumOperations {
	return m.checksumOps
}

// Verify returns file verification operations interface
func (m *Manager) Verify() FileVerificationOperations {
	return m.verifyOps
}

// DirValid returns directory validation operations interface
func (m *Manager) DirValid() DirectoryValidationOperations {
	return m.dirValidOps
}

// Pattern returns pattern matching operations interface
func (m *Manager) Pattern() PatternMatchingOperations {
	return m.patternOps
}

// Convenience methods untuk backward compatibility
func (m *Manager) CopyFile(src, dst string, info os.FileInfo) error {
	return m.fileOps.CopyWithInfo(src, dst, info)
}

func (m *Manager) EnsureParentDir(filePath string) error {
	return m.fileOps.EnsureDir(filepath.Dir(filePath))
}

func (m *Manager) CreateDir(path string) error {
	return m.dirOps.Create(path)
}

func (m *Manager) DirExists(path string) bool {
	return m.dirOps.Exists(path)
}

func (m *Manager) SetFilePermissions(filePath string, mode os.FileMode, owner, group string) error {
	return m.permMgr.SetFilePerms(filePath, mode, owner, group)
}
