# Analisis Pemusatan Operasi File

Dokumen ini menganalisa semua fungsi yang melakukan operasi file di repository `sfDBTools`, menandai fungsi yang harus dimigrasi ke paket terpusat `utils/file`, serta memeriksa paket `utils/dir` untuk memastikan operasi direktori tetap di sana dan operasi file diekstrak.

## Ringkasan singkat & tujuan
- Tujuan: memusatkan semua operasi file (baca, tulis, copy, remove, rename, temp file, hash, stat, chmod terhadap file) ke paket `utils/file`.
- Operasi direktori (membuat/validasi/scan/walk/permission direktori) tetap berada di `utils/dir`.
- Dokumen ini berisi: temuan (file & fungsi yang memakai file I/O), fungsi di `utils/dir` yang perlu disesuaikan, rekomendasi API untuk `utils/file`, rencana migrasi bertahap, dan checklist verifikasi.

## Checklist requirement (status)
- [x] Analisa project dan daftar semua fungsi operasi file — Done (ringkasan dan pengecekan)
- [x] Beri nama paket terpusat: `utils/file` — Done
- [x] Pastikan operasi direktori tetap di `utils/dir` dan cek apakah `utils/dir` melakukan file ops — Done (lihat bagian "File ops di utils/dir")
- [x] Buat rencana migrasi dan API minimal — Done

## Ringkasan temuan (high level)
- Pencarian global menemukan banyak penggunaan fungsi I/O: `os.Open`, `os.Create`, `os.Remove`, `os.Rename`, `os.MkdirAll`, `io.Copy`, `io.ReadAll`/`ioutil.ReadFile`, `afero.*`, `filepath.Walk`, dsb.
- Lokasi utama yang melakukan operasi file:
  - `utils/` (banyak): `utils/file`, `utils/backup`, `utils/crypto`, `utils/compression`, `utils/dbconfig`, `utils/restore`, `utils/backup_restore`, dll.
  - `internal/` (core features): backup/restore, mariadb, dbconfig, config loader.
  - `cmd/` (beberapa command untuk scanning atau drop menggunakan `os.Open`, `ReadDir`).

> Catatan: daftar lengkap ada di hasil pemindaian kode. Saya menyorot file dan fungsi yang paling relevan di bawah.

## File operasi penting (pilihan, bukan exhaustively per baris)
Berikut file yang secara eksplisit menggunakan operasi file tingkat rendah dan harus dievaluasi untuk migrasi ke `utils/file`:

- utils/file/
  - `file_operations.go` — os.Create (sudah di folder file, pastikan API sentral)
  - `file_permission.go` — os.Chmod, os.MkdirAll (chmod untuk file tetap di file package; mkdir untuk direktori harus tetap di dir package)

- utils/dbconfig/
  - `filemanager.go` — os.Open, os.Create (copy/backup file), os.Remove, os.Rename
  - `validation.go` — os.Stat

- utils/backup/
  - `utils.go`, `grant.go`, `list.go`, `cleanup.go` — open/read/hash/copy/create/remove/readdir

- utils/crypto/
  - `file_utils.go` — Open, OpenFile, io.Copy, stat, dll.

- utils/compression/
  - `compression.go` — Open input file, Create output file, io.Copy

- utils/restore/
  - `file_selector.go` — filepath.Walk, os.Stat, os.Open

- utils/backup_restore/
  - `executor.go` — os.Remove, cleanup file

