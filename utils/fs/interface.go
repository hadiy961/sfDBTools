package fs

import "os"

// DirectoryManager defines directory-level operations used across packages
type DirectoryManager interface {
	Create(path string) error
	CreateWithPermissions(path string, mode os.FileMode, owner, group string) error
	Validate(path string) error
	Exists(path string) bool
	IsDirectory(path string) bool
	IsWritable(path string) error
}

// FileManager defines file-level operations used across packages
type FileManager interface {
	EnsureParentDir(filePath string) error
	CopyFile(src, dst string, info os.FileInfo) error
	SetFilePermissions(filePath string, mode os.FileMode, owner, group string) error
}
