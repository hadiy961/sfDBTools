## Audit: Fungsi yang menggunakan operasi direktori (tanpa package `utils/dir`)

Tujuan: memusatkan semua operasi direktori ke package `sfDBTools/utils/dir`.

Ringkasan tindakan yang saya lakukan:
- Mencari pemanggilan API filesystem umum (os.*, filepath.*, ioutil.*, afero.*) di seluruh codebase.
- Mengecualikan file-file yang sudah mengimpor `sfDBTools/utils/dir`.
- Mengumpulkan nama file dan fungsi yang jelas melakukan operasi direktori/file.

Asumsi & catatan:
- Pencarian berbasis pattern (mis. `os.MkdirAll`, `os.RemoveAll`, `os.ReadDir`, `filepath.Walk`, `os.Stat`, `os.Open`, `os.Create`, dsb.).
- Beberapa fungsi melakukan operasi file dan direktori campuran (baca tulis file + pembuatan direktori). Saya tetap mencantumkan mereka jika ada pemanggilan pembuatan/hapus/list direktori.
- Saya tidak memodifikasi source code di audit ini — hanya mengumpulkan lokasi untuk refactor.

Checklist requirement pengguna:
- [x] Tampilkan file yang melakukan operasi direktori tetapi tidak menggunakan `utils/dir`.
- [x] Tampilkan nama fungsi di tiap file yang relevan.
- [x] Simpan hasil sebagai dokumen markdown di direktori `dokumen`.

Cara saya memverifikasi:
- Menggunakan pencarian teks di repo untuk pola filesystem.
- Membaca file-file penting untuk mengekstrak nama fungsi.

=== Daftar file & fungsi ===

Catatan: file-filenya tidak mengimpor `sfDBTools/utils/dir` (dipastikan dari pencarian import).

- `utils/dbconfig/filemanager.go`
  - NewFileManager
  - ListConfigFiles
  - FindConfigFile
  - DeleteConfigFile
  - DeleteMultipleFiles
  - GetConfigFilePath
  - BackupConfigFile
  - RestoreBackup
  - CleanupBackups
  - EnsureConfigDir
  - GetConfigDir
  - isValidConfigFile
  - DisplayFileListSummary

- `utils/restore/file_selector.go`
  - FindBackupFiles
  - extractDatabaseNameFromFilename
  - SelectBackupFileInteractive
  - SelectGrantsFileInteractive
  - FindGrantsFiles
  - extractGrantsTypeFromFilename
  - ValidateBackupFile

- `utils/compression/compression.go`
  - CompressFile

- `cmd/system_cmd/system_storage_monitor_cmd.go`
  - computeImmediateSubdirSizes

- `internal/core/mariadb/remove/data_cleanup.go`
  - handleDataBackup
  - backupDefaultDataDirectory
  - backupCustomDataDirectories
  - copyFile
  - removeDataAndConfig
  - removeDefaultDataDirectories
  - removeCustomDataDirectories
  - removeConfigFiles
  - removeUserConfigFiles
  - copyDirectory

- `utils/backup/cleanup.go`
  - CleanupOldBackups

- `utils/file/file_operations.go`
  - (JSONWriter) WriteToFile
  - WriteJSON

- `utils/file/file_permission.go`
  - SetFilePermissions
  - SetDirectoryPermissions
  - CreateDirectoryWithPermissions

- `utils/common/config_utils.go`
  - FindEncryptedConfigFiles
  - SelectConfigFileInteractive
  - LoadEncryptedConfigFromFile
  - GetDatabaseConfigFromEncrypted
  - ValidateConfigFile
  - HandleDecryptionError

- `utils/backup/list.go`
  - ResolveDBListFile
  - ValidateDatabaseList
  - DisplayDatabaseListValidation
  - DisplayMultiBackupSummary
  - TestDatabaseConnection

- `internal/config/validate/util.go`
  - DirExistsAndWritable

- `internal/config/encrypted.go`
  - LoadEncryptedDatabaseConfig
  - GetDatabaseConfigWithEncryption
  - GetDatabaseConfigWithPassword
  - ValidateEncryptedDatabaseConfig
  - LoadEncryptedDatabaseConfigFromFile

- `internal/core/mariadb/remove/repo_remove.go`
  - removeMariaDBRepository
  - removeDebianRepository
  - removeRPMRepository

- `internal/logger/logger.go`
  - setupFileOutput

=== Rekomendasi prioritas refactor (awal) ===
1) Centralize creation/removal of directories: functions that call `os.MkdirAll`, `os.RemoveAll`, `os.Remove` — tinggi prioritas.
   - contoh: `EnsureConfigDir`, `CompressFile` (pembuatan output dir), `GenerateGrantOutputPaths` + `Backup*` yang memanggil `os.MkdirAll`, `removeDefaultDataDirectories`.
2) Scanning/traversal logic that uses `filepath.Walk` / `filepath.WalkDir` / `os.ReadDir` — pindahkan ke `utils/dir.Scanner` atau gunakan `dir.Scanner` yang sudah ada.
   - contoh: `FindBackupFiles`, `computeImmediateSubdirSizes`, `FindGrantsFiles`.
3) Permission & ownership helpers: gunakan `dir.Manager.SetPermissions`/`SetOwnership`.
   - contoh: `SetDirectoryPermissions`, `SetFilePermissions`.
4) File-copy helpers that create directories before writing: prefer a helper that ensures directories via `dir.Manager.Create`.
   - contoh: `copyDirectory`, `copyFile`, `BackupConfigFile`.

=== Next steps saya bisa bantu ===
- Buat PR template dengan daftar fungsi yang perlu direfactor (skeleton change) — saya bisa membuat branch dan patch awal.
- Implement small wrappers in `utils/dir` (jika ada gap) dan mulai migrasi beberapa panggilan `os.MkdirAll` / `os.RemoveAll` ke `dir.Manager`.
- Tambahkan unit tests untuk fungsi `dir.Manager` jika belum lengkap.

Jika Anda ingin, saya bisa mulai men-generate PR yang memigrasi 1-2 fungsi prioritas (mis. `CompressFile` dan `GenerateGrantOutputPaths` + pembuatan direktori) ke penggunaan `utils/dir` sebagai contoh.

---
Dokumen ini dihasilkan otomatis berdasarkan pencarian pattern filesystem pada kode sumber.
