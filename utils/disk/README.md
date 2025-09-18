# Disk Utilities Package

This package provides disk space checking, monitoring, and usage statistics functionality. The package has been refactored to be more modular, reusable, and maintainable.

## Architecture

The package is now split into three focused modules:

### 1. Core Disk Operations (`disk.go`)
- **Purpose**: Basic disk space validation and checking
- **Functions**:
  - `CheckDiskSpace(path, minFreeSpace)` - Check if minimum free space (MB) is available
  - `CheckDiskSpaceBytes(path, minFreeBytes)` - Check if minimum free space (bytes) is available
  - `mbToBytes(mb)` - Utility function for MB to bytes conversion

### 2. Disk Monitoring (`monitor.go`)
- **Purpose**: Background monitoring and alerting
- **Types**:
  - `Monitor` - A configurable disk monitor
  - `MonitorConfig` - Configuration for monitoring
- **Functions**:
  - `NewMonitor(config)` - Create a new monitor instance
  - `MonitorDisk(path, interval, threshold, callback)` - Simple monitoring function (backward compatible)

### 3. Usage Statistics (`usage.go`)
- **Purpose**: Detailed disk usage information and reporting
- **Types**:
  - `UsageStatistics` - Comprehensive usage information
- **Functions**:
  - `GetUsageStatistics(path)` - Get detailed usage statistics
  - `GetFreeBytes(path)` - Get free bytes for a path
  - `GetTotalBytes(path)` - Get total bytes for a path
  - `GetUsedPercent(path)` - Get used percentage for a path
  - `GetAllPartitions()` - Get statistics for all mounted partitions
  - `FindBestStorageLocation(candidates)` - Find path with most free space

## Key Improvements

### 1. Eliminated Code Duplication
- **Before**: Custom path finding logic duplicated between `disk` and `fs` packages
- **After**: Uses `fs.Manager.Dir().GetDiskUsage()` for consistent path handling

### 2. Removed Over-Engineering
- **Before**: Complex `DiskProvider` interface for simple testing needs
- **After**: Direct use of `gopsutil` with clean, focused interfaces

### 3. Consistent Formatting
- **Before**: Mixed formatting and duplicate size formatting functions
- **After**: Uses `common.FormatSizeWithPrecision()` and `common.FormatPercent()` consistently

### 4. Improved Naming and Documentation
- **Before**: Mixed English and Indonesian names and comments
- **After**: Consistent English naming with clear documentation

### 5. Better Separation of Concerns
- **Before**: Single file with mixed responsibilities
- **After**: Three focused modules with clear purposes

## Usage Examples

### Basic Disk Space Checking
```go
// Check if at least 1GB is available
err := disk.CheckDiskSpace("/backup/path", 1024) // 1024 MB
if err != nil {
    log.Fatal("Insufficient disk space:", err)
}

// Check specific byte threshold
err = disk.CheckDiskSpaceBytes("/backup/path", 1024*1024*1024) // 1GB
```

### Disk Monitoring
```go
// Simple monitoring with callback
stopMonitor := disk.MonitorDisk("/backup/path", 30*time.Second, 90.0, func(usage *fs.DiskUsage) {
    log.Printf("ALERT: Disk usage at %.1f%%", usage.UsedPercent)
})
defer stopMonitor()

// Advanced monitoring with configuration
config := disk.MonitorConfig{
    Path:             "/backup/path",
    Interval:         time.Minute,
    ThresholdPercent: 85.0,
    AlertCallback: func(usage *fs.DiskUsage) {
        // Send alert notification
    },
}
monitor := disk.NewMonitor(config)
monitor.Start()
defer monitor.Stop()
```

### Usage Statistics and Reporting
```go
// Get detailed usage statistics
stats, err := disk.GetUsageStatistics("/backup/path")
if err != nil {
    log.Fatal(err)
}
fmt.Println(stats.FormatUsageReport())

// Find best storage location
candidates := []string{"/backup1", "/backup2", "/backup3"}
best, err := disk.FindBestStorageLocation(candidates)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Best location: %s with %s free\n", best.Path, 
    common.FormatSizeWithPrecision(best.Free, 2))

// Get all partition information
partitions, err := disk.GetAllPartitions()
if err != nil {
    log.Fatal(err)
}
for _, p := range partitions {
    fmt.Printf("%s: %s free\n", p.Mountpoint, 
        common.FormatSizeWithPrecision(p.Free, 2))
}
```

## Dependencies

- `sfDBTools/utils/fs` - For consistent path handling and basic disk usage
- `sfDBTools/utils/common` - For consistent formatting functions
- `sfDBTools/internal/logger` - For structured logging
- `github.com/shirou/gopsutil/v3/disk` - For cross-platform disk information

## Migration Guide

### From Old API to New API

| Old Function | New Function | Notes |
|-------------|-------------|-------|
| `GetUsage(path)` | `GetUsageStatistics(path)` | Returns more detailed information |
| `GetFreeBytes(path)` | `GetFreeBytes(path)` | Same function, improved implementation |
| `GetTotalBytes(path)` | `GetTotalBytes(path)` | Same function, improved implementation |
| `GetUsedPercent(path)` | `GetUsedPercent(path)` | Same function, improved implementation |
| `FindBestOutputMount(candidates)` | `FindBestStorageLocation(candidates)` | Better naming and error handling |
| `GetAllUsages()` | `GetAllPartitions()` | Better naming and consistent return type |
| `MonitorDisk(...)` | `MonitorDisk(...)` | Same API, improved implementation |

The refactored package maintains backward compatibility for most functions while providing cleaner, more maintainable code.