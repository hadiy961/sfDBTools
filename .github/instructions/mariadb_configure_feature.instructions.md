## Fitur: Konfigurasi custom MariaDB

Dokumen ini menjelaskan langkah implementasi fitur/command baru untuk meng-konfigurasi MariaDB setelah instalasi atau yang sudah berjalan.
Tujuan: sediakan command yang aman, teruji, dan sesuai pola arsitektur repo `sfDBTools`.

### Ringkasan singkat — apa yang akan dibuat

- Command CLI baru: `mariadb configure` (file: `cmd/mariadb_cmd/mariadb_configure_cmd.go`).
- Resolver/flag helper: `utils/mariadb/config.go` dengan `ResolveMariaDBConfigureConfig(cmd *cobra.Command)`.
- Logika inti: `internal/core/mariadb/configure/configure.go`.
- Reuse helper: `utils/common`, `utils/terminal`, `utils/crypto`, `utils/database`, `utils/system`, `utils/mariadb` sesuai pola repo.

### Pola Arsitektur yang Harus Diikuti

#### 1. Command Pattern (cmd/)
```go
// cmd/mariadb_cmd/mariadb_configure_cmd.go
var mariadbConfigureCmd = &cobra.Command{
    Use:   "configure",
    Short: "Configure MariaDB server",
    RunE:  executeMariaDBConfigure,
}

func executeMariaDBConfigure(cmd *cobra.Command, args []string) error {
    // 1. Resolve config dari flags/env/file
    config, err := mariadb_utils.ResolveMariaDBConfigureConfig(cmd)
    if err != nil {
        return err
    }

    // 2. Panggil core business logic
    return mariadb_configure.RunMariaDBConfigure(context.Background(), config)
}

func init() {
    mariadb_utils.AddMariaDBConfigureFlags(mariadbConfigureCmd)
    mariadbCmd.AddCommand(mariadbConfigureCmd)
}
```

#### 2. Config Resolution Pattern (utils/mariadb/)
Ikuti pola yang ada di `utils/mariadb/config.go`:
```go
// MariaDBConfigureConfig berisi konfigurasi untuk setup MariaDB
type MariaDBConfigureConfig struct {
    ServerID               int    
    Port                   int    
    DataDir                string 
    LogDir                 string 
    BinlogDir              string 
    InnodbEncryptTables    bool   
    EncryptionKeyFile      string 
    InnodbBufferPoolSize   string 
    InnodbBufferPoolInstances int 
    NonInteractive         bool   
}

// ResolveMariaDBConfigureConfig menggunakan pola priority: flags > env > config > defaults
func ResolveMariaDBConfigureConfig(cmd *cobra.Command) (*MariaDBConfigureConfig, error) {
    // Ikuti pola yang ada: flags -> env -> config file -> defaults
    serverID := common.GetIntFlagOrEnv(cmd, "server-id", "SFDBTOOLS_MARIADB_SERVER_ID", 1)
    port := common.GetIntFlagOrEnv(cmd, "port", "SFDBTOOLS_MARIADB_PORT", 3306)
    // dst...
    
    // Validasi input user (penting untuk konfigurasi sistem)
    if err := validateConfigureInput(cfg); err != nil {
        return nil, fmt.Errorf("validasi konfigurasi gagal: %w", err)
    }
    
    return cfg, nil
}
```

#### 3. Flag Helper Pattern (utils/mariadb/)
```go
func AddMariaDBConfigureFlags(cmd *cobra.Command) {
    cmd.Flags().Int("server-id", 0, "Server ID for replication")
    cmd.Flags().Int("port", 0, "MariaDB port")
    cmd.Flags().String("data-dir", "", "Data directory path")
    cmd.Flags().String("log-dir", "", "Log directory path") 
    cmd.Flags().String("binlog-dir", "", "Binary log directory path")
    cmd.Flags().Bool("innodb-encrypt-tables", false, "Enable table encryption")
    cmd.Flags().String("encryption-key-file", "", "Encryption key file path")
    cmd.Flags().Bool("non-interactive", false, "Non-interactive mode")
}
```

