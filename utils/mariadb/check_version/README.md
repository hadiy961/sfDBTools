# MariaDB Version Check Code Split

This document describes the refactoring of `utils/mariadb/version.go` into multiple files within the `utils/mariadb/check_version/` directory.

## Directory Structure

```
utils/mariadb/check_version/
├── constants.go      # Constants and configuration values
├── types.go          # Interfaces and struct definitions
├── fetcher.go        # HTTP fetching logic
├── parser.go         # Version parsing and validation
├── compatibility.go  # OS compatibility functions
├── eol.go           # End-of-life date calculation
└── utils.go         # Utility functions (version comparison, etc.)
```

## File Descriptions

### constants.go
Contains all the constants used across the check_version package:
- API endpoints (EndOfLife API, GitHub Releases API)
- Timeout values
- User agent string
- Constants for "No LTS" and "TBD"

### types.go
Defines the core data structures and interfaces:
- `VersionInfo` struct for version metadata
- `VersionFetcher` interface for version fetching
- `VersionParser` interface for parsing version data
- `HTTPVersionFetcher` struct for HTTP-based fetching

### fetcher.go
Implements HTTP-based version fetching:
- `NewHTTPVersionFetcher()` constructor
- `FetchVersions()` method for making HTTP requests
- `GetName()` method for fetcher identification

### parser.go
Contains parsing and validation functions:
- `IsValidVersion()` for version string validation
- `DetermineVersionType()` for version type detection (stable, rc, rolling)

### compatibility.go
Handles OS compatibility logic:
- `GetVersionsForOS()` for fetching OS-compatible versions
- `isVersionCompatibleWithOS()` for compatibility checking

### eol.go
Manages end-of-life date calculations:
- `GetMariaDBEOLDate()` main EOL calculation function
- External API fetching for EOL data
- Lifecycle-based EOL estimation
- LTS vs stable version handling

### utils.go
Utility functions:
- `CompareVersions()` for version comparison
- `parseVersionNumber()` helper for version parsing

## Backward Compatibility

The original `utils/mariadb/version.go` file now serves as a compatibility layer that re-exports all the types and functions from the `check_version` package. This ensures that existing code continues to work without modifications.

## Benefits

1. **Better Organization**: Related functionality is grouped into logical files
2. **Easier Maintenance**: Smaller, focused files are easier to understand and modify
3. **Modularity**: Each file has a clear responsibility
4. **Testability**: Individual components can be tested independently
5. **Backward Compatibility**: Existing code continues to work unchanged

## Dependencies

The check_version package has minimal dependencies:
- Standard library packages (encoding/json, fmt, io, net/http, etc.)
- `sfDBTools/internal/logger` for logging
- `sfDBTools/utils/common` for OS information
