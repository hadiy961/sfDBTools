package file

import (
	"fmt"
	"os"
	"syscall"

	"sfDBTools/internal/logger"
)

// EnsureParentDir makes sure the parent directory of filePath exists.
// It creates parents with mode 0755 if necessary.
func EnsureParentDir(filePath string) error {
	m := NewManager()
	return m.EnsureParentDir(filePath)
}

// TestWrite attempts to create a small test file at filePath to verify writability.
// It uses the provided mode when creating the file and removes it afterwards.
func TestWrite(filePath string, mode os.FileMode) error {
	m := NewManager()
	if err := m.EnsureParentDir(filePath); err != nil {
		return err
	}

	f, err := m.fs.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("failed to create test file '%s': %w", filePath, err)
	}
	f.Close()
	if err := m.fs.Remove(filePath); err != nil {
		return fmt.Errorf("failed to cleanup test file '%s': %w", filePath, err)
	}
	return nil
}

// PreserveOwnership sets ownership of the given path using uid/gid from stat info if available.
// It is a noop on Windows.
func PreserveOwnership(path string, info os.FileInfo) {
	if info == nil {
		return
	}
	if statT, ok := info.Sys().(*syscall.Stat_t); ok {
		_ = os.Chown(path, int(statT.Uid), int(statT.Gid))
	}
}

// PreserveSymlinkOwnership attempts to set symlink ownership without following the link.
// On platforms that support lchown, it will call it. It's a noop on Windows.
func PreserveSymlinkOwnership(path string, info os.FileInfo) {
	if info == nil {
		return
	}
	if statT, ok := info.Sys().(*syscall.Stat_t); ok {
		// Use Lchown when available
		_ = os.Lchown(path, int(statT.Uid), int(statT.Gid))
	}
}

// SetPermissionsAndOwnership sets file permissions and ownership (non-recursive).
// Logs warnings instead of returning errors for non-fatal operations.
func SetPermissionsAndOwnership(path string, mode os.FileMode, info os.FileInfo) {
	// Use Manager to set permissions; ignore owner/group here
	if err := NewManager().SetFilePermissions(path, mode, "", ""); err != nil {
		lg, _ := logger.Get()
		lg.Warn("SetPermissionsAndOwnership: failed to set permissions via manager", logger.String("path", path), logger.Error(err))
	}
	PreserveOwnership(path, info)
}

// CopyFile copies a regular file from src to dst, preserving permissions and ownership when possible.
// It ensures the parent directory of dst exists.
func CopyFile(src, dst string, info os.FileInfo) error {
	m := NewManager()
	return m.CopyFile(src, dst, info)
}
