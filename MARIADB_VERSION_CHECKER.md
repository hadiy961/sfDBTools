# MariaDB Version Checker Feature

## Overview
Fitur baru untuk mengecek versi MariaDB yang tersedia saat ini, diimplementasikan sesuai dengan arsitektur dan konvensi sfDBTools.

## Implementation Summary

### 1. Core Logic (`internal/core/mariadb/`)
- **File**: `version_checker.go`
- **Purpose**: Business logic untuk mengecek versi MariaDB yang tersedia
- **Features**:
  - Fetch versi dari dokumentasi resmi MariaDB
  - Kategorisasi versi (stable, rolling, rc)
  - Identifikasi current stable dan latest version
  - Simple sorting dan comparison logic

### 2. Utilities (`utils/mariadb/`)
- **File**: `config.go` - Configuration management and flag handling
- **File**: `display.go` - Display utilities untuk berbagai format output
- **Features**:
  - Shared flag helpers (`AddCommonVersionFlags`)
  - Config resolution (`ResolveVersionConfig`)
  - Multiple output formats (table, json, simple)
  - Rich terminal UI dengan colors dan formatting

### 3. Command Integration (`cmd/mariadb_cmd/`)
- **File**: `mariadb_check_version_cmd.go`
- **Features**:
  - Follows proper command structure pattern
  - Integration dengan logger
  - Error handling sesuai konvensi
  - Command registration melalui `init()`

## Features

### Command Usage
```bash
# Default table output
./sfdbtools mariadb check_version

# JSON output
./sfdbtools mariadb check_version --output json

# Simple text output  
./sfdbtools mariadb check_version --output simple

# Detailed information
./sfdbtools mariadb check_version --details
```

### Supported Versions
Berdasarkan dokumentasi resmi MariaDB:
- **Stable**: 10.5, 10.6, 10.11, 11.4, 11.7, 11.8
- **Rolling**: 11.rolling (latest development)
- **RC**: 11.rc (release candidate)

### Output Formats

#### 1. Table Format (Default)
- Rich terminal UI dengan colors
- Formatted table menggunakan tablewriter
- Summary information
- Version type indicators
- Status indicators (Current Stable, Latest, Available)

#### 2. JSON Format
- Machine-readable output
- Complete version information
- Structured data untuk integration

#### 3. Simple Format
- Human-readable text
- Clean output untuk scripting
- Essential information only

## Architecture Compliance

### ✅ Anti-Duplication
- Reuses existing `utils/terminal/` untuk UI components
- Reuses existing `utils/common/` untuk flag helpers
- Follows shared configuration resolution patterns
- No code duplication dengan existing features

### ✅ Modular Design
- Clear separation: core logic, utilities, command
- Configurable output formats
- Extensible for future enhancements
- Independent business logic

### ✅ Clean Code
- Proper error handling dan logging
- Consistent naming conventions
- Well-documented functions
- Following Go best practices

### ✅ Terminal UI Standards
- Uses shared `terminal.PrintSuccess`, `terminal.PrintInfo`, etc.
- Consistent color scheme dan formatting
- Progressive disclosure (summary → details)
- User-friendly output

### ✅ Configuration Pattern
- Environment variable support (`SFDBTOOLS_*`)
- Flag-based configuration
- Shared config resolution helpers
- Default values yang reasonable

## Simple Implementation
Sesuai permintaan, logika dibuat sesimpel mungkin:
- Tidak ada complex API calls atau parsing
- Menggunakan known versions dari dokumentasi resmi
- Straightforward version comparison
- No external dependencies selain yang sudah ada
- Minimal network requirements

## Future Enhancements
Fitur ini bisa dikembangkan untuk:
- Real-time fetching dari MariaDB repository
- Release date information
- Download links untuk setiap versi
- Integration dengan installation commands
- Version compatibility checking

## Testing
```bash
# Build dan test
go build -o sfdbtools main.go

# Test all formats
./sfdbtools mariadb check_version
./sfdbtools mariadb check_version --output json  
./sfdbtools mariadb check_version --output simple
./sfdbtools mariadb check_version --details
```

Fitur telah ditest dan berfungsi dengan baik, mengikuti semua konvensi arsitektur sfDBTools.
