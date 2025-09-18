// Package fs - Backward compatibility layer
// File ini menyediakan fungsi-fungsi untuk backward compatibility
// dengan kode existing yang menggunakan interface lama
package fs

import (
	"os"
)

// Global manager instance untuk backward compatibility
var globalManager = NewManager()

// Deprecated: Use `Manager.File().CopyWithInfo()` instead. This wrapper
// preserves the original signature which accepts an `os.FileInfo`.
func CopyFile(src, dst string, info os.FileInfo) error {
	return globalManager.File().CopyWithInfo(src, dst, info)
}

// Deprecated: Use Manager.File().EnsureDir() instead
func EnsureParentDir(filePath string) error {
	return globalManager.EnsureParentDir(filePath)
}

// Deprecated: Use Manager.Dir().Create() instead
func CreateDir(path string) error {
	return globalManager.Dir().Create(path)
}

// Deprecated: Use Manager.Dir().Exists() instead
func DirExists(path string) bool {
	return globalManager.Dir().Exists(path)
}

// Deprecated: Use Manager.File().WriteJSON() instead
func WriteJSON(filePath string, data interface{}) error {
	return globalManager.File().WriteJSON(filePath, data)
}

// Deprecated: Use Manager.Perm().SetFilePerms() instead
func SetFilePermissions(filePath string, mode os.FileMode, owner, group string) error {
	return globalManager.Perm().SetFilePerms(filePath, mode, owner, group)
}

// Deprecated: Use `Manager.Perm().PreserveOwnership()` instead.
// Note: this compatibility wrapper intentionally ignores the returned
// error to preserve the original legacy behavior where callers did not
// handle an error value.
func PreserveOwnership(path string, info os.FileInfo) {
	_ = globalManager.Perm().PreserveOwnership(path, info)
}

// Deprecated: Use `Manager.Perm().PreserveOwnership()` (or a dedicated
// lchown-capable function) for symlink ownership preservation.
// This wrapper provides a simplified fallback for legacy callers and
// ignores the returned error.
func PreserveSymlinkOwnership(path string, info os.FileInfo) {
	// Simplified implementation for backward compatibility
	_ = globalManager.Perm().PreserveOwnership(path, info)
}

// Deprecated: Use `Manager.Perm().SetFilePerms()` followed by
// `Manager.Perm().PreserveOwnership()` instead. This wrapper applies both
// operations and ignores returned errors to remain backward compatible.
func SetPermissionsAndOwnership(path string, mode os.FileMode, info os.FileInfo) {
	_ = globalManager.Perm().SetFilePerms(path, mode, "", "")
	_ = globalManager.Perm().PreserveOwnership(path, info)
}

// TestWrite untuk testing writability
func TestWrite(filePath string, mode os.FileMode) error {
	if err := globalManager.File().EnsureDir(GetDirectory(filePath)); err != nil {
		return err
	}

	f, err := globalManager.fs.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	f.Close()

	return globalManager.fs.Remove(filePath)
}

// Helper functions
func GetDirectory(filePath string) string {
	dir, _ := SplitPath(filePath)
	return dir
}
