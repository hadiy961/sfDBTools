//go:build windows
// +build windows

package dir

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
	"unsafe"

	"sfDBTools/internal/logger"

	"golang.org/x/sys/windows"
)

var (
	advapi32                  = windows.NewLazySystemDLL("advapi32.dll")
	procGetNamedSecurityInfoW = advapi32.NewProc("GetNamedSecurityInfoW")
)

// setUnixOwnership tidak berlaku di Windows (stub)
func (m *Manager) setUnixOwnership(path, owner, group string) error {
	m.logger.Debug("Unix ownership tidak berlaku di Windows, diabaikan",
		logger.String("path", path),
		logger.String("owner", owner),
		logger.String("group", group))
	return nil
}

// SetPermissions sets directory permissions untuk Windows systems
func (m *Manager) SetPermissions(path string, mode os.FileMode, owner, group string) error {
	normalizedPath := filepath.Clean(path)

	if !m.Exists(normalizedPath) {
		return fmt.Errorf("direktori tidak ada: %s", normalizedPath)
	}

	// Di Windows, hanya set file mode (ownership diabaikan)
	if err := m.fs.Chmod(normalizedPath, mode); err != nil {
		return fmt.Errorf("gagal set permission untuk '%s': %w", normalizedPath, err)
	}

	m.logger.Info("Permission direktori berhasil diset (Windows)",
		logger.String("path", normalizedPath),
		logger.String("mode", mode.String()))

	if owner != "" || group != "" {
		m.logger.Debug("Windows ownership tidak didukung dalam implementasi sederhana ini",
			logger.String("owner", owner),
			logger.String("group", group))
	}

	return nil
}

// GetPermissions mendapatkan informasi permission direktori Windows
func (m *Manager) GetPermissions(path string) (*PermissionInfo, error) {
	normalizedPath := filepath.Clean(path)

	info, err := m.fs.Stat(normalizedPath)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan info file '%s': %w", normalizedPath, err)
	}

	permInfo := &PermissionInfo{
		Path: normalizedPath,
		Mode: info.Mode(),
	}

	// Coba dapatkan owner info menggunakan Windows API
	if owner, err := m.getWindowsOwner(normalizedPath); err == nil {
		permInfo.Owner = owner
	}

	return permInfo, nil
}

// getWindowsOwner mendapatkan owner file/direktori menggunakan Windows API
func (m *Manager) getWindowsOwner(path string) (string, error) {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}

	var ownerSid *windows.SID
	var securityDescriptor windows.Handle

	ret, _, _ := procGetNamedSecurityInfoW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		1, // SE_FILE_OBJECT
		1, // OWNER_SECURITY_INFORMATION
		uintptr(unsafe.Pointer(&ownerSid)),
		0, // group
		0, // dacl
		0, // sacl
		uintptr(unsafe.Pointer(&securityDescriptor)),
	)

	if ret != 0 {
		return "", fmt.Errorf("GetNamedSecurityInfo failed with error %d", ret)
	}
	defer windows.LocalFree(securityDescriptor)

	// Convert SID ke username
	account, domain, _, err := ownerSid.LookupAccount("")
	if err != nil {
		return "", fmt.Errorf("gagal lookup account: %w", err)
	}

	if domain != "" {
		return domain + "\\" + account, nil
	}
	return account, nil
}

// CanRead mengecek apakah current user dapat membaca direktori (Windows)
func (m *Manager) CanRead(path string) bool {
	// Simple check: coba buka file untuk read
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// CanWrite mengecek apakah current user dapat menulis ke direktori (Windows)
func (m *Manager) CanWrite(path string) bool {
	// Windows write check lebih kompleks, gunakan test file approach
	return m.IsWritable(path) == nil
}

// CanExecute mengecek apakah current user dapat execute/traverse direktori (Windows)
func (m *Manager) CanExecute(path string) bool {
	// Di Windows, jika bisa list directory contents berarti bisa execute
	_, err := os.ReadDir(path)
	return err == nil
}

// GetCurrentUser mendapatkan informasi user Windows yang sedang menjalankan program
func GetCurrentUser() (*UserInfo, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan current user: %w", err)
	}

	return &UserInfo{
		UID:      0, // Windows tidak menggunakan UID
		GID:      0, // Windows tidak menggunakan GID
		Username: currentUser.Username,
		Name:     currentUser.Name,
		HomeDir:  currentUser.HomeDir,
	}, nil
}

// IsRoot mengecek apakah program berjalan sebagai Administrator (Windows equivalent of root)
func IsRoot() bool {
	// Check jika running sebagai Administrator
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	return err == nil
}

// IsAdministrator adalah alias untuk IsRoot() yang lebih sesuai untuk Windows
func IsAdministrator() bool {
	return IsRoot()
}