- internal/*
  - Banyak file di `internal/core/backup`, `internal/core/restore`, `internal/core/mariadb/remove` yang memakai os.Create, os.Open, io.Copy, os.RemoveAll, os.Rename, os.Stat.

- cmd/*
  - `cmd/system_cmd/system_storage_monitor_cmd.go` — os.ReadDir, WalkDir untuk monitoring

Jika Anda ingin saya sertakan daftar baris lengkap (file:baris), saya bisa generate file CSV/MD terperinci.

## Pemeriksaan khusus: `utils/dir` — apakah ada operasi file?
Saya meninjau folder `utils/dir` (file: `manager.go`, `scanner.go`, `cleanup.go`, `permission_unix.go`, `permission_windows.go`, dsb.). Hasil pemeriksaan:

- `manager.go` (package dir):
  - Menggunakan `afero.Fs` untuk operasi direktori (MkdirAll, DirExists, RemoveAll, Walk, ReadDir).
  - Namun beberapa fungsi membuat file untuk validasi: `IsWritable` membuat file test dengan `m.fs.Create(testPath)` dan menghapusnya. Ini adalah operasi file (create + remove).
  - `GetSize` memanggil `afero.Walk` dan menjumlah `info.Size()` — membaca metadata file (stat).

- `scanner.go` (package dir):
  - Banyak fungsi mengembalikan file entries (List, Walk, Find) dan menggunakan `afero.ReadDir` / `afero.Walk`.
  - Ini adalah operasi direktori/pemindaian, tapi juga membaca file metadata (stat via info) untuk ukuran/modtime.

- `cleanup.go` (package dir):
  - Memanggil `c.manager.fs.Remove(entry.Path)` dan `c.manager.fs.RemoveAll(entry.Path)` untuk menghapus file serta direktori.
  - `CleanupOldFiles` dan `CleanupByPattern` secara langsung menghapus file melalui `manager.fs.Remove`.

Kesimpulan: `utils/dir` saat ini mengandung beberapa operasi file (create test file, remove file, stat size). Untuk memisahkan concerns, saya rekomendasikan:

1. Biarkan `utils/dir` bertanggung jawab atas alur (scan/list/validate/remove-direktori/permission). Ia boleh memanggil API di `utils/file` untuk melakukan operasi file aktual (create, remove, stat, hash, copy).
2. Ekstrak fungsi file-level dari `utils/dir` ke `utils/file`, mis.:
   - `m.fs.Create(testPath)` → `file.CreateTestFile(fs, dir)` atau `file.CreateTemp(fs, dir, pattern)`
   - `m.fs.Remove(path)` → `file.Remove(fs, path)`
   - `GetSize` tetap berada di `dir.Manager` karena geneal concern direktori, tetapi pemanggilan `info.Size()`/`fs.Stat` dapat menggunakan `file.Stat(fs, path)` helper.
   - `cleanup` harus memanggil `file.Remove` dan `file.RemoveAll` (atau `file.RemoveFile` / `file.RemoveRecursive`) agar logic penghapusan file terkonsisten.

Ini menjaga `utils/dir` sebagai orchestrator direktori dan memindahkan detail operasi file ke `utils/file`.

## Rekomendasi nama paket dan struktur
- Paket terpusat: `utils/file` (directory: `utils/file`).
- Tetap pertahankan `utils/dir` untuk operasi direktori.

Direktori yang disarankan:

- `utils/file/`  (operasi file saja)
- `utils/dir/`   (operasi direktori dan orchestration, sudah ada)

## API minimal yang saya sarankan untuk `utils/file`
Desain API fokus pada operasi file yang sering dipakai dalam repo.

- func Open(fs afero.Fs, path string) (afero.File, error)
- func ReadFile(fs afero.Fs, path string) ([]byte, error)
- func WriteFileAtomic(fs afero.Fs, path string, data []byte, perm os.FileMode) error
- func Create(fs afero.Fs, path string, perm os.FileMode) (afero.File, error)
- func Copy(fs afero.Fs, src, dst string, perm os.FileMode) error
- func Move(fs afero.Fs, src, dst string) error
- func Remove(fs afero.Fs, path string) error
- func RemoveAll(fs afero.Fs, path string) error
- func Exists(fs afero.Fs, path string) (bool, error)
- func Stat(fs afero.Fs, path string) (os.FileInfo, error)
- func TempFile(fs afero.Fs, dir, pattern string) (afero.File, error)
- func HashFile(fs afero.Fs, path string, hashFunc func() hash.Hash) (string, error)
- func SetPerm(fs afero.Fs, path string, mode os.FileMode) error

Catatan: semua fungsi menerima `afero.Fs` sebagai parameter (atau paket dapat expose `NewWithFs(fs)` constructors). Ini memudahkan testing dengan `afero.NewMemMapFs()`.

## Mapping perubahan konkret untuk `utils/dir`
Contoh mapping perubahan kecil ketika mengekstrak file ops:

- `Manager.IsWritable`:
  - Saat ini: membuat test file `m.fs.Create(testPath)` lalu `m.fs.Remove(testPath)`.
  - Setelah: panggil `file.TempFile(fs, normalizedPath, ".sfdbtools_write_test_*")` lalu `file.Remove(fs, tempPath)`.

- `Cleanup.CleanupOldFiles` dan `Cleanup.EmptyDirectory`:
  - Saat ini: `c.manager.fs.Remove(entry.Path)` dan `c.manager.fs.RemoveAll(entry.Path)`.
  - Setelah: `file.Remove(fs, entry.Path)` / `file.RemoveAll(fs, entry.Path)`.

- `GetSize`:
  - Tetap di `dir.Manager`, tapi gunakan `file.Stat(fs, filePath)` helper untuk membaca `Size` sehingga logic stat tersentralisasi.

Dengan pola ini, `dir` tetap bertanggung jawab untuk keputusan direktori, dan semua mutasi file delegasikan ke `utils/file`.

## Rencana migrasi (langkah-langkah praktis)
1. Buat paket `utils/file` dengan API minimal (tulis implementasi menggunakan `afero` dan implementasi default `afero.NewOsFs()`). Tambahkan tests menggunakan `afero.NewMemMapFs()`.
2. Refactor `utils/dir` untuk mengganti pemanggilan langsung `m.fs.Create` / `m.fs.Remove` / `m.fs.Stat` menjadi pemanggilan `utils/file` baru. Lakukan ini untuk: `manager.go` (IsWritable, Remove/RemoveAll?), `cleanup.go` (penghapusan file), `scanner.go` (opsional: gunakan file.Stat helper untuk ukuran).
3. Refactor modul di `utils/` (backup, crypto, compression, dbconfig, restore) untuk memakai `utils/file` API alih-alih `os.*`/`afero.*` langsung. Mulai dari modul yang paling kecil seperti `utils/crypto/file_utils.go`.
4. Jalankan `go test ./...` dan perbaiki error.
5. Setelah stabil, lanjutkan refactor ke `internal/*` dan `cmd/*`.

## Quality gates & tests
- Unit tests untuk semua fungsi `utils/file` (happy path + permission/not-found errors).
- Gunakan `afero.NewMemMapFs()` untuk unit tests.
- Integration smoke test: setelah refactor beberapa modul, jalankan CLI command yang relevan (mis. backup/restore quick run) di lingkungan dev.

## Risiko & mitigasi
- Risiko: perubahan perilaku (symlink, permission, atomic-rename semantics). Mitigasi: tambahkan tests untuk symlink & file permissions; migrasi bertahap.
- Risiko: regressi karena penggunaan `afero` vs `os` langsung. Mitigasi: pastikan implementasi `utils/file` default memakai `afero.NewOsFs()`.

## Checklist migrasi singkat
- [ ] Implement `utils/file` (API minimal) + tests
- [ ] Update `utils/dir` untuk delegasi file ops ke `utils/file`
- [ ] Migrate `utils/crypto`, `utils/compression`, `utils/dbconfig` ke `utils/file`
- [ ] Migrate `internal/*` modul secara bertahap
- [ ] Jalankan `go test ./...` dan perbaiki
- [ ] Hapus duplikasi kode file ops setelah verifikasi

## Next steps yang saya bisa lakukan sekarang (pilih salah satu)
1. Implementasi awal `utils/file` + 6-8 unit tests (recommended). — saya bisa buat patch sekarang.
2. Hasilkan CSV/MD yang berisi daftar baris lengkap semua pemanggilan file I/O untuk review — saya bisa generate.
3. Buat PR patch contoh: refactor `utils/dir/cleanup.go` untuk memakai `utils/file.Remove` sebagai proof-of-concept.

Balas pilihan Anda dan saya lanjutkan implementasi yang Anda pilih.
