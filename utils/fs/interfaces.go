// Package fs menyediakan abstraksi filesystem yang aman dan dapat diuji
package fs

import (
	"io/fs"
	"os"
)

// FileSystem menyediakan operasi filesystem dasar
type FileSystem interface {
	// Operasi file
	Create(name string) (*os.File, error)
	Open(name string) (*os.File, error)
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	Remove(name string) error
	Stat(name string) (os.FileInfo, error)

	// Operasi direktori
	Mkdir(name string, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	RemoveAll(path string) error

	// File system traversal
	ReadDir(dirname string) ([]fs.DirEntry, error)
	Walk(root string, fn fs.WalkDirFunc) error
}

// FileOperations menyediakan operasi file tingkat tinggi
type FileOperations interface {
	Copy(src, dst string) error
	CopyWithInfo(src, dst string, info os.FileInfo) error
	Move(src, dst string) error
	EnsureDir(path string) error
	WriteJSON(path string, data interface{}) error
}

// DirectoryOperations menyediakan operasi direktori tingkat tinggi
type DirectoryOperations interface {
	Create(path string) error
	CreateWithPerms(path string, mode os.FileMode, owner, group string) error
	Exists(path string) bool
	IsWritable(path string) error
	GetSize(path string) (int64, error)
	GetDiskUsage(path string) (*DiskUsage, error)
}

// PermissionManager mengelola permission dan ownership
type PermissionManager interface {
	SetFilePerms(path string, mode os.FileMode, owner, group string) error
	SetDirPerms(path string, mode os.FileMode, owner, group string) error
	PreserveOwnership(path string, info os.FileInfo) error
}

// DiskUsage merepresentasikan informasi penggunaan disk
type DiskUsage struct {
	Path        string  `json:"path"`
	Total       int64   `json:"total"`
	Used        int64   `json:"used"`
	Free        int64   `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}
