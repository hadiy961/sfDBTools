# Task 000 — Migrasi operasi direktori ke `utils/dir`

Ringkasan tujuan:
- Sentralisasi pembuatan, pemeriksaan, penelusuran, penghapusan direktori dan operasi permission ke package `sfDBTools/utils/dir`.

Priority: Tinggi -> Sedang -> Rendah

Daftar task (urut prioritas):

1) (Tinggi) Ganti `os.MkdirAll`/`os.RemoveAll` pada alur backup utama
   - Files / fungsi:
     - `utils/compression/compression.go` — `CompressFile` (buat output dir)
     - `utils/backup/grant.go` — `GenerateGrantOutputPaths`, `BackupDatabaseGrants`, `BackupSystemUserGrants` (pembuatan dir sebelum tulis)
     - `internal/core/backup/...` (lokasi lain yang memanggil `os.MkdirAll`/`os.Create`) — migrasi bertahap
   - Hasil: panggil `dir.Create` atau `dir.NewManager().Create(...)` dan gunakan `dir.Manager` untuk pengecekan/validasi.
   - Estimasi: 1--2 PR kecil.

2) (Tinggi) Migrasi traversal / scan direktori ke `dir.Scanner`
   - Files / fungsi:
     - `utils/restore/file_selector.go` — `FindBackupFiles`, `FindGrantsFiles`, `SelectBackupFileInteractive`
     - `cmd/system_cmd/system_storage_monitor_cmd.go` — `computeImmediateSubdirSizes`
   - Hasil: gunakan `dir.Scanner.List`, `dir.Scanner.Find` atau `dir.Scanner.Walk`.
   - Estimasi: 1--3 PR.

3) (Tinggi) Hapus/cleanup direktori menggunakan `dir.Cleanup`/`Manager.RemoveAll`
   - Files / fungsi:
     - `internal/core/mariadb/remove/data_cleanup.go` — `removeDefaultDataDirectories`, `removeCustomDataDirectories`, `removeConfigFiles`, `removeUserConfigFiles`
     - `utils/backup/cleanup.go` — `CleanupOldBackups`
   - Hasil: gunakan `dir.Cleanup` (sudah ada) atau `Manager.RemoveAll`.

4) (Sedang) Permission & ownership
   - Files / fungsi:
     - `utils/file/file_permission.go` — `SetDirectoryPermissions`, `SetFilePermissions`, `CreateDirectoryWithPermissions`
     - `utils/dir/permission_unix.go` dan `permission_windows.go` menyediakan Manager API; adaptasi pemanggil.
   - Hasil: panggil `dir.NewManager().SetPermissions(...)` atau `CreateWithPermissions`.

5) (Sedang) FileManager dan config handling
   - Files / fungsi:
     - `utils/dbconfig/filemanager.go` — banyak operasi `os.ReadDir`, `os.Stat`, `os.Remove`, `os.MkdirAll`, `os.Rename`.
   - Hasil: gunakan `dir.Scanner.List` untuk list dan `dir.Manager` untuk create/remove.

6) (Rendah) Utility kecil dan validasi
   - Files / fungsi:
     - `internal/config/validate/util.go` — `DirExistsAndWritable` → ganti implementasi dengan `dir.Manager`.
     - `utils/common/config_utils.go` — `FindEncryptedConfigFiles`, `SelectConfigFileInteractive` → gunakan `dir.Scanner`.

Task metadata (per task):
- owner: (tbd)
- branch: task/000-migrate-dir-ops/<subtask>
- estimate: small/medium
- notes: selalu jalankan `go build` dan `go test ./...` sebelum PR

Langkah implementasi per subtask:
1. Baca file target dan identifikasi baris yang memanggil API filesystem.
2. Tambah import `"sfDBTools/utils/dir"`.
3. Ganti panggilan `os.MkdirAll`/`os.RemoveAll`/`os.ReadDir`/`filepath.Walk` dengan wrapper `dir`.
4. Jalankan `go test ./...` dan build.
5. Buat PR kecil dan deskripsikan perubahan.

Jika setuju, saya bisa mulai dengan subtask (1) migrasi `CompressFile` dan pembuatan direktori grants sebagai contoh dan membuat patch PR.

---

Detil subtasks per-file (dari `dokumen/directory-ops-audit.md`)

