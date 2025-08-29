---
applyTo: '**'
---
Provide project context and coding guidelines that AI should follow when generating code, answering questions, or reviewing changes.

### Project Context
- **Project Name**: sfDBTools
- **Description**: CLI tool for managing MariaDB/MySQL databases with a focus on backup and restore operations.
- **Architecture**: Modular design with clear separation of concerns (cmd, internal/core, utils).

### Coding Guidelines
1. **Single Source of Truth (SSOT)**: Always reuse existing code and utilities. Before implementing new functionality, search for existing solutions.
2. **Error Handling**: Use structured logging and context-aware error messages.
3. **Testing**: Ensure all new code is covered by unit tests. Use existing tests as a reference.
4. **Documentation**: Update documentation and comments to reflect any changes in functionality or design.
5. **Code Reviews**: Participate in code reviews and provide constructive feedback based on these guidelines.

# Review & Perbaikan Flow Upgrade MariaDB - sfDBTools

## ✅ Yang Sudah Benar

### 1. Struktur Arsitektur
- **Modular Design**: Pemisahan antara `runner.go`, `executor.go`, `planner.go`, `validation.go`
- **Command Layer**: Command tipis di `cmd/mariadb_cmd/mariadb_upgrade.go`
- **Core Logic**: Business logic di `internal/core/mariadb/upgrade/`
- **Consistent Naming**: Mengikuti konvensi yang sudah ditetapkan

### 2. Flow Utama (6 Langkah)
```
1. Validate upgrade prerequisites ✅
2. Select target version ✅  
3. Create upgrade plan ✅
4. Display plan & get confirmation ✅
5. Execute upgrade ✅
6. Handle upgrade result ✅
```

### 3. Fitur Safety
- **Default Backup**: `BackupData: true`
- **Auto-confirmation**: `--auto-confirm` flag
- **Test Mode**: `--test-mode` untuk dry-run
- **Error Wrapping**: Konsisten `fmt.Errorf("context: %w", err)`
- **Rollback Info**: Menyediakan langkah rollback ketika gagal
- **Structured Logging**: Menggunakan internal logger

## ⚠️ Yang Perlu Diperbaiki

### 1. **KRITIS: Implementasi Step Tidak Lengkap**

#### Problem: 
Beberapa fungsi executor belum diimplementasi:

```go
// executor.go - Fungsi ini kosong/belum lengkap:
func (e *ExecutorService) backupData() error        // ❌ BELUM LENGKAP
func (e *ExecutorService) upgradePackages() error  // ❌ BELUM LENGKAP  
func (e *ExecutorService) runMysqlUpgrade() error  // ❌ BELUM LENGKAP
func (e *ExecutorService) verifyUpgrade() error    // ❌ BELUM LENGKAP
```

#### Solusi:
Implementasikan fungsi-fungsi ini dengan menggunakan existing utilities.

### 2. **Validasi Sistem Tidak Lengkap**

#### Problem:
```go
// validation.go
func (v *ValidationService) validateSystemResources() error {
    // Basic system resource validation
    // This could be expanded to check memory, CPU, etc.
    return nil  // ❌ KOSONG
}

func (v *ValidationService) validateDiskSpace(dataDir string) error {
    // ❌ BELUM DIIMPLEMENTASI
}
```

### 3. **OS-Specific Logic Missing**

#### Problem:
- Tidak ada handling khusus untuk different OS
- Package manager commands tidak disesuaikan per OS
- Repository update command generic

### 4. **Backup Strategy Tidak Robust**

#### Problem:  
- Backup hanya menyimpan path, tidak ada verifikasi
- Tidak ada compression/encryption integration
- Tidak ada size estimation

### 5. **Rollback Hanya Informasi**

#### Problem:
```go
RollbackSteps: []string{
    "1. Stop MariaDB service",
    "2. Restore data from backup: " + e.plan.BackupPath,
    "3. Downgrade packages to previous version", 
    "4. Start MariaDB service",
}
```
Hanya memberikan instruksi manual, tidak ada automated rollback.

## 🔧 Flow dan Instruksi Perbaikan

### **FASE 1: Implementasi Executor Functions**