#### 4. Core Logic Pattern (internal/core/mariadb/configure/)
```
internal/core/mariadb/configure/
├── configure.go        // Entry point RunMariaDBConfigure(ctx, cfg)
├── precheck.go         // Cek privilege, installasi, service status
├── template.go         // Load template config dari /etc/sfDBTools/server.cnf
├── validation.go       // Validasi input, direktori, port, permissions
├── interactive.go      // Input interaktif untuk nilai konfigurasi
├── autotuning.go       // Auto-tune berdasarkan CPU/RAM
├── migration.go        // Data migration untuk direktori
└── service.go          // Restart service, verify connection
```

### Flow Implementasi yang Harus Diikuti
```plain
1. Installation Checks
  - Apakah user memiliki privilege sudo/root
  - Cek apakah MariaDB sudah terinstall.
  - Cek versi MariaDB yang terinstall.
  - Cek apakah service berjalan.
  - Cek koneksi ke database (pakai unix socket atau TCP).
2. Cari template konfigurasi (harus ada di `/etc/sfDBTools/server.cnf`).
3. Cari file konfigurasi saat ini (misal `/etc/my.cnf.d/50-server.cnf`, `/etc/my.cnf.d/server.cnf`, `/etc/my.cnf.d/mariadb-server.cnf`, `/etc/my.cnf` dan sejenisnya).
4. Baca template konfigurasi
  - server_id = 1
  - file_key_management_encryption_algorithm        = AES_CTR
  - file_key_management_encryption_key_file        = /var/lib/mysql/encryption/keyfile
  - innodb-encrypt-tables                           = ON
  - log_bin                                         = /var/lib/mysqlbinlogs/mysql-bin
  - datadir                                         = /var/lib/mysql
  - socket                                          = /var/lib/mysql/mysql.sock
  - port                                            = 3306
  - innodb_buffer_pool_size                         = 128M
  - innodb_data_home_dir                            = /var/lib/mysql
  - innodb_log_group_home_dir                       = /var/lib/mysql
  - log_error                                       = /var/lib/mysql/mysql_error.log
  - slow_query_log_file                             = /var/lib/mysql/mysql_slow.log
  - innodb_buffer_pool_instances                    = 8
5. Baca file konfigurasi apps (/etc/sfDBTools/config/config.yaml)
  - mariadb_installation:
      base_dir: "/var/lib/mysql"
      data_dir: "/var/lib/mysql"
      log_dir: "/var/lib/mysql"
      binlog_dir: "/var/lib/mysqlbinlogs"
      port: 3306
6. Tampilkan interaktif ke user untuk input nilai konfigurasi (pakai `utils/terminal`) default dari template :
  - server_id
  - port
  - datadir (untuk datadir, socket, innodb_data_home_dir, innodb_log_group_home_dir)
  - log_bin
  - innodb-encrypt-tables (bool ON/OFF)
  - file_key_management_encryption_key_file (Jika ON)
  - logdir (untuk log_error dan slow_query_log_file)
7. Validasi input user:
  - server_id harus integer > 0
  - port harus integer 1024-65535
  - datadir, logdir, binlog_dir harus absolute path
  - datadir, logdir, binlog_dir tidak boleh sama
  - file_key_management_encryption_key_file harus absolute path (jika innodb-encrypt-tables = ON)
8. Verifikasi lokasi direktori (datadir, logdir, binlog_dir) ada dan bisa ditulis.
9. Verifikasi port tidak bentrok (pakai `utils/system`).
10. Verifikasi lokasi encryption key file file_key_management_encryption_key_file ada dan bisa dibaca.
11. Tampilkan ringkasan konfigurasi yang akan diterapkan, minta konfirmasi user (Y/n).
12. Check Device space untuk datadir, logdir, binlog_dir (pakai `utils/system`).
13. Check permission direktori (harus bisa ditulis oleh user mysql/mariadb).
14. Check Device CPU cores dan RAM (untuk autotuning innodb_buffer_pool_size dan innodb_buffer_pool_instances).
  - innodb_buffer_pool_size = 70-80% dari total RAM
  - innodb_buffer_pool_instances = min(CPU cores, buffer_pool_size/1GB)
15. Backup mariadb config file (point 3).
16. Copy template config ke tempat sementara untuk diubah.
17. Ganti placeholder di config sementara dengan nilai dari user (point 6) dan hasil check (point 8, 9, 10).
18. Tulis config sementara ke file config asli (point 3).
19. Data Migration untuk direktori (datadir, logdir, binlog_dir) (jika diubah).
    - Stop service sebelum copy data
    - Verify data integrity setelah copy
    - Cleanup old location setelah berhasil
    - Restart service setelah cleanup
    - Handle case jika lokasi sama (skip copy)
20. Restart mariadb service (pakai `utils/system/service_manager.go`).
21. Verifikasi service berjalan.
22. Verifikasi koneksi ke database (pakai `utils/database/connection_wrapper.go`).
23. Verifikasi konfigurasi diterapkan (baca dari `SHOW VARIABLES LIKE '...'`).
24. Hapus config sementara. Log hasil ke terminal (pakai `utils/terminal`).
25. Update file konfigurasi apps (`/etc/sfDBTools/config/config.yaml`) (point 5).
```

