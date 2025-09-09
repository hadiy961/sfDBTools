## Fitur: Konfigurasi custom MariaDB

Dokumen ini menjelaskan langkah implementasi fitur/command baru untuk meng-konfigurasi MariaDB setelah instalasi atau yang sudah berjalan.
Tujuan: sediakan command yang aman, teruji, dan sesuai pola arsitektur repo `sfDBTools`.

### Ringkasan singkat — apa yang akan dibuat

- Command CLI baru: `mariadb configure` (file: `cmd/mariadb_cmd/mariadb_configure_cmd.go`).
- Resolver/flag helper: `utils/mariadb/config.go` dengan `ResolveMariaDBConfigureConfig(cmd *cobra.Command)`.
- Logika inti: `internal/core/mariadb/configure/configure.go`.
- Reuse helper: `utils/common`, `utils/terminal`, `utils/crypto`, `utils/database`, `utils/system`, `utils/mariadb` sesuai pola repo.

### Flow fitur

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

Semua item di atas harus tercakup oleh command baru.

### Detail implementasi
Prinsip implementasi dan best practices Go

- Single Responsibility: satu fungsi melakukan satu tugas. Jika fungsi > 50 baris atau bercabang berat, pecah.
- Dependency Injection: terima dependency (package manager, process manager, service manager) sebagai interface di argumen supaya mudah di-mock di test.
- Idempotency: buat operasi yang dapat dijalankan ulang tanpa menimbulkan efek samping berbahaya — cek kondisi sebelum melakukan perubahan (contoh: cek repo sudah ada / paket sudah terinstall).
- No duplication: jika menemukan logic yang sama di beberapa tempat, buat helper di `utils/` dan gunakan kembali. Favor composition over copy-paste.
- Small exported surface: hanya export fungsi yang diperlukan; prefer unexported (lowercase) untuk helper internal.
- Linters & formatting: pakai `gofmt`/`gofumpt`, `go vet`, dan linting (`golangci-lint`) di CI.
- Logging: gunakan logger yang sudah ada (`internal/logger`) dan hindari mencetak langsung ke stdout dari dalam core logic; gunakan `utils/terminal` di `cmd/` untuk UX interaktif.

### Rancangan struktur kode
- cmd/
  - mariadb_cmd/
    - mariadb_install_cmd.go   // flags -> build config -> call core

- internal/core/mariadb/install/
  - install.go                // entry point RunMariaDBInstall(ctx, cfg, deps)
  - precheck.go               // cek existing mariadb, os, internet
  - repo_setup.go             // download + run mariadb_repo_setup
  - package_install.go        // pm.Install + pin/version logic
  - service.go                // start/enable/verify service

- utils/
  - system/                   // os_detection, package_manager, process, service_manager
  - terminal/                 // terminal helpers
  - mariadb/ (opsional)       // mariadb-specific small helpers (version parsing, repo file parsing)


Dengan panduan ini, implementasi akan terstruktur, mudah diuji, dan mudah dipelihara.