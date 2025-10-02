// File : utils/common/structs/backup_structs.go
// Description : Structs untuk opsi backup, hasil, dan metadata terkait.
// Author : Hadiyatna Muflihun

package structs

import (
	"time"
)

// CompressionOptions
type CompressionOptions struct {
	Required  bool   `flag:"compress" env:"SFDB_COMPRESS" default:"false"`
	Algorithm string `flag:"compression" env:"SFDB_COMPRESSION" default:"pgzip"`
	Level     string `flag:"compression-level" env:"SFDB_COMPRESSION_LEVEL" default:"default"`
}

// VerificationOptions
type VerificationOptions struct {
	CalculateChecksum bool `flag:"checksum" env:"SFDB_CALCULATE_CHECKSUM" default:"true"` // Menggunakan "checksum" sesuai flag
	VerifyDisk        bool `flag:"verify-disk" env:"SFDB_VERIFY_DISK" default:"true"`
}

// RetentionOptions
type RetentionOptions struct {
	Enabled bool `flag:"retention" env:"SFDB_RETENTION_ENABLED" default:"false"`
	Days    int  `flag:"retention-days" env:"SFDB_RETENTION_DAYS" default:"30"`
}

// BackupGeneralOptions
type BackupGeneralOptions struct {
	ConnectionOptions
	OutputDir string `flag:"output-dir" env:"SFDB_OUTPUT_DIR" default:"./backups"`
	CompressionOptions
	EncryptionOptions
	VerificationOptions
	RetentionOptions

	// Opsi Dump (Opsional, tergantung penggunaan)
	IncludeData bool `flag:"data" env:"SFDB_INCLUDE_DATA" default:"true"`
	SystemUser  bool `flag:"system-user" env:"SFDB_SYSTEM_USER" default:"false"`
}

// BackupAllDBOptions
type BackupAllDBOptions struct {
	BackupGeneralOptions

	// Tambahkan SingleTransaction di sini karena krusial untuk replikasi
	SingleTransaction bool     `flag:"single-transaction" env:"SFDB_SINGLE_TRANSACTION" default:"true"`
	MasterData        bool     `flag:"master-data" env:"SFDB_MASTER_DATA" default:"true"`
	ExcludeDatabases  []string `flag:"exclude-database" env:"SFDB_EXCLUDE_DATABASES" default:""` // Default kosong untuk []string
	ExcludeSystem     bool     `flag:"exclude-system" env:"SFDB_EXCLUDE_SYSTEM" default:"true"`
	IncludeUsers      bool     `flag:"include-users" env:"SFDB_INCLUDE_USERS" default:"false"`
	MysqldumpArgs     string   `flag:"mysqldump-args" env:"SFDB_MYSQLDUMP_ARGS" default:""`
	CaptureGTID       bool     `flag:"capture-gtid" env:"SFDB_CAPTURE_GTID" default:"false"`
}

// BackupOperationResult mendefinisikan hasil umum dari operasi backup,
// yang dapat digunakan sebagai dasar untuk semua jenis backup (All DB, Single DB, User Grants).
type BackupOperationResult struct {
	// Hasil Utama Operasi Dump
	Success         bool          // Status akhir operasi: true jika berhasil tanpa error fatal
	OutputFile      string        // Path lengkap ke file backup yang dihasilkan
	OutputSize      int64         // Ukuran file backup yang dihasilkan (dalam bytes)
	ChecksumValue   string        // Nilai checksum dari file output (hanya jika CalculateChecksum=true)
	CompressionUsed string        // Jenis kompresi yang digunakan (misalnya "gzip", "none")
	Encrypted       bool          // Apakah file output terenkripsi
	Duration        time.Duration // Total waktu yang dibutuhkan untuk operasi backup
	AverageSpeed    float64       // Kecepatan rata-rata (misalnya: MB/s)
	BackupMetaFile  string        // Path ke file metadata (.json) yang dihasilkan

	// Hasil Operasi Sekunder/Pasca-Backup
	DiskVerificationPassed bool           // Hasil pemeriksaan ruang disk sebelum dump
	RetentionCleanupResult *CleanupResult // Hasil operasi cleanup/retensi (jika diaktifkan)

	// Status Error
	Error          error                 // Menyimpan error jika ada masalah non-fatal atau fatal
	MysqldumpError *MysqldumpErrorOutput // Detail output error dan exit code mysqldump
}

// MysqldumpErrorOutput mencatat semua output diagnostik dan status error
// dari eksekusi mysqldump/MariaDB dump.
type MysqldumpErrorOutput struct {
	ExitCode int    // Kode keluar (exit code) dari proses mysqldump (0 = sukses)
	Stderr   string // Output dari Standard Error (stderr) mysqldump.
	Stdout   string // Output dari Standard Output (stdout) mysqldump (terpisah dari dump data utama)
	Message  string // Pesan ringkasan atau deskripsi error yang lebih mudah dibaca
}

// CleanupResult mendefinisikan hasil dari operasi retensi/cleanup file backup lama.
type CleanupResult struct {
	Success         bool  // Apakah operasi cleanup berhasil
	FilesCleaned    int   // Jumlah file yang berhasil dihapus
	TotalSpaceFreed int64 // Total ruang disk yang dikosongkan (dalam bytes)
	FilesSkipped    int   // Jumlah file yang dilewati (misalnya, karena masih dalam masa retensi)
	CleanupError    error // Menyimpan error jika operasi cleanup gagal
	RetentionPolicy int   // Jumlah hari retensi yang diterapkan
}

type BackupAllDBResult struct {
	BackupOperationResult // Menyematkan hasil umum (Success, OutputFile, Duration, Checksum, CleanupResult, Error)

	// Detail Cakupan Database
	DatabasesIncluded []string // Daftar semua database yang berhasil disertakan dalam dump
	DatabasesExcluded []string // Daftar database yang dikecualikan (misalnya, berdasarkan ExcludeDatabases di opsi)
	SystemDatabases   bool     // Mereplikasi status apakah database sistem (mysql, sys) dieksklusikan atau tidak

	// Detail Grants (jika IncludeUsers aktif)
	UserGrantsFile string // Path ke file terpisah yang berisi output SHOW GRANTS dan/atau CREATE USER (jika IncludeUsers=true)
	UserGrantsSize int64  // Ukuran file grants

	// Detail Konsistensi dan Replikasi
	MasterDataEnabled bool             // Mereplikasi status MasterData dari opsi
	ReplicationMeta   *ReplicationMeta // Metadata posisi replikasi (File Log Biner / GTID) jika MasterData=true
	DumpExecutionTime time.Duration    // Waktu yang dihabiskan khusus untuk proses mysqldump (terpisah dari keseluruhan proses I/O)

	// Detail Eksekusi Mysqldump
	MysqldumpCommand string // Perintah mysqldump lengkap yang dieksekusi
	MysqldumpArgs    string // Argumen mysqldump yang digunakan (dari opsi)
}