### Pola yang Harus Digunakan dari Existing Code
1. **Flag Resolution**: Gunakan `common.GetStringFlagOrEnv()`, `common.GetIntFlagOrEnv()`, `common.GetBoolFlagOrEnv()`
2. **Config Loading**: Ikuti pola di `internal/config/loader.go` untuk encrypted config
3. **Terminal UI**: Gunakan `utils/terminal` untuk spinner, progress, tables, colored output
4. **Service Operations**: Gunakan `utils/system/service_manager.go` untuk systemctl operations
5. **File Operations**: Gunakan `utils/common/file_ops.go` untuk permissions, directory creation
6. **Database Operations**: Gunakan `utils/database/connection_wrapper.go` untuk connections
7. **Logging**: Gunakan `internal/logger` dengan structured logging


### Validasi Input Khusus untuk MariaDB Configure

```go
func validateConfigureInput(cfg *MariaDBConfigureConfig) error {
    // Server ID validation
    if cfg.ServerID <= 0 || cfg.ServerID > 4294967295 {
        return fmt.Errorf("server_id must be between 1 and 4294967295")
    }
    
    // Port validation  
    if cfg.Port < 1024 || cfg.Port > 65535 {
        return fmt.Errorf("port must be between 1024 and 65535")
    }
    
    // Directory validation
    for _, dir := range []string{cfg.DataDir, cfg.LogDir, cfg.BinlogDir} {
        if !filepath.IsAbs(dir) {
            return fmt.Errorf("directory must be absolute path: %s", dir)
        }
    }
    
    // Directories must be different
    if cfg.DataDir == cfg.LogDir || cfg.DataDir == cfg.BinlogDir || cfg.LogDir == cfg.BinlogDir {
        return fmt.Errorf("directories must be different")
    }
    
    return nil
}
```

### Structure Directory yang Harus Dibuat

```
cmd/
├── mariadb_cmd/
│   └── mariadb_configure_cmd.go    // Command definition + flag registration

internal/core/mariadb/configure/
├── configure.go                    // Entry point RunMariaDBConfigure()
├── precheck.go                     // Permission, installation, service checks
├── template.go                     // Template config loading
├── validation.go                   // Input validation
├── interactive.go                  // Interactive input gathering
├── autotuning.go                   // CPU/RAM based auto-tuning
├── migration.go                    // Directory migration logic
└── service.go                      // Service restart & verification

utils/mariadb/
├── config.go                       // Config resolution (sudah ada, tambah configure)
└── flags.go                        // Flag definitions (buat baru)
```

### Dependencies yang Harus Digunakan

- `utils/common` - Flag resolution, file operations
- `utils/terminal` - Interactive UI, progress indicators
- `utils/system` - Service management, package management, system checks
- `utils/database` - Database connections, validation
- `utils/crypto` - Encrypted config file handling
- `utils/disk` - Disk operations, space checks
- `utils/file` - File operations, permissions checks
- `internal/logger` - Structured logging
- `internal/config` - Config file loading

### Error Handling Pattern

```go
// Gunakan error wrapping yang konsisten
if err := validateConfiguration(cfg); err != nil {
    return fmt.Errorf("configuration validation failed: %w", err)
}

// Log errors dengan context
lg.Error("Failed to restart MariaDB service", 
    logger.String("service", "mariadb"),
    logger.Error(err))
```
