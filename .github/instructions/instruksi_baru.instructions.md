---
applyTo: '**'
---
# sfDBTools – Arsitektur & Konvensi (Ringkas)

Fokus: CLI Go untuk backup, restore, dan manajemen MariaDB/MySQL.

## Tujuan Kualitas Kode
- Bersih: fungsi pendek, nama jelas.
- Modular: paket berdasar domain, dependensi jelas.
- Scalable: tambah fitur tanpa refactor besar.
- Reusable: logic umum di `internal/core/*` (domain) & `utils/*` (stateless).
- Testable: fungsi pure / side-effect minimal.
- Konsisten: error wrapping `fmt.Errorf("context: %w", err)`, logger internal.
- Nol Duplikasi: tidak ada copy–paste logic; satu sumber kebenaran (SSOT).

## Struktur Direktori
```
cmd/{module}                 -> Definisi Cobra command (flag, binding, UX entry)
cmd/{module}/{features}      -> Fitur spesifik untuk modul / fitur
internal/core/{modul}/features  -> Logic bisnis utama (backup, restore, validation, orchestration)
internal/core/{modul}        -> Helper / utilitas khusus modul
internal/config              -> Konfigurasi (Viper)
internal/logger              -> Logging terstruktur
utils/{area} (atau common)  -> Helper stateless & generic (terminal, io, network, flag helpers)
compression/, crypto/        -> Kompresi & enkripsi
config/                      -> Contoh konfigurasi
```

Prinsip pemisahan:
- `cmd/*`: parsing flag + panggil fungsi core.
- `internal/core/*`: orkestrasi domain (boleh pakai `utils/*`).
- `utils/*`: tidak mengenal domain (NO import `core`).
- Hindari import silang antar domain; gunakan interface bila perlu.

## Anti-Duplikasi & Reusability
Wajib: tidak ada kode logika berulang atau fungsi identik di lokasi berbeda.

Aturan:
1. Sebelum tulis fungsi baru: cari dulu (`grep`, `go list`, pencarian IDE) apakah sudah ada.
2. Jika logic dipakai ≥2 command → pindah / ekstrak ke:
   - Domain-specific: `internal/core/<domain>/`
   - Generic/stateless: `utils/<kategori>/`
3. Jangan duplikasi flag parsing pattern; gunakan helper bersama jika pola terulang.
4. Hindari variabel global baru; gunakan parameter atau struct context.
5. Jika dua fungsi mirip 80% → refactor menjadi satu fungsi configurable (opsi via struct param).
6. Setiap util baru harus:
   - Nama jelas & sempit.
   - Pure (ideal), tidak log kecuali layer atas.
   - Disertai minimal 1 test jika non-trivial.
7. Proses “promote”:
   - Temukan duplikasi.
   - Ekstrak ke fungsi baru.
   - Ganti semua pemanggilan lama.
   - Jalankan `go test ./...`.
8. Review checklist (PR):
   - Ada fungsi mirip? (cek diff & existing paket)
   - Ada konstanta literal terulang? → jadikan const.
   - Ada pipeline langkah identik di beberapa command? → buat runner di core.

Naming:
- Util generic: kata kerja jelas (`BuildDumpArgs`, `FormatDuration`).
- Core orchestrator: `RunBackupPipeline`, `ValidateRestorePlan`, dsb.
- Hindari prefiks modul kecuali perlu mencegah tabrakan.

SSOT Contoh:
- Cara test koneksi DB → hanya lewat `utils/backup.TestDatabaseConnection` (jangan ulang langsung pakai driver).
- Resolusi konfigurasi backup → `backup.ResolveBackupConfigWithoutDB`.

## Pola Command
- File utama: `cmd/<name>/<name>_cmd.go`
- Registrasi via `init()` -> `rootCmd.AddCommand(...)`
- Flags di `init()`
- Minim logic (≤ ~30 LOC ideal); delegasi ke core.

## Config & Secrets
- Path: `/etc/sfDBTools/config/config.yaml`
- Akses: `config.Get()`, `GetBackupDefaults()`
- Password via env: `SFDB_PASSWORD`, `SFDB_ENCRYPTION_PASSWORD`
- Dilarang baca file langsung di command/core (kecuali lewat layer config)

## Alur Backup (Contoh Modular)
1. Command parsing
2. Resolve config → `utils/backup.ResolveBackupConfigWithoutDB`
3. Test koneksi → `backup.TestDatabaseConnection`
4. Orkestrasi pipeline → `internal/core/backup/runner.go`
5. Kompres / enkripsi → `compression/`, `crypto/`
6. UX (spinner/tabel) → `utils/terminal`

## Terminal UX
```
spinner := terminal.NewProgressSpinner("Checking connectivity...")
spinner.Start()
...
spinner.Stop()
terminal.PrintSuccess("Done")
```
Tabel: `terminal.FormatTable(headers, rows)`

## Logging
- `lg := logger.Get()`
- Hindari log ganda atas event sama (pilih layer paling tepat)
- Gunakan konteks: `lg.Info("starting backup", "db", cfg.Name)`

## Network / Preconditions
Gunakan: `RequireInternetForOperation()`, `CheckMariaDBConnectivity()`

## Error Handling
- Selalu wrap: `fmt.Errorf("prepare dump: %w", err)`
- Tidak panic kecuali fatal saat startup
- Command layer: translasi error → output user

## Ekspor Simbol
- Ekspor hanya jika dipakai lintas paket
- Jaga surface area kecil

## Testing
- Pure function: wajib test
- Pipeline: gunakan interface dependency (misal executor mysqldump) → mockable
- Tambah test saat ekstrak util baru (mencegah regresi duplikasi)

## Tambah Fitur Baru (Checklist)
1. Definisikan domain → folder `internal/core/<domain>/`
2. Buat interface bila butuh substitusi eksternal
3. Command baru di `cmd/<command>/`
4. Daftarkan flags + help
5. Cari duplikasi sebelum buat fungsi
6. Tambah test minimal
7. Pastikan lint + build + `go test ./...`
8. Review: no duplicate logic

## Refactor Bertahap (Migrasi ke `internal/core`)
1. Identifikasi cluster fungsi/logic di command
2. Ekstrak ke file `internal/core/<domain>/<role>.go`
3. Ganti pemanggilan di command
4. Tambah test sederhana
5. Hapus versi lama
6. Commit terpisah (mudah direview)

## Skalabilitas
- Orkestrasi sebagai pipeline (prepare → dump → compress → encrypt → verify)
- ≤150 LOC per file ideal
- Gunakan `context.Context` untuk fungsi blok besar (future cancel)
- Hindari state global selain config & logger

## Interface Abstraksi (Contoh)
```go
type Dumper interface {
  Dump(ctx context.Context, cfg DumpConfig) (string, error)
}
```
Memudahkan test & substitusi implementasi.

## Ringkas Prinsip Inti
- Command tipis
- Core kuat
- Utils murni & reusable
- No duplicate logic
- Error dibungkus
- Logger konsisten
- Modular & testable

---
Terakhir diperbarui: 2023-10-05