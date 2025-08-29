# sfDBTools – Arsitektur & Konvensi (Disesuaikan)

Fokus: CLI Go untuk backup, restore, migration, dan manajemen MariaDB/MySQL.

## Tujuan Kualitas Kode
- **Bersih**: fungsi pendek, nama jelas, single responsibility
- **Modular**: paket berdasar domain, dependensi jelas, feature-based organization  
- **Scalable**: tambah fitur tanpa refactor besar
- **Reusable**: logic umum di `utils/*` (stateless) & `internal/core/*` (domain)
- **Testable**: fungsi pure / side-effect minimal
- **Konsisten**: error wrapping `fmt.Errorf("context: %w", err)`, logger internal
- **Zero Duplikasi**: tidak ada copy–paste logic; satu sumber kebenaran (SSOT)

## Struktur Direktori Aktual

```
cmd/                           -> Command definitions dan entry points
├── {module}_cmd.go           -> Root command untuk setiap module (mariadb_cmd.go)
├── {module}_cmd/             -> Subcommands untuk fitur spesifik
│   ├── {module}_{feature}.go -> Implementation command (mariadb_check_version.go)
│   └── ...

internal/core/                 -> Core business logic
├── {module}/                 -> Module-specific logic (mariadb/)
│   ├── types.go              -> Shared domain types
│   ├── {utility}.go          -> Shared utilities (os_validator.go)
│   ├── {feature}/            -> Feature-specific logic (check_version/)
│   │   ├── types.go          -> Feature-specific types
│   │   ├── service.go        -> Business logic service
│   │   ├── runner.go         -> Orchestrator/workflow
│   │   └── *_test.go         -> Unit tests
│   └── ...
├── restore/                  -> Restore module dengan sub-features
│   ├── single/               -> Single database restore
│   ├── all/                  -> Multi database restore  
│   ├── user/                 -> User grants restore
│   └── utils/                -> Restore utilities
├── backup/                   -> Backup module (future)
└── migrate/                  -> Migration module (future)

internal/config/              -> Configuration management
├── model/                    -> Configuration models/structs
└── *.go                      -> Config loaders dan utilities

internal/logger/              -> Structured logging dengan zap
└── logger.go                 -> Logger wrapper dan utilities

utils/                        -> Reusable stateless utilities
├── common/                   -> Generic helpers
├── terminal/                 -> Terminal UI utilities
├── compression/              -> Compression utilities
├── crypto/                   -> Encryption utilities
├── database/                 -> Database connection helpers
├── {feature}/               -> Feature-specific helpers (restore/, migrate/, backup/)
└── ...

config/                       -> Configuration files dan templates
logs/                         -> Runtime logs
```

## Prinsip Arsitektur Aktual

### Command Layer (`cmd/`)
- **Root commands**: `{module}_cmd.go` (misal `mariadb_cmd.go`)  
- **Subcommands**: `cmd/{module}_cmd/{module}_{feature}.go`
- **Registrasi**: via `init()` ke parent command
- **Tanggung jawab**: flag parsing, parameter binding, call core logic

### Core Logic (`internal/core/`)
- **Module organization**: berdasar domain (mariadb, restore, backup, migrate)
- **Feature isolation**: setiap feature punya direktori sendiri
- **Shared utilities**: di level module root untuk cross-feature reuse
- **Service pattern**: `service.go` untuk business logic, `runner.go` untuk orchestration

### Utilities (`utils/`)
- **Stateless helpers**: tidak mengenal domain spesifik
- **Cross-module reuse**: bisa dipakai semua module
- **Feature-specific utils**: `utils/{feature}/` untuk helper spesifik fitur

## Anti-Duplikasi & Reusability

### Aturan SSOT (Single Source of Truth)
1. **Sebelum buat fungsi baru**: search existing (`grep`, `go list`, IDE search)
2. **Logic dipakai ≥2 command**: ekstrak ke shared location
   - Domain-specific → `internal/core/{module}/`
   - Generic/stateless → `utils/{kategori}/`
3. **Flag parsing patterns**: gunakan helper functions (contoh: `AddCommonRestoreFlags`)
4. **Database connections**: hanya lewat `utils/database` helpers
5. **Configuration resolution**: lewat `utils/{feature}/config.go` patterns

