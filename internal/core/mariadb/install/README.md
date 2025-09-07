# MariaDB Installation Module Refactoring

## Overview
Modul instalasi MariaDB telah direfactor untuk meningkatkan maintainability dengan memisahkan tanggung jawab ke dalam file-file terpisah yang lebih kecil dan fokus.

## Struktur File Baru

### 1. `installer.go` (Main Orchestrator)
- Koordinator utama proses instalasi
- Menggunakan komponen-komponen specialized lainnya
- Berisi logika flow instalasi secara high-level

### 2. `system_checker.go` (System Validation)
**Tanggung jawab:**
- Validasi konektivitas internet
- Pemeriksaan ketersediaan repository
- Deteksi instalasi MariaDB/MySQL yang sudah ada
- Validasi service dan package yang ada

**Fungsi utama:**
- `PerformAllChecks()` - Menjalankan semua pemeriksaan sistem
- `CheckInternetConnectivity()` - Validasi koneksi internet
- `CheckRepositoryAvailability()` - Pemeriksaan akses repository
- `CheckExistingInstallation()` - Deteksi instalasi existing

### 3. `version_selector.go` (Version Selection)
**Tanggung jawab:**
- Mengambil daftar versi MariaDB yang tersedia
- Menangani pemilihan versi (interactive/auto)
- Mempersiapkan opsi versi untuk menu

**Fungsi utama:**
- `SelectVersion()` - Entry point pemilihan versi
- `fetchAvailableVersions()` - Mengambil versi yang tersedia
- `autoSelectVersion()` - Auto-selection untuk mode non-interaktif
- `interactiveVersionSelection()` - Pemilihan versi interaktif

### 4. `repository_setup.go` (Repository Management)
**Tanggung jawab:**
- Setup repository MariaDB official
- Pembersihan repository existing
- Update package cache

**Fungsi utama:**
- `SetupRepository()` - Setup repository untuk versi tertentu
- `cleanExistingRepositories()` - Pembersihan repository lama
- `setupOfficialRepository()` - Setup repository official
- `updatePackageCache()` - Update cache package manager

### 5. `package_installer.go` (Package Installation)
**Tanggung jawab:**
- Instalasi package MariaDB
- Menentukan package yang sesuai berdasarkan OS
- Menangani proses instalasi dengan feedback

**Fungsi utama:**
- `InstallPackages()` - Instalasi package MariaDB
- `getPackagesToInstall()` - Menentukan package berdasarkan OS
- `GetPackageList()` - Mendapatkan daftar package (untuk dry-run)

### 6. `service_manager.go` (Service Management)
**Tanggung jawab:**
- Konfigurasi service MariaDB
- Start dan enable service
- Verifikasi status service

**Fungsi utama:**
- `ConfigureService()` - Konfigurasi service MariaDB
- `VerifyInstallation()` - Verifikasi instalasi berhasil
- `startService()` / `enableService()` - Operasi service

### 7. `validator.go` (Validation & Utilities)
**Tanggung jawab:**
- Validasi konfigurasi
- Pembuatan hasil instalasi
- Logging operasi
- Utility functions

**Fungsi utama:**
- `CreateInstallResult()` - Membuat result standar
- `CreateErrorResult()` / `CreateSuccessResult()` - Result dengan data spesifik
- `ValidateConfig()` - Validasi konfigurasi
- `LogInstallationStart()` / `LogInstallationSuccess()` - Logging

### 8. `dry_run.go` (Dry Run Implementation)
- Implementasi dry-run yang menggunakan komponen real untuk validasi
- Simulasi operasi tanpa melakukan perubahan aktual

### 9. `types.go` (Data Structures)
- Definisi struct dan tipe data yang digunakan bersama

## Keuntungan Refactoring

### 1. **Separation of Concerns**
- Setiap file memiliki tanggung jawab yang jelas dan terbatas
- Lebih mudah untuk memahami dan memodifikasi kode

### 2. **Reusability**
- Komponen dapat digunakan kembali di tempat lain
- Dry-run menggunakan komponen yang sama dengan instalasi real

### 3. **Testability**
- Setiap komponen dapat ditest secara terpisah
- Mock dan stub lebih mudah dibuat

### 4. **Maintainability**
- Bug lebih mudah diisolasi dan diperbaiki
- Perubahan pada satu aspek tidak mempengaruhi yang lain

### 5. **Menghindari Duplikasi**
- Fungsi-fungsi umum digunakan bersama
- Tidak ada duplikasi logic antara installer dan dry-run

## Penggunaan

### Normal Installation
```go
installer, err := NewInstaller(config)
if err != nil {
    return err
}

result, err := installer.Install()
```

### Dry Run
```go
dryRunner, err := NewDryRunInstaller()
if err != nil {
    return err
}

result, err := dryRunner.DryRun()
```

## Dependencies
Setiap komponen menggunakan utility packages yang sudah ada:
- `utils/common` - OS detection, network checks
- `utils/system` - Package manager, service manager
- `utils/repository` - Repository management
- `utils/terminal` - UI components
- `internal/logger` - Logging

## Future Improvements
1. **Interface Abstractions** - Membuat interface untuk setiap komponen
2. **Dependency Injection** - Menggunakan DI container
3. **Configuration Validation** - Validasi konfigurasi yang lebih comprehensive
4. **Error Recovery** - Mekanisme rollback untuk operasi yang gagal
5. **Progress Tracking** - Progress tracking yang lebih detailed
