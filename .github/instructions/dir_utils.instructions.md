# Struktur Paket utils/dir

## File Structure
```
utils/dir/
├── manager.go          // DirectoryManager utama
├── permissions.go      // Operasi permission direktori
├── scanner.go          // Operasi scan/list directory
├── cleanup.go          // Operasi cleanup directory
└── types.go           // Type definitions dan constants
```

## manager.go
```go
package dir

import (
    "fmt"
    "os"
    "path/filepath"
    "sfDBTools/internal/logger"
)

// Manager handles all directory operations
type Manager struct {
    logger *logger.Logger
}

// NewManager creates a new directory manager
func NewManager() *Manager {
    lg, _ := logger.Get()
    return &Manager{logger: lg}
}

// Core directory operations
func (m *Manager) Create(path string) error
func (m *Manager) CreateWithPermissions(path string, mode os.FileMode, owner, group string) error
func (m *Manager) Exists(path string) bool
func (m *Manager) IsDirectory(path string) bool
func (m *Manager) IsWritable(path string) error
func (m *Manager) Validate(path string) error
func (m *Manager) Remove(path string) error
func (m *Manager) RemoveAll(path string) error

// Convenience functions (backward compatibility)
func Create(path string) error
func Validate(path string) error
func Exists(path string) bool
```

## permissions.go
```go
package dir

// Permission-related operations
func (m *Manager) SetPermissions(path string, mode os.FileMode, owner, group string) error
func (m *Manager) GetPermissions(path string) (os.FileMode, error)
func (m *Manager) CheckWritePermission(path string) error

// Convenience functions
func SetPermissions(path string, mode os.FileMode, owner, group string) error
func CheckWritable(path string) error
```

## scanner.go
```go
package dir

// Scanner handles directory scanning and listing
type Scanner struct {
    manager *Manager
}

// NewScanner creates directory scanner
func NewScanner() *Scanner
func (s *Scanner) List(path string, filter ...FilterFunc) ([]Entry, error)
func (s *Scanner) Find(path string, pattern string) ([]string, error)
func (s *Scanner) FindByExtension(path string, ext string) ([]string, error)
func (s *Scanner) Walk(path string, walkFunc WalkFunc) error

// Filter functions
type FilterFunc func(entry Entry) bool
func FilterByExtension(ext string) FilterFunc
func FilterByPattern(pattern string) FilterFunc
func FilterByModTime(before, after time.Time) FilterFunc
```

## cleanup.go
```go
package dir

// Cleanup operations
func (m *Manager) CleanupOldDirectories(path string, retentionDays int, pattern string) ([]string, error)
func (m *Manager) CleanupOldFiles(path string, retentionDays int, pattern string) ([]string, error)
func (m *Manager) EmptyDirectory(path string) error

// Convenience functions
func CleanupOld(path string, retentionDays int) ([]string, error)
```

## types.go
```go
package dir

import (
    "os"
    "time"
)

// Entry represents a directory entry
type Entry struct {
    Name    string
    Path    string
    IsDir   bool
    Size    int64
    Mode    os.FileMode
    ModTime time.Time
}

// WalkFunc for directory walking
type WalkFunc func(path string, entry Entry, err error) error

// Options for directory operations
type CreateOptions struct {
    Mode  os.FileMode
    Owner string
    Group string
}

type ScanOptions struct {
    Recursive bool
    IncludeHidden bool
    Filter FilterFunc
}
```