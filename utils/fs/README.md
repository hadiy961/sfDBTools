# Package fs - Filesystem Utilities

Package `fs` menyediakan abstraksi filesystem yang aman, dapat diuji, dan modular untuk sfDBTools.

## Overview

Package ini telah direfactor untuk menjadi lebih modular dan mengikuti best practices Go:

- **Centralized API**: Semua operasi filesystem melalui satu interface terpusat
- **Separation of Concerns**: Operasi file, direktori, dan permission dipisah
- **Testability**: Menggunakan `afero` untuk abstraksi filesystem
- **Backward Compatibility**: Mempertahankan fungsi-fungsi lama untuk compatibility
- **Cross-Platform**: Mendukung Unix dan Windows
- **Advanced Operations**: Checksum, verification, pattern matching, dan directory validation

## Architecture

```
fs/
├── interfaces.go      # Interface definitions
├── manager.go         # Main manager dengan unified API
├── file_ops.go        # File operations implementation
├── dir_ops.go         # Directory operations implementation  
├── perm_ops.go        # Permission management implementation
├── checksum.go        # File checksum operations
├── verification.go    # File verification utilities
├── dir_validation.go  # Directory validation operations
├── patterns.go        # Pattern matching utilities
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

### Advanced Operations

#### Checksum Operations
```go
// Calculate checksums
md5Sum, err := manager.Checksum().CalculateMD5("file.txt")
sha256Sum, err := manager.Checksum().CalculateSHA256("file.txt")

// Compare files by checksum
equal, err := manager.Checksum().CompareFiles("file1.txt", "file2.txt")

// Verify file against expected checksum
valid, err := manager.Checksum().VerifyChecksum("file.txt", "expected_hash", "md5")
```

#### File Verification
```go
// Check file existence
exists := manager.Verify().FileExists("/path/to/file")

// Compare file sizes
equal, err := manager.Verify().CompareSizes("file1.txt", "file2.txt")

// Comprehensive file verification
result, err := manager.Verify().VerifyFileIntegrity("source.txt", "dest.txt", 100*1024*1024)

// Batch verification
results := manager.Verify().VerifyFiles("/source/dir", "/dest/dir", []string{"file1", "file2"})
```

#### Directory Validation
```go
// Validate directory structure
err := manager.DirValid().ValidateDirectoryStructure("/source", "/dest")

// Verify essential directories exist
err := manager.DirValid().VerifyEssentialDirectories("/data", []string{"mysql", "logs"})

// Check if directory is empty
empty, err := manager.DirValid().IsDirectoryEmpty("/path/to/dir")

// Ensure directory structure exists
err := manager.DirValid().EnsureDirectoryStructure("/base", []string{"dir1", "dir2"})
```

#### Pattern Matching
```go
// File type detection
isLog := manager.Pattern().IsLogFile("/var/log/mysql.log")
isConfig := manager.Pattern().IsConfigFile("/etc/my.cnf")
isDB := manager.Pattern().IsDatabaseFile("/data/table.ibd")

// Directory type detection
isData := manager.Pattern().IsDataDirectory("/var/lib/mysql/db1", "/var/lib/mysql")
isSystem := manager.Pattern().IsSystemDirectory("/usr/lib")

// Custom pattern matching
matches := manager.Pattern().MatchesExtension("file.log", []string{".log", ".err"})
filtered := manager.Pattern().FilterFilesByPattern(files, "*.log")

// Group files by type
groups := manager.Pattern().GroupFilesByType([]string{"file.log", "config.cnf", "data.ibd"})
// Returns: map[string][]string{"log": ["file.log"], "config": ["config.cnf"], "database": ["data.ibd"]}
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