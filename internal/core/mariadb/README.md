# MariaDB Core Module

This module provides core business logic for MariaDB-related operations in sfDBTools.

## Architecture

The module follows the clean architecture principles with feature-based organization:

```
/cmd/mariadb_cmd.go                    -> Root command for MariaDB module
/cmd/mariadb_cmd/mariadb_*.go          -> Subcommands/features for MariaDB
/internal/core/mariadb/feature/*.go    -> Core logic for each feature
/internal/core/mariadb/*.go            -> Shared utilities and types
```

### Current Structure

#### Root Level (Shared)
- `types.go` - Shared domain types across all MariaDB features
- `os_validator.go` - Operating system validation (shared utility)
- `os_validator_test.go` - Tests for OS validation

#### Feature: check_version
- `check_version/types.go` - Types specific to version checking
- `check_version/service.go` - Version fetching and processing service
- `check_version/runner.go` - Main orchestrator for version checking

#### Commands
- `cmd/mariadb_cmd.go` - Root MariaDB command registration
- `cmd/mariadb_cmd/mariadb_check_version.go` - Check version subcommand

### Future Features (Planned)

```
/cmd/mariadb_cmd/mariadb_install.go
/cmd/mariadb_cmd/mariadb_remove.go
/cmd/mariadb_cmd/mariadb_check_config.go
/cmd/mariadb_cmd/mariadb_tune_config.go
/cmd/mariadb_cmd/mariadb_monitor.go
/cmd/mariadb_cmd/mariadb_edit_config.go

/internal/core/mariadb/install/*.go
/internal/core/mariadb/remove/*.go
/internal/core/mariadb/config/*.go
/internal/core/mariadb/tune/*.go
/internal/core/mariadb/monitor/*.go
```

### Design Principles

1. **Feature Isolation**: Each feature has its own directory under `internal/core/mariadb/`
2. **Shared Utilities**: Common functionality lives at the module root level
3. **Command → Logic → Helper**: Clear separation of responsibilities
4. **Single Responsibility**: Each file handles a specific concern
5. **Dependency Injection**: Services are configurable via dependency injection

## Usage

### Basic Usage (check_version)

```go
// Create default configuration
config := check_version.DefaultCheckVersionConfig()

// Create and run version checker
runner := check_version.NewCheckVersionRunner(config)
err := runner.Run()
```

### Shared Utilities

```go
// Validate OS (shared across features)
err := mariadb.ValidateOperatingSystem()
```

## Features

### Version Checking (`check_version/`)
- Fetches MariaDB versions from official API
- Filters versions based on minimum version requirement (default: 10.6+)
- Shows only stable releases
- Displays latest minor version for each major version
- Formats EOL dates in human-readable format

### Shared Utilities
- **OS Validation**: Supports CentOS, Ubuntu, RHEL, Rocky Linux, AlmaLinux
- **Common Types**: Shared configuration and status types
- **Network Operations**: Uses existing `utils/common` network validation

## Testing

Run tests for this module:

```bash
# All MariaDB tests
go test ./internal/core/mariadb/... -v

# Specific feature tests  
go test ./internal/core/mariadb/check_version/... -v
```

## Integration

Each feature is integrated into the CLI following this pattern:

1. **Command Layer** (`cmd/mariadb_cmd/mariadb_*.go`): Minimal command wrapper
2. **Core Logic** (`internal/core/mariadb/feature/*.go`): Business logic and orchestration  
3. **Utilities** (`utils/*`, `internal/core/mariadb/*.go`): Reusable helpers

### Example Integration Flow

```
User runs: ./sfdbtools mariadb check_version
↓
cmd/mariadb_cmd/mariadb_check_version.go (command parsing)
↓  
internal/core/mariadb/check_version/runner.go (orchestration)
↓
internal/core/mariadb/check_version/service.go (business logic)
↓
utils/terminal, utils/common (shared utilities)
```