### Contoh SSOT Implementation
- **Database config resolution**: `utils/restore/config.go:ResolveRestoreConfig()`
- **File selection**: `utils/common/SelectConfigFileInteractive()` 
- **Terminal UI**: `utils/terminal` untuk semua spinner/progress/display
- **Error wrapping**: konsisten `fmt.Errorf("operation failed: %w", err)`

## Pola Command Implementation

### Structure Pattern
```go
var FeatureCmd = &cobra.Command{
    Use:   "feature",
    Short: "Brief description",
    Long:  `Detailed description with examples`,
    Run: func(cmd *cobra.Command, args []string) {
        if err := executeFeature(cmd); err != nil {
            lg, _ := logger.Get()
            lg.Error("Operation failed", logger.Error(err))
            os.Exit(1)
        }
    },
}

func executeFeature(cmd *cobra.Command) error {
    // 1. Get logger
    // 2. Resolve configuration
    // 3. Call core logic
    // 4. Handle result
}

func init() {
    utils.AddCommonFeatureFlags(FeatureCmd)
}
```

### Flag Registration
- Flags di `init()` function atau dedicated helper
- Gunakan shared flag helpers: `AddCommonRestoreFlags`, `AddCommonMigrationFlags`
- Environment variable fallback via `GetStringFlagOrEnv`

## Config & Environment

### Configuration Hierarchy
1. **System config**: `/etc/sfdbtools/config/config.yaml`
2. **User config**: `$HOME/.config/sfdbtools/`
3. **Environment variables**: `SFDBTOOLS_*` prefix
4. **Command flags**: highest priority

### Secret Management
- **Passwords**: via environment (`SFDB_PASSWORD`, `SFDB_ENCRYPTION_PASSWORD`)
- **Encrypted configs**: `.cnf.enc` files dengan crypto utilities
- **Config validation**: `config.LoadConfig()`, `config.Get()`

## Alur Eksekusi Tipikal

### Restore Example Flow
```
Command Layer (restore_single_cmd.go)
    ↓ minimal logic, flag parsing
Core Logic (internal/core/restore/single/)
    ↓ business logic, orchestration  
Utilities (utils/restore/, utils/database/, utils/terminal/)
    ↓ reusable helpers, UI, database connections
```

### Migration Example Flow
```
Command (migrate_selection_cmd.go) 
    ↓ resolve configurations
Utils (utils/migrate/config.go)
    ↓ get source/target configs
Core Logic (internal/core/restore/, internal/core/backup/)
    ↓ orchestrate backup → drop → restore
Utils (utils/database/, utils/terminal/)
    ↓ database operations, user feedback
```

## Logging & Error Handling

### Structured Logging
```go
lg, err := logger.Get()
lg.Info("operation started", 
    logger.String("module", "restore"),
    logger.String("file", filePath))
lg.Error("operation failed", logger.Error(err))
```

### Error Wrapping Pattern
```go
// Consistent throughout codebase
if err != nil {
    return fmt.Errorf("failed to resolve config: %w", err)
}
```

## Terminal UX Patterns

### Progress & Feedback
```go
spinner := terminal.NewProgressSpinner("Processing...")
spinner.Start()
// ... operation
spinner.Stop()
terminal.PrintSuccess("✅ Operation completed")
terminal.PrintError("❌ Operation failed")
```

### Interactive Selection
```go
selectedFile, err := common.SelectConfigFileInteractive()
selectedDB, err := database.SelectDatabaseInteractive(conn)
```

## Struktur Feature Development

### Menambah Feature Baru
1. **Command**: `cmd/{module}_cmd/{module}_{feature}.go`
2. **Core logic**: `internal/core/{module}/{feature}/`
3. **Utilities**: `utils/{feature}/` jika butuh shared helpers
4. **Tests**: `*_test.go` files
5. **Flag helpers**: di utils untuk reusability