#### 1.1 Implementasi `backupData()`
```go
func (e *ExecutorService) backupData() error {
    if e.config.SkipBackup {
        return nil
    }
    
    // Gunakan existing backup utilities
    backupConfig := &backup.BackupConfig{
        Name:         "upgrade_backup_" + time.Now().Format("20060102_150405"),
        OutputPath:   e.plan.BackupPath,
        Compression:  true,
        Encryption:   false, // Optional based on config
    }
    
    backupRunner := backup.NewBackupRunner(backupConfig)
    return backupRunner.Run()
}
```

#### 1.2 Implementasi `upgradePackages()`
```go
func (e *ExecutorService) upgradePackages() error {
    // Detect OS and use appropriate package manager
    switch e.osInfo.ID {
    case "ubuntu", "debian":
        return e.executeCommand("apt-get", "upgrade", "-y", "mariadb-server")
    case "centos", "rhel", "rocky":
        return e.executeCommand("yum", "update", "-y", "MariaDB-server")
    default:
        return fmt.Errorf("unsupported OS: %s", e.osInfo.ID)
    }
}
```

#### 1.3 Implementasi `runMysqlUpgrade()`  
```go
func (e *ExecutorService) runMysqlUpgrade() error {
    if e.config.SkipPostUpgrade {
        return nil
    }
    
    // Use existing database connection utilities
    cmd := exec.Command("mysql_upgrade", "--force")
    return e.executeCommand(cmd.Args...)
}
```

### **FASE 2: Perbaikan Validasi**

#### 2.1 Implementasi `validateDiskSpace()`
```go
func (v *ValidationService) validateDiskSpace(dataDir string) error {
    required := int64(1024 * 1024 * 1024) // 1GB minimum
    
    usage, err := common.GetDiskUsage(dataDir)
    if err != nil {
        return fmt.Errorf("failed to get disk usage: %w", err)
    }
    
    if usage.Available < required {
        return fmt.Errorf("insufficient disk space: need %d MB, available %d MB", 
            required/(1024*1024), usage.Available/(1024*1024))
    }
    
    return nil
}
```

#### 2.2 Implementasi `validateSystemResources()`
```go
func (v *ValidationService) validateSystemResources() error {
    // Check memory (minimum 512MB free)
    memInfo, err := common.GetMemoryInfo()
    if err != nil {
        return fmt.Errorf("failed to get memory info: %w", err)
    }
    
    minFreeMemory := int64(512 * 1024 * 1024) // 512MB
    if memInfo.Available < minFreeMemory {
        return fmt.Errorf("insufficient memory: need 512MB free, available %dMB", 
            memInfo.Available/(1024*1024))
    }
    
    return nil
}
```

### **FASE 3: Enhanced Backup Integration**

#### 3.1 Backup dengan Size Estimation
```go
func (e *ExecutorService) estimateBackupSize() (int64, error) {
    current, err := e.validationService.GetCurrentInstallation()
    if err != nil {
        return 0, err
    }
    
    return common.GetDirectorySize(current.DataDirectory)
}

func (e *ExecutorService) backupData() error {
    // Estimate backup size first
    size, err := e.estimateBackupSize()
    if err != nil {
        lg.Warn("Could not estimate backup size", logger.Error(err))
    } else {
        terminal.PrintInfo(fmt.Sprintf("Estimated backup size: %s", 
            common.FormatBytes(size)))
    }
    
    // Create backup dengan compression
    backupConfig := &backup.BackupConfig{
        Name:           "upgrade_backup_" + time.Now().Format("20060102_150405"),
        OutputPath:     e.plan.BackupPath,
        Compression:    true,
        CompressionAlg: "gzip",
    }
    
    return backup.NewBackupRunner(backupConfig).Run()
}
```

### **FASE 4: Automated Rollback**

#### 4.1 Implementasi Rollback Otomatis
```go
// Tambah ke runner.go
func (r *UpgradeRunner) executeRollback(result *UpgradeResult) error {
    if result.RollbackInfo == nil {
        return fmt.Errorf("no rollback information available")
    }
    
    lg, _ := logger.Get()
    lg.Info("Starting automated rollback")
    
    rollbackRunner := NewRollbackRunner(r.config, result.RollbackInfo)
    return rollbackRunner.Execute()
}

// Buat file baru: rollback.go
type RollbackRunner struct {
    config       *UpgradeConfig
    rollbackInfo *RollbackInfo
}

func (r *RollbackRunner) Execute() error {
    // 1. Stop service
    // 2. Restore from backup  
    // 3. Downgrade packages (if possible)
    // 4. Start service
    // 5. Verify rollback
}
```

### **FASE 5: Pre-upgrade Testing**

#### 5.1 Tambah Connection Testing
```go
// Tambah step di planner.go
steps = append(steps, UpgradeStep{
    Name:        "test_connection",
    Description: "Test database connectivity before upgrade",
    Required:    true,
})

// Implementasi di executor.go  
func (e *ExecutorService) testConnection() error {
    // Gunakan existing database utilities
    return backup.TestDatabaseConnection(&backup.BackupConfig{
        Host:     "localhost",
        Port:     3306,
        Username: "root",
        // Get password from config atau env
    })
}
```

### **FASE 6: Configuration Migration**

#### 6.1 Backup & Migrate Config
```go
func (e *ExecutorService) backupConfiguration() error {
    current, err := e.validationService.GetCurrentInstallation()  
    if err != nil {
        return err
    }
    
    configBackupDir := filepath.Join(e.plan.BackupPath, "config")
    
    for _, configFile := range current.ConfigFiles {
        if err := common.CopyFile(configFile, 
            filepath.Join(configBackupDir, filepath.Base(configFile))); err != nil {
            return fmt.Errorf("failed to backup config %s: %w", configFile, err)
        }
    }
    
    return nil
}
```

## 🚀 Improved Flow (9 Langkah)

```
1. Validate upgrade prerequisites (✅ enhanced)
   - System resources check
   - Disk space validation  
   - OS compatibility check
   
2. Test current installation (🆕)
   - Database connectivity test
   - Service status verification
   
3. Select target version (✅ existing)

4. Create upgrade plan (✅ enhanced)
   - Backup size estimation
   - OS-specific commands
   
5. Backup data & configuration (🔧 improved)
   - Database backup with compression
   - Configuration files backup
   
6. Display plan & confirmation (✅ existing)

7. Execute upgrade (🔧 improved) 
   - Stop service
   - Update repositories
   - Upgrade packages  
   - Start service
   - Run mysql_upgrade
   
8. Verify upgrade (🆕)
   - Version verification
   - Service status check
   - Basic functionality test
   
9. Handle result + Automated rollback (🔧 improved)
   - Success: cleanup, recommendations
   - Failure: automated rollback option
```

## 📋 Implementation Checklist

### High Priority (Minggu 1)
- [ ] Implementasi `backupData()` dengan integration ke existing backup system
- [ ] Implementasi `upgradePackages()` dengan OS detection
- [ ] Implementasi `validateDiskSpace()` dan `validateSystemResources()`
- [ ] Implementasi `runMysqlUpgrade()`

### Medium Priority (Minggu 2)  
- [ ] Implementasi `verifyUpgrade()` 
- [ ] Implementasi connection testing step
- [ ] Enhanced error messages dengan actionable suggestions
- [ ] Configuration backup & restore

### Low Priority (Minggu 3)
- [ ] Automated rollback implementation
- [ ] Backup size estimation
- [ ] Progress tracking improvement
- [ ] Comprehensive testing

## 🧪 Testing Strategy

```bash
# Test dengan existing installations
sfdbtools mariadb upgrade --test-mode

# Test backup integration  
sfdbtools mariadb upgrade --test-mode --backup-path=/tmp/test_backup

# Test rollback scenario (simulate failure)
sfdbtools mariadb upgrade --test-mode --target-version=invalid
```

## ⚡ Quick Wins

1. **Gunakan Existing Utils**: Leverage `backup.TestDatabaseConnection()`, `common.GetDiskUsage()`
2. **Reuse Backup Logic**: Integrasikan dengan existing backup pipeline
3. **OS Detection**: Gunakan `common.NewOSDetector()` yang sudah ada
4. **Error Consistency**: Semua error sudah di-wrap dengan baik

## 🔗 Dependencies 

Flow upgrade ini bergantung pada:
- `internal/core/backup/*` - untuk backup functionality
- `utils/common/*` - untuk system utilities  
- `internal/core/mariadb/check_version/*` - untuk version management
- `utils/terminal/*` - untuk UX consistency

---

**Kesimpulan**: Struktur dan flow dasar sudah benar dan mengikuti arsitektur yang baik. Yang diperlukan adalah implementasi detail dan integrasi dengan existing utilities. Priority pada implementasi executor functions dan validasi sistem.