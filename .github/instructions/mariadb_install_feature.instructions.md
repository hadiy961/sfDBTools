## Fitur: Install MariaDB via MariaDB repository setup script

Dokumen ini menjelaskan langkah implementasi fitur/command baru untuk meng-install MariaDB menggunakan official MariaDB repository setup script (mis. `mariadb_repo_setup`). Tujuan: sediakan command yang aman, teruji, dan sesuai pola arsitektur repo `sfDBTools`.

### Ringkasan singkat — apa yang akan dibuat

- Command CLI baru: `mariadb install` (file: `cmd/mariadb_cmd/mariadb_install_cmd.go`).
- Resolver/flag helper: `utils/mariadb/config.go` dengan `ResolveMariaDBInstallConfig(cmd *cobra.Command)`.
- Logika inti: `internal/core/mariadb/install/install.go` (fungsi: `RunMariaDBInstall(ctx, cfg)`).
- Reuse helper: `utils/common`, `utils/terminal`, `utils/crypto` sesuai pola repo.

### Checklist requirement (diekstrak dari permintaan)

1. Pre-Installation Checks
2. Repository Setup via MariaDB repo script
3. Repository Verification
4. Installation of specific MariaDB version
5. Post-install configuration (start service, secure installation)
6. Verification of installed version

Semua item di atas harus tercakup oleh command baru.

### Kontrak kecil (inputs/outputs/error modes)

- Input: flags atau environment variables (contoh: `--version`, `--non-interactive`, `--repo-arch`, `--skip-repo-setup`).
- Output: exit code, logs via `utils/terminal`, file repo config backup jika memodifikasi repo.
- Error modes: unsupported OS/distribution, no internet, script download failure, package manager error, insufficient privileges.

### Desain & file yang dibuat / diubah

- `cmd/mariadb_cmd/mariadb_install_cmd.go`
  - Tambah `cobra.Command` baru `install`.
  - Tambah common flags (`utils/common.AddCommon*Flags`) dan mariadb-specific flags.

- `utils/mariadb/config.go`
  - Implementasi `ResolveMariaDBInstallConfig(cmd *cobra.Command)` yang memuat flags/env dan interaktif fallback.

- `internal/core/mariadb/install/install.go`
  - Fungsi `RunMariaDBInstall(ctx context.Context, cfg MariaDBInstallConfig) error` yang menjalankan langkah-langkah berikut:
    1. Pre-checks (cek instalasi existing dan versi)
    2. Download `mariadb_repo_setup` script ke temp
    3. Jalankan script dengan parameter versi/distribusi (non-interactive bila diminta)
    4. Update package manager cache
    5. Install paket server & client versi spesifik
    6. (Opsional) lock versi / pin paket
    7. Start service dan run secure installation steps
    8. Verifikasi versi terinstall

- Tambahan kecil: `utils/mariadb/script_downloader.go` (helper download + checksum verify) — optional tapi direkomendasikan.

### Langkah implementasi (developer-focused)

1. Buat command skeleton di `cmd/mariadb_cmd/mariadb_install_cmd.go` mengikuti pola di `cmd/mariadb_cmd/mariadb_*_cmd.go`.
2. Tambah flags: `--version`, `--non-interactive`, `--skip-repo-setup`, `--backup-repo`, `--force`, `--arch`.
3. Implement `ResolveMariaDBInstallConfig` di `utils/mariadb/config.go`:
   - baca flags/env (`SFDBTOOLS_MARIADB_VERSION`),
   - jika versi kosong, minta input interaktif (pakai `utils/terminal`).
4. Implement `internal/core/mariadb/install/install.go`:
   - cek apakah `mysql`/`mariadb` binary sudah ada; jika ada dan versi cocok -> exit/return dengan message.
   - download `mariadb_repo_setup` dari `https://downloads.mariadb.com/MariaDB/mariadb_repo_setup`
   - beri permission + execute script: contoh `sh mariadb_repo_setup --mariadb-server-version=<ver> --os-type <auto> --skip-maxscale`.
   - jalankan package manager update (apt/yum/dnf/zypper) sesuai distribusi (pakai existing helper di `utils/system` atau `utils/common`).
   - install paket `mariadb-server=<version>` atau `mariadb-server` kemudian cek versi yang terpasang.
   - secure installation: jalankan `mysql_secure_installation` non-interactive (atau dokumentasikan manual steps jika tidak otomatis).

### Edge cases & handling

- OS/distribution tidak didukung — detect lebih awal dan tampilkan instruksi manual.
- Script repo sudah ada (tidak overwrite kecuali `--force`); buat backup jika `--backup-repo`.
- Koneksi internet hilang saat download — retry 3x dengan exponential backoff.
- Konflik paket versi — beri opsi `--force` atau dokumen fallback manual.
- Tidak punya hak sudo — abort dengan pesan yang jelas.

### Quality gates / tests

- Unit tests: buat test untuk `utils/mariadb/config.go` resolve logic (happy path + missing flags).
- Integration smoke test (manual/CI): run command in a disposable VM/container for these cases:
  - fresh OS, install MariaDB 10.11 non-interactive
  - system with existing MariaDB same version (should skip)
- Run `go build ./...` and `go test ./...` as part of PR verification.

### Contoh penggunaan (developer / ops)

Jalankan interaktif untuk memilih versi:

```bash
# build local binary
go build -o sfdbtools main.go
./sfdbtools mariadb install
```

Non-interactive install untuk MariaDB 11.4 di Debian/Ubuntu:

```bash
sudo ./sfdbtools mariadb install --version 11.4 --non-interactive
```

Catatan: perintah `mariadb_repo_setup` membutuhkan hak root untuk menulis file repo; command harus dijalankan dengan `sudo` jika user bukan root.

### Verifikasi pasca-instalasi

- Cek service status: `systemctl status mariadb` atau `service mariadb status`.
- Cek versi: `mysql --version` atau `mariadb --version`.
- Cek repository file di `/etc/apt/sources.list.d/` (Debian/Ubuntu) atau `/etc/yum.repos.d/` (RHEL/CentOS) sesuai distribusi.

### Catatan keamanan dan UX

- Jangan menampilkan password atau menuliskannya ke log.
- Jika menjalankan `mysql_secure_installation` non-interactive, gunakan mekanisme env atau stdin yang aman.

## Integrasi dengan `utils/` yang sudah ada

Berikut peta konkret fungsi/helper yang sudah tersedia di `utils/` dan direkomendasikan untuk digunakan saat mengimplementasikan fitur ini. Sertakan import path relatif seperti pada contoh di masing-masing file implementasi.

- OS detection & validation
  - File: `utils/system/os_detection.go`
  - Fungsi: `DetectOS()` — gunakan untuk menentukan distribusi & `PackageType`.
  - Fungsi: `ValidateOperatingSystem()` — panggil awal untuk menolak OS yang tidak didukung.

- Package manager abstraction
  - File: `utils/system/package_manager.go`
  - Fungsi/factory: `NewPackageManager()` — kembalikan interface `PackageManager`.
  - Methods: `Install([]string) error`, `Remove([]string) error`, `IsInstalled(pkg string) bool`, `GetInstalledPackages() ([]string, error)`.
  - Catatan: gunakan agar tidak meng-`exec` per-distro sendiri.

- Process execution (download & menjalankan script)
  - File: `utils/system/process.go`
  - Fungsi/factory: `NewProcessManager()` → gunakan `Execute`, `ExecuteWithOutput` atau `ExecuteWithTimeout` untuk menjalankan `curl`/`sh` dan command package manager secara terkontrol.

- Service management (start/stop/cek status)
  - File: `utils/system/service_manager.go`
  - Fungsi/factory: `NewServiceManager()` → methods `Start(name)`, `Stop(name)`, `IsActive(name)`, `GetStatus(name)` untuk mengontrol `mariadb` service.

- Terminal / interaksi pengguna
  - File: `utils/terminal/terminal.go`
  - Fungsi contoh: `SafePrintln`, `WaitForEnter`, `ClearScreen`, `GetTerminalSize` — gunakan untuk output konsisten dan spinner-safe UX.

- Common flags & env helper
  - File: `utils/common/common_flag.go`
  - Fungsi: `GetStringFlagOrEnv(cmd, flag, env, default)`, `GetBoolFlagOrEnv`, `GetIntFlagOrEnv` — pakai ini saat membaca flags di `cmd/...`.
  - Fungsi lain: `FindEncryptedConfigFiles`, `SelectConfigFileInteractive`, `LoadEncryptedConfigFromFile` untuk menangani config terenkripsi bila perlu.

- Crypto / password helper
  - File: `utils/crypto/crypto.go`
  - Fungsi: `DeriveKeyFromAppConfig`, `EncryptData`, `DecryptData` — gunakan bila menyimpan/backup repo config sensitif atau menulis `.cnf.enc`.

- Database connection helper
  - File: `utils/database/connection_wrapper.go` (ekspos via `utils/database`)
  - Fungsi: `ValidateConnection`, `GetDatabaseConnection`, `GetWithoutDB` — berguna untuk verifikasi pasca-instal bila perlu menguji koneksi.

- Common config helpers
  - File: `utils/common/config_utils.go`
  - Fungsi: `FindEncryptedConfigFiles`, `SelectConfigFileInteractive`, `LoadEncryptedConfigFromFile`, `ValidateConfigFile` — mempermudah pemilihan/validasi file `.cnf.enc`.

Tips pemakaian singkat

- Untuk langkah "cek OS" panggil `ValidateOperatingSystem()` awal di `RunMariaDBInstall`.
- Untuk men-download script repo gunakan `NewProcessManager().ExecuteWithOutput("curl", []string{"-fsSL", "https://downloads.mariadb.com/MariaDB/mariadb_repo_setup"})` lalu simpan ke temp dan jalankan `sh` lewat `ExecuteWithTimeout`.
- Untuk mengupdate cache dan meng-install gunakan `pm := NewPackageManager()` lalu `pm.Install([]string{"mariadb-server"})` atau paket spesifik versi jika tersedia.
- Untuk start/cek service gunakan `sm := NewServiceManager()` dan `sm.Start("mariadb")` serta `sm.IsActive("mariadb")`.
- Baca flags dengan `utils/common.GetStringFlagOrEnv(cmd, "version", "SFDBTOOLS_MARIADB_VERSION", "")` sehingga sesuai pola repo.

## Panduan gaya kode & arsitektur (command > core > utils)

Tuliskan kode yang bersih, modular, dan mudah dibaca sesuai praktik Go profesional; pisahkan tanggung jawab menjadi tiga lapisan jelas: `command`, `core`, dan `utils`.

1. Command layer (`cmd/`)
   - Hanya berisi parsing flags, validasi input ringan, dan mapping ke struktur konfigurasi.
   - Jangan letakkan logika bisnis di sini — panggil fungsi dari `internal/core/...`.
   - Contoh: `cmd/mariadb_cmd/mariadb_install_cmd.go` membuat `MariaDBInstallConfig` dari flags lalu memanggil `internal/core/mariadb/install.RunMariaDBInstall(ctx, cfg)`.

2. Core layer (`internal/core/...`)
   - Semua logika instalasi berada di sini: pre-check, download script, repo setup, package install, service config, verifikasi.
   - Bagi file menjadi unit kecil (mis. `precheck.go`, `repo_setup.go`, `install.go`, `service.go`) — setiap file bertanggung jawab satu hal.
   - Ekspos fungsi jelas yang menerima `context.Context` dan konfigurasi minimal.

3. Utils layer (`utils/` atau `internal/pkg/...` untuk helper yang hanya dipakai internal)
   - Fungsi reusable (OS detection, package manager abstraction, process runner, service manager, terminal helpers, crypto) diletakkan di `utils/` seperti saat ini.
   - Jika helper hanya dipakai oleh core mariadb, pertimbangkan membuat `utils/mariadb` untuk menghindari kebocoran API global.

Prinsip implementasi dan best practices Go

- Single Responsibility: satu fungsi melakukan satu tugas. Jika fungsi > 50 baris atau bercabang berat, pecah.
- Dependency Injection: terima dependency (package manager, process manager, service manager) sebagai interface di argumen supaya mudah di-mock di test.
- Context & timeouts: semua operasi eksternal (download, apt/yum, service ops) harus menerima `context.Context` dan timeouts; gunakan `ExecuteWithTimeout` dari `utils/system/process.go`.
- Error handling: wrap error dengan kontekstual pesan (`fmt.Errorf("download repo script: %w", err)`) dan return error ke caller untuk logging tersentral.
- Idempotency: buat operasi yang dapat dijalankan ulang tanpa menimbulkan efek samping berbahaya — cek kondisi sebelum melakukan perubahan (contoh: cek repo sudah ada / paket sudah terinstall).
- No duplication: jika menemukan logic yang sama di beberapa tempat, buat helper di `utils/` dan gunakan kembali. Favor composition over copy-paste.
- Small exported surface: hanya export fungsi yang diperlukan; prefer unexported (lowercase) untuk helper internal.
- Tests: unit test tiap helper + integration test minimal untuk `RunMariaDBInstall` (mock package manager & process untuk CI). Sertakan happy path + 2 edge cases (no-internet, unsupported OS).
- Linters & formatting: pakai `gofmt`/`gofumpt`, `go vet`, dan linting (`golangci-lint`) di CI.
- Logging: gunakan logger yang sudah ada (`internal/logger`) dan hindari mencetak langsung ke stdout dari dalam core logic; gunakan `utils/terminal` di `cmd/` untuk UX interaktif.
- Security: jangan log secret (password, keys). Saat perlu pass credentials ke commands, gunakan stdin atau file permission-restricted temp files.

Rancangan contoh file/folder (minimal ideal)

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

Quality gates terkait gaya

- Pastikan `go build ./...` sukses.
- Linting pass: `golangci-lint run` (jika tersedia di CI).
- Unit tests: `go test ./...` (tulis test untuk resolver dan critical helpers).

Dengan panduan ini, implementasi akan terstruktur, mudah diuji, dan mudah dipelihara. Saya dapat segera membuat PR skeleton yang mengikuti struktur dan prinsip ini (command + core + utils) jika Anda ingin saya lanjutkan.