Untuk tiap subtask: buat branch `task/000-<nomor>-<file-short>` dan PR kecil.

Tinggi (immediate)

000-1: `utils/compression/compression.go` — migrasi pembuatan direktori di `CompressFile`
- Branch: `task/000-1-compression-create-outputdir`
- Estimate: small (30-60m)
- Tindakan: ganti `os.MkdirAll(filepath.Dir(outputPath), ...)` dengan `dir.Create(filepath.Dir(outputPath))`.
- Verifikasi: build + unit test (jika ada), jalankan contoh compress lokal.

000-2: `utils/backup/grant.go` — pastikan penggunaan `dir.Create` pada `GenerateGrantOutputPaths` / sebelum menulis file
- Branch: `task/000-2-backup-grant-create-dir`
- Estimate: small (30-60m)
- Tindakan: impor `sfDBTools/utils/dir`; saat membuat `grantsDir`, panggil `dir.Create(grantsDir)` atau `dir.NewManager().Create(grantsDir)` sebelum `os.Create`.
- Verifikasi: build + jalankan fungsi writing grant flow (unit/integration minimal).

000-3: `utils/restore/file_selector.go` — migrasi scanning/Walk ke `dir.Scanner`
- Branch: `task/000-3-restore-use-scanner`
- Estimate: medium (1-3h)
- Tindakan: gunakan `dir.NewScanner().Find` / `.List` untuk mencari file; ganti `filepath.Walk` dengan `dir.Scanner.Walk`.
- Verifikasi: test interactive selection flows.

000-4: `cmd/system_cmd/system_storage_monitor_cmd.go` — gunakan `dir.Scanner` untuk list dan `dir.Manager.GetSize` untuk menghitung
- Branch: `task/000-4-system-monitor-dir`
- Estimate: medium (1-2h)
- Tindakan: ganti `os.ReadDir` + `filepath.WalkDir` logic dengan `dir.Scanner.List` dan `dir.Manager.GetSize` per subdir.

000-5: `internal/core/mariadb/remove/data_cleanup.go` — centralize delete/backup ops
- Branch: `task/000-5-mariadb-remove-dir`
- Estimate: medium (2-4h)
- Tindakan: ganti `os.MkdirAll`, `os.RemoveAll`, `os.Stat`, `os.Remove`, `filepath.Glob` with `dir.Manager`/`dir.Scanner` equivalents and wrappers for copy (keep rsync fallback).

000-6: `utils/backup/cleanup.go` — gunakan `dir.Manager`/`dir.Cleanup`
- Branch: `task/000-6-backup-cleanup`
- Estimate: small (1h)
- Tindakan: gunakan `dir.NewCleanup()` APIs (exists) or directly call `dir.NewManager().RemoveAll`.

Sedang (next)

000-7: `utils/file/file_permission.go` — gunakan `dir.Manager.SetPermissions` dan `CreateWithPermissions`
- Branch: `task/000-7-file-permissions`
- Estimate: small (1-2h)

000-8: `utils/dbconfig/filemanager.go` — migrasi listing dan create/remove ke `dir.Scanner`/`dir.Manager`
- Branch: `task/000-8-dbconfig-filemanager`
- Estimate: medium (2-3h)

000-9: `utils/common/config_utils.go` — gunakan `dir.Scanner.List` untuk `FindEncryptedConfigFiles`
- Branch: `task/000-9-common-config-utils`
- Estimate: small (1h)

Rendah (low effort)

000-10: `internal/config/validate/util.go` — ganti `DirExistsAndWritable` implementasi ke `dir.Manager`
- Branch: `task/000-10-validate-dir`
- Estimate: small (30-60m)

000-11: `internal/config/encrypted.go` — gunakan `dir.Manager` untuk `os.Stat` checks and `os.ReadFile` remains for file read
- Branch: `task/000-11-config-encrypted`
- Estimate: small (1h)

---

Instruksi PR:
- Minimal perubahan per PR: 1 file (maks 2) untuk memudahkan review.
- Selalu tambahkan test atau langkah manual untuk verifikasi (build + quick smoke run).

Jika Anda setuju, sebut nomor subtask mana yang mau saya mulai patch sekarang (mis. `000-1` dan `000-2`), dan saya akan membuat patch serta menjalankan build/tests lokal.