### Feature Structure Template
```
internal/core/{module}/{feature}/
├── types.go      -> Feature-specific types & interfaces
├── service.go    -> Business logic implementation  
├── runner.go     -> Main orchestrator/workflow
└── *_test.go     -> Unit tests

cmd/{module}_cmd/{module}_{feature}.go  -> Command implementation
utils/{feature}/config.go               -> Configuration helpers (if needed)
```

## Testing Strategy

### Unit Testing
- **Pure functions**: wajib test
- **Core services**: mock dependencies via interfaces
- **Integration**: database operations dengan test containers
- **Command testing**: flag parsing dan flow validation

### Test Location Pattern
```go
// Co-located with implementation
internal/core/mariadb/os_validator_test.go  -> tests os_validator.go
internal/core/{module}/{feature}/*_test.go -> tests feature logic
```

## Configuration Management

### Config Resolution Pattern
```go
// Standard pattern across all features
func ResolveFeatureConfig(cmd *cobra.Command) (*FeatureConfig, error) {
    config := &FeatureConfig{}
    
    // 1. Try config file first
    if configFile := getConfigFile(cmd); configFile != "" {
        if err := loadFromConfigFile(config, configFile); err != nil {
            return nil, err
        }
    }
    
    // 2. Override with flags/env
    if err := applyFlagsAndEnv(cmd, config); err != nil {
        return nil, err
    }
    
    // 3. Interactive prompts for missing required fields
    if err := promptMissingFields(config); err != nil {
        return nil, err
    }
    
    return config, nil
}
```

## Module Integration Patterns

### Cross-Module Usage
- **Migration uses Restore**: import `internal/core/restore`
- **Backup uses Compression**: import `utils/compression`
- **All use Database**: import `utils/database`

### Interface Abstraction
```go
// Example for testability
type DatabaseConnector interface {
    Connect(config DatabaseConfig) (*sql.DB, error)
    TestConnection(*sql.DB) error
}
```

## Build & Release

### Build Process
```bash
go build -o sfdbtools main.go
go test ./...
```

### Installation Structure
- **Binary**: `/usr/local/bin/sfdbtools` (system) atau `$HOME/.local/bin/sfdbtools` (user)
- **Config**: `/etc/sfdbtools/` (system) atau `$HOME/.config/sfdbtools/` (user)  
- **Logs**: `/var/log/sfdbtools/` (system) atau `$HOME/.local/share/sfdbtools/logs/` (user)

## Development Checklist

### Adding New Feature
1. ✅ Define feature directory structure
2. ✅ Create command with proper flag helpers  
3. ✅ Implement core logic with separation of concerns
4. ✅ Add configuration resolution
5. ✅ Add terminal UI feedback
6. ✅ Write unit tests
7. ✅ Check for code duplication
8. ✅ Update documentation

### Code Quality Gates
- **No duplication**: search before implementing
- **Error handling**: proper wrapping dan logging
- **Testing**: minimal unit tests untuk business logic
- **Conventions**: consistent naming dan structure
- **Documentation**: update README dan examples

## Migration dari Existing Patterns

### Refactor Gradual  
1. **Identifikasi**: duplicate atau tightly coupled code
2. **Extract**: ke appropriate utils atau core modules
3. **Replace**: all call sites dengan extracted version
4. **Test**: ensure no regression
5. **Clean**: remove old implementations

## Agent Instructions (AI Development)

### Quick Reference untuk AI Agent
1. **Read first**: `README.md`, `ARCHITECTURE.md`, command examples di `cmd/`
2. **Pattern matching**: ikuti existing patterns di module yang sama
3. **No duplication**: search existing implementation sebelum create new
4. **Flag helpers**: gunakan atau extend existing helpers di `utils/`
5. **Error handling**: consistent wrapping dengan context
6. **Testing**: add minimal tests untuk non-trivial logic
7. **Safety checks**: ask untuk system paths atau security-related changes

### Build Verification
```bash
go build ./...           # Compile check
go test ./...           # Unit tests  
go run main.go --help   # Smoke test
```

---

**Terakhir diperbarui**: 2024-08-29  
**Versi**: Disesuaikan dengan implementasi sfDBTools aktual

**Catatan**: Dokumen ini mencerminkan struktur dan pattern yang benar-benar diimplementasikan dalam proyek sfDBTools, bukan template generic.