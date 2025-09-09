# Task: Memusatkan Operasi File ke `utils/file`

Deskripsi singkat
- Memindahkan dan menyatukan semua operasi file (baca, tulis, copy, move, remove, stat, temp file, hash, permission pada file, dan atomic write) ke paket terpusat `utils/file`.
- Tujuan: menghilangkan duplikasi, mempermudah testing (mock FS), meningkatkan konsistensi error handling, dan memisahkan concern antara operasi file dan operasi direktori (`utils/dir`).

Prioritas & estimasi
- Prioritas: High
- Estimasi effort: 2–4 hari kerja untuk implementasi awal + proof-of-concept (tergantung cakupan refactor). Implementasi penuh di seluruh repo bisa lebih lama (mingguan) bergantung review.

Checklist tugas (urutan direkomendasikan)

1) Persiapan paket dasar `utils/file` (Implementasi & tests)
   - [ ] Buat direktori `utils/file`
   - [ ] Implement API minimal:
     - ReadFile, WriteFileAtomic, Open, Create, Copy, Move, Remove, RemoveAll, Exists, Stat, TempFile, HashFile, SetPerm
   - [ ] Tambahkan unit tests menggunakan `afero.NewMemMapFs()` (6-10 test: happy path + not found + permission error)
   - [ ] Jalankan `go test ./utils/file` dan pastikan semua test lulus

2) Proof-of-Concept (refactor kecil)
   - [ ] Pilih satu modul kecil di `utils/` (rekomendasi: `utils/crypto/file_utils.go`)
   - [ ] Refactor modul tersebut untuk memakai `utils/file` API
   - [ ] Jalankan unit tests modul terkait
   - [ ] Verifikasi CLI smoke test (contoh: command yang menggunakan fitur tersebut)

3) Refactor `utils/dir` untuk delegasi file ops
   - [ ] Ganti `m.fs.Create` (test file) dengan `file.TempFile` helper
   - [ ] Ganti `m.fs.Remove` / `m.fs.RemoveAll` calls pada `cleanup.go` menjadi `file.Remove` / `file.RemoveAll`
   - [ ] Update `GetSize` untuk memakai `file.Stat` helper jika diperlukan
   - [ ] Jalankan `go test` untuk paket `utils/dir`

4) Bertahap refactor `utils/*` lainnya
   - [ ] `utils/compression` — gunakan `file.Open`/`file.Create`/`file.Copy`
   - [ ] `utils/dbconfig` — gunakan `file.Copy`/`file.Move`/`file.Remove` dan delegasikan ReadDir ke `utils/dir` jika cocok
   - [ ] `utils/backup` / `utils/backup_restore` — gunakan `file` untuk file mutasi
   - [ ] `utils/restore` — gunakan `file` untuk open/stat dan biarkan scanning di `dir`
   - [ ] Untuk setiap modul: tambahkan/ubah unit tests sesuai

5) Refactor `internal/*` dan `cmd/*` (phase terakhir)
   - [ ] Migrasi internal core modules (backup/restore/mariadb) ke `utils/file`
   - [ ] Uji end-to-end: `go test ./...` dan jalankan smoke CLI

6) Cleanup & dokumentasi
   - [ ] Hapus/arsipkan fungsi file-level duplikat
   - [ ] Update `docs/` dan `dokumen/file-ops-centralization.md` dengan status migrasi
   - [ ] Tambahkan contoh penggunaan `utils/file` di README atau docs developer

Verification & Quality Gates
- Unit tests: setiap perubahan fungsional harus memiliki unit test (use MemMapFs where applicable).
- Build & vet: jalankan `go vet` dan `go build` pada module utama.
- Full test: jalankan `go test ./...` dan pastikan tidak ada regresi.
- Smoke: contoh command CLI yang relevan (backup/restore) diuji di environ dev.

Commands (copyable)
```bash
# jalankan unit tests untuk package baru
go test ./utils/file -v

# jalankan seluruh test suite
go test ./... -v

# build untuk cek error compile
go build ./...
```

Risiko & mitigasi singkat
- Risiko: change in behavior (permissions, symlink handling, atomic rename semantics)
  - Mitigasi: tambahkan tests untuk symlink, permissions, gunakan atomic write pattern (tmp+rename)
- Risiko: regressi besar saat migrasi masif
  - Mitigasi: lakukan bertahap, gunakan proof-of-concept, dan run full test suite setiap fase

Pemilik tugas
- Owner: (assign to) developer yang familiar dengan `utils/` dan core backup/restore
- Reviewer: developer lain dengan knowledge internal core

Catatan tambahan
- Jika ingin, saya bisa langsung membuat patch awal untuk `utils/file` + tests dan proof-of-concept refactor untuk `utils/crypto` — sebutkan jika Anda mau saya lanjutkan.
