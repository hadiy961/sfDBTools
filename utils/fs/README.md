# Package fs - Filesystem Utilities

Package `fs` menyediakan abstraksi filesystem yang aman, dapat diuji, dan modular untuk sfDBTools.

## Overview

Package ini telah direfactor untuk menjadi lebih modular dan mengikuti best practices Go:

- **Centralized API**: Semua operasi filesystem melalui satu interface terpusat
- **Separation of Concerns**: Operasi file, direktori, dan permission dipisah
- **Testability**: Menggunakan `afero` untuk abstraksi filesystem
- **Backward Compatibility**: Mempertahankan fungsi-fungsi lama untuk compatibility
- **Cross-Platform**: Mendukung Unix dan Windows

## Architecture

```
fs/
├── interfaces.go      # Interface definitions
├── manager.go         # Main manager dengan unified API
├── file_ops.go        # File operations implementation
├── dir_ops.go         # Directory operations implementation  
├── perm_ops.go        # Permission management implementation
├── scanner.go         # Directory scanning utilities
├── cleaner.go         # Cleanup utilities
├── utils.go           # Helper utilities
└── compat.go          # Backward compatibility layer
```

## Usage

### Basic Usage

```go
import "sfDBTools/utils/fs"

// Create manager
manager := fs.NewManager()

// Directory operations
err := manager.Dir().Create("/path/to/dir")
exists := manager.Dir().Exists("/path/to/dir")
err := manager.Dir().IsWritable("/path/to/dir")

// File operations  
err := manager.File().Copy("src.txt", "dst.txt")
err := manager.File().WriteJSON("config.json", data)

// Permission operations
err := manager.Perm().SetFilePerms("/path/file", 0644, "user", "group")
```

### Backward Compatibility

Fungsi-fungsi lama masih tersedia untuk backward compatibility:

```go
import "sfDBTools/utils/fs"

// Legacy functions (deprecated but working)
err := fs.CreateDir("/path/to/dir")
exists := fs.DirExists("/path/to/dir") 
err := fs.CopyFile("src", "dst", info)
```

### Advanced Usage

```go
// Custom filesystem (for testing)
memFs := afero.NewMemMapFs()
manager := fs.NewManagerWithFs(memFs)

// Directory scanning
scanner := fs.NewScanner()
entries, err := scanner.List("/path", fs.ScanOptions{
    Recursive: true,
    Filter: fs.FilterByExtension(".log"),
})

// Cleanup operations
cleaner := fs.NewCleaner()
result, err := cleaner.CleanupOldFiles("/logs", 30, "*.log")
```

## Key Features

### 1. **Unified API**
- Single entry point through `Manager`
- Consistent error handling
- Integrated logging

### 2. **Modular Design**
- `FileOperations`: Copy, move, JSON writing
- `DirectoryOperations`: Create, validate, disk usage
- `PermissionManager`: Cross-platform permission handling

### 3. **Advanced Utilities**
- `Scanner`: Directory traversal with filtering
- `Cleaner`: Automated cleanup with retention policies
- `Utils`: Helper functions for common operations

### 4. **Safety Features**
- Path validation and normalization
- Path traversal prevention
- Cross-platform compatibility
- Graceful error handling

## Migration from Old API

### Old way (deprecated):
```go
import dir "sfDBTools/utils/fs/dir"
import file "sfDBTools/utils/fs/file"

dirManager := dir.NewManager()
fileManager := file.NewManager()
```

### New way:
```go
import "sfDBTools/utils/fs"

manager := fs.NewManager()
// All operations through manager.Dir(), manager.File(), manager.Perm()
```

## Dependencies

- `github.com/spf13/afero`: Filesystem abstraction
- `github.com/shirou/gopsutil/v3/disk`: Disk usage information
- `sfDBTools/internal/logger`: Integrated logging

## Benefits of Refactoring

1. **Eliminated Duplications**: 
   - Removed 4+ different `CopyFile` implementations
   - Centralized path validation
   - Unified error handling

2. **Improved Modularity**:
   - Single responsibility for each module
   - Clear interfaces
   - Easier testing

3. **Better Maintainability**:
   - Consistent API design
   - Centralized configuration
   - Simplified imports

4. **Enhanced Safety**:
   - Path validation
   - Permission checking
   - Cross-platform compatibility

## Migration Impact

Package migration telah diupdate untuk menggunakan API baru:
- `internal/core/mariadb/configure/migration/copy.go` 
- Semua backup operations
- Directory validation functions

Backward compatibility terjamin melalui wrapper functions di `compat.go`.