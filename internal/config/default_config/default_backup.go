package defaultconfig

import (
	"sfDBTools/internal/config"
	"sfDBTools/utils/common/structs"
	"strings"
)

// GetBackupDefaults returns default values for backup command flags
func GetBackupGeneralDefaults() (*structs.BackupGeneralOptions, error) {

	// 1. Definisikan semua Default Settings dalam satu Struct Literal yang bersih.
	BackupOptions := &structs.BackupGeneralOptions{
		ConnectionOptions: structs.ConnectionOptions{
			Host: "localhost",
			Port: 3306,
			User: "root",
		},
		OutputDir: "./backups",
		CompressionOptions: structs.CompressionOptions{
			Algorithm: "pgzip",
			Level:     "default",
			Required:  false,
		},
		EncryptionOptions: structs.EncryptionOptions{
			Required: false,
		},
		VerificationOptions: structs.VerificationOptions{
			VerifyDisk:        true,
			CalculateChecksum: true,
		},
		RetentionOptions: structs.RetentionOptions{
			Days:    30,
			Enabled: false,
		},
		IncludeData: true,
		SystemUser:  false,
	}

	// 2. Coba muat konfigurasi
	cfg, err := config.Get()
	if err != nil || cfg == nil {
		// Jika gagal memuat config, kembalikan hanya dengan hardcoded defaults.
		return BackupOptions, err
	}

	// 3. Override Defaults dengan nilai dari Config (Menggunakan assignment yang ringkas)

	// Koneksi (Embedded Struct)
	opts := BackupOptions.ConnectionOptions // Ambil salinan untuk akses field yang lebih mudah
	if cfg.Database.Host != "" {
		opts.Host = cfg.Database.Host
	}
	if cfg.Database.Port != 0 {
		opts.Port = cfg.Database.Port
	}
	if cfg.Database.User != "" {
		opts.User = cfg.Database.User
	}
	if cfg.Database.Password != "" {
		opts.Password = cfg.Database.Password
	}
	// Kembalikan salinan yang dimodifikasi (karena ConnectionOptions adalah embedded struct)
	BackupOptions.ConnectionOptions = opts

	// Output & Compression
	if cfg.Backup.Storage.BaseDirectory != "" {
		BackupOptions.OutputDir = cfg.Backup.Storage.BaseDirectory
	}
	if cfg.Backup.Compression.Algorithm != "" {
		BackupOptions.CompressionOptions.Algorithm = cfg.Backup.Compression.Algorithm
	}
	if cfg.Backup.Compression.Level != "" {
		BackupOptions.CompressionOptions.Level = cfg.Backup.Compression.Level
	}
	// Perbarui Required (logika OR eksplisit)
	BackupOptions.CompressionOptions.Required = cfg.Backup.Compression.Required || BackupOptions.CompressionOptions.Required

	// Security & Verification
	BackupOptions.EncryptionOptions.Required = cfg.Backup.Security.EncryptionRequired || BackupOptions.EncryptionOptions.Required
	BackupOptions.VerificationOptions.CalculateChecksum = cfg.Backup.Security.ChecksumVerification || cfg.Backup.Verification.CompareChecksums || BackupOptions.VerificationOptions.CalculateChecksum
	BackupOptions.VerificationOptions.VerifyDisk = cfg.Backup.Verification.VerifyAfterWrite || cfg.Backup.Verification.DiskSpaceCheck || BackupOptions.VerificationOptions.VerifyDisk

	// Retention
	if cfg.Backup.Retention.Days != 0 {
		BackupOptions.RetentionOptions.Days = cfg.Backup.Retention.Days
	}
	BackupOptions.RetentionOptions.Enabled = cfg.Backup.Retention.CleanupEnabled || BackupOptions.RetentionOptions.Enabled

	// Dump Options (IncludeData)
	if cfg.Mysqldump.Args != "" {
		argsLower := strings.ToLower(cfg.Mysqldump.Args)
		// Ringkas: IncludeData = TRUE kecuali jika string mengandung --no-data
		BackupOptions.IncludeData = !strings.Contains(argsLower, "--no-data")
	}

	// System user presence
	if len(cfg.SystemUsers.Users) > 0 {
		BackupOptions.SystemUser = true
	}

	return BackupOptions, nil
}

func GetBackupAllDBDefaults() (BackupOptions *structs.BackupAllDBOptions, err error) {
	// Mulai dengan mendapatkan default umum
	generalDefaults, err := GetBackupGeneralDefaults()
	if err != nil {
		return nil, err
	}

	// Inisialisasi BackupAllDBOptions dengan embedded BackupGeneralOptions
	BackupOptions = &structs.BackupAllDBOptions{
		BackupGeneralOptions: *generalDefaults, // Embed general defaults
		MasterData:           false,
		ExcludeDatabases:     []string{},
		ExcludeSystem:        true,  // Default exclude system databases
		IncludeUsers:         false, // Default tidak menyertakan user
		MysqldumpArgs:        "",
		CaptureGTID:          true, // Default capture GTID
	}

	return BackupOptions, nil
}
