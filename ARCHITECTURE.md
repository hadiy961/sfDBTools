# sfDBTools - Arsitektur Modular

## Struktur Direktori

Aplikasi ini mengikuti pola arsitektur modular yang scalable dan maintainable:

```
/cmd/
├── module_cmd.go                     -> Root command untuk module
└── module_cmd/
    ├── module_feature1.go            -> Subcommand untuk fitur 1
    ├── module_feature2.go            -> Subcommand untuk fitur 2
    └── ...

/internal/core/
├── module/
│   ├── types.go                      -> Shared types untuk module
│   ├── utilities.go                  -> Shared utilities
│   ├── feature1/
│   │   ├── types.go                  -> Types spesifik feature1
│   │   ├── service.go                -> Business logic service
│   │   ├── runner.go                 -> Orchestrator/workflow
│   │   └── *_test.go                 -> Unit tests
│   ├── feature2/
│   │   ├── types.go                  -> Types spesifik feature2
│   │   ├── service.go                -> Business logic service
│   │   ├── runner.go                 -> Orchestrator/workflow
│   │   └── *_test.go                 -> Unit tests
│   └── ...

/utils/
├── common/                           -> Shared utilities lintas module
├── terminal/                         -> Terminal UI utilities
└── ...
```

## Alur Eksekusi

```
Command Layer (cmd/) 
    ↓ (minimal logic, parsing args/flags)
Core Logic (internal/core/)
    ↓ (business logic, orchestration)
Utilities (utils/)
    ↓ (reusable helpers, UI, network, etc.)
```

## Contoh Implementasi: MariaDB Module

### Struktur Saat Ini

```
/cmd/mariadb_cmd.go                               -> Root MariaDB command
/cmd/mariadb_cmd/
├── mariadb_check_version.go                      -> Check version subcommand
├── mariadb_install.go                            -> Install subcommand (placeholder)
├── mariadb_remove.go                             -> Remove subcommand (placeholder)
└── mariadb_check_config.go                       -> Check config subcommand (placeholder)

/internal/core/mariadb/
├── types.go                                      -> Shared MariaDB types
├── os_validator.go                               -> OS validation utility
├── os_validator_test.go                          -> OS validation tests
└── check_version/
    ├── types.go                                  -> Check version specific types
    ├── service.go                                -> Version API service
    ├── runner.go                                 -> Check version orchestrator
    └── types_test.go                             -> Check version tests
```

### Struktur Masa Depan

```
/internal/core/mariadb/
├── install/
│   ├── types.go                                  -> Install specific types
│   ├── service.go                                -> Installation service
│   ├── runner.go                                 -> Install orchestrator
│   └── *_test.go                                 -> Install tests
├── remove/
│   ├── types.go                                  -> Remove specific types
│   ├── service.go                                -> Removal service
│   ├── runner.go                                 -> Remove orchestrator
│   └── *_test.go                                 -> Remove tests
├── config/
│   ├── types.go                                  -> Config specific types
│   ├── validator.go                              -> Config validation service
│   ├── tuner.go                                  -> Config tuning service
│   ├── runner.go                                 -> Config orchestrator
│   └── *_test.go                                 -> Config tests
├── monitor/
│   ├── types.go                                  -> Monitor specific types
│   ├── collector.go                              -> Metrics collection service
│   ├── analyzer.go                               -> Analysis service
│   ├── runner.go                                 -> Monitor orchestrator
│   └── *_test.go                                 -> Monitor tests
└── tune/
    ├── types.go                                  -> Tuning specific types
    ├── optimizer.go                              -> Performance optimization service
    ├── runner.go                                 -> Tune orchestrator
    └── *_test.go                                 -> Tune tests
```

## Prinsip Design

### 1. **Command Layer (Minimal)**
- Hanya parsing arguments/flags
- Validasi input dasar
- Delegasi ke core logic
- Error handling user-friendly
- 20-50 lines per command ideal

### 2. **Core Logic (Domain)**
- Business logic dan orchestration
- Feature-specific types dan services
- Pure functions sebisa mungkin
- Dependency injection
- Comprehensive error handling

### 3. **Utilities (Reusable)**
- Stateless functions
- Cross-module reusability
- No business logic
- Well-tested
- Generic interfaces

## Konvensi

### File Naming
- `types.go` - Domain types dan configuration
- `service.go` - Main business logic service
- `runner.go` - Workflow orchestrator/pipeline
- `*_test.go` - Unit tests
- `validator.go` - Validation logic
- `helper.go` - Utility functions

### Package Naming
- `package module` - Root module package
- `package feature` - Feature-specific package
- Hindari underscore dalam package names

### Error Handling
```go
// Selalu wrap errors dengan context
return fmt.Errorf("operation failed: %w", err)

// Gunakan structured logging
lg.Error("operation failed", 
    logger.String("feature", "install"),
    logger.Error(err))
```

### Testing
```go
// Test functions, services, dan runners
func TestServiceMethod(t *testing.T) { ... }
func TestRunnerWorkflow(t *testing.T) { ... }

// Mock dependencies untuk testing
type MockService struct { ... }
```

## Benefits

1. **Scalability** - Easy menambah features baru
2. **Maintainability** - Clear separation of concerns
3. **Testability** - Isolated components
4. **Reusability** - Shared utilities
5. **Consistency** - Standard patterns across modules

## Example Usage

### Menambah Feature Baru

1. Buat command file: `cmd/module_cmd/module_newfeature.go`
2. Buat feature directory: `internal/core/module/newfeature/`
3. Implement core logic: `types.go`, `service.go`, `runner.go`
4. Add tests: `*_test.go`
5. Register command di `cmd/module_cmd.go`

### Development Workflow

```bash
# 1. Implement core logic first
go test ./internal/core/module/feature/... -v

# 2. Add command wrapper
go build -o sfdbtools main.go

# 3. Test end-to-end
./sfdbtools module feature --help
./sfdbtools module feature [args]

# 4. Integration tests
go test ./... -v
```

Struktur ini memungkinkan pengembangan yang cepat, maintainable, dan scalable untuk semua module di sfDBTools.
