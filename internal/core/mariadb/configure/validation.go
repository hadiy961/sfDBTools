package configure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/disk"
	mariadb_utils "sfDBTools/utils/mariadb"
	"sfDBTools/utils/system"
)

// validateSystemRequirements melakukan validasi sistem dan input user
// Sesuai dengan Step 7-11 dalam flow implementasi
func validateSystemRequirements(ctx context.Context, config *mariadb_utils.MariaDBConfigureConfig) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting system requirements validation")

	// Step 7: Validasi input user sudah dilakukan di ResolveMariaDBConfigureConfig
	// Step 8: Verifikasi lokasi direktori (datadir, logdir, binlog_dir)
	if err := validateDirectories(config); err != nil {
		return fmt.Errorf("directory validation failed: %w", err)
	}

	// Step 9: Verifikasi port tidak bentrok
	if err := validatePort(config.Port); err != nil {
		return fmt.Errorf("port validation failed: %w", err)
	}

	// Step 10: Verifikasi encryption key file (jika diperlukan)
	if config.InnodbEncryptTables {
		if err := validateEncryptionKeyFile(config.EncryptionKeyFile); err != nil {
			return fmt.Errorf("encryption key file validation failed: %w", err)
		}
	}

	// Step 12: Check device space untuk direktori
	if err := validateDiskSpace(config); err != nil {
		return fmt.Errorf("disk space validation failed: %w", err)
	}

	// Step 13: Check permission direktori
	if err := validateDirectoryPermissions(config); err != nil {
		return fmt.Errorf("directory permissions validation failed: %w", err)
	}

	lg.Info("All system requirements validation passed")
	return nil
}

// validateDirectories memverifikasi bahwa direktori ada dan bisa ditulis
func validateDirectories(config *mariadb_utils.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()

	directories := map[string]string{
		"data-dir":   config.DataDir,
		"log-dir":    config.LogDir,
		"binlog-dir": config.BinlogDir,
	}

	for name, dir := range directories {
		lg.Debug("Validating directory", logger.String("type", name), logger.String("path", dir))

		// Cek apakah path absolute
		if !filepath.IsAbs(dir) {
			return fmt.Errorf("%s must be absolute path: %s", name, dir)
		}

		// Cek apakah direktori ada, jika tidak buat
		if err := ensureDirectoryExists(dir); err != nil {
			return fmt.Errorf("failed to ensure %s exists: %w", name, err)
		}

		// Cek apakah direktori bisa ditulis
		if err := checkDirectoryWritable(dir); err != nil {
			return fmt.Errorf("%s is not writable: %w", name, err)
		}
	}

	// Pastikan direktori tidak sama
	if config.DataDir == config.LogDir {
		return fmt.Errorf("data-dir and log-dir cannot be the same: %s", config.DataDir)
	}
	if config.DataDir == config.BinlogDir {
		return fmt.Errorf("data-dir and binlog-dir cannot be the same: %s", config.DataDir)
	}
	if config.LogDir == config.BinlogDir {
		return fmt.Errorf("log-dir and binlog-dir cannot be the same: %s", config.LogDir)
	}

	return nil
}

// validatePort memverifikasi bahwa port tidak bentrok
func validatePort(port int) error {
	lg, _ := logger.Get()
	lg.Debug("Validating port", logger.Int("port", port))

	// Cek apakah port dalam range yang valid
	if port < 1024 || port > 65535 {
		return fmt.Errorf("port must be between 1024-65535, got: %d", port)
	}

	// Cek apakah port sudah digunakan
	if isPortInUse(port) {
		return fmt.Errorf("port %d is already in use", port)
	}

	return nil
}

// validateEncryptionKeyFile memverifikasi file kunci enkripsi
func validateEncryptionKeyFile(keyFile string) error {
	lg, _ := logger.Get()
	lg.Debug("Validating encryption key file", logger.String("path", keyFile))

	if keyFile == "" {
		return fmt.Errorf("encryption key file path is required when encryption is enabled")
	}

	// Cek apakah path absolute
	if !filepath.IsAbs(keyFile) {
		return fmt.Errorf("encryption key file must be absolute path: %s", keyFile)
	}

	// Pastikan parent directory ada
	keyDir := filepath.Dir(keyFile)
	if err := ensureDirectoryExists(keyDir); err != nil {
		return fmt.Errorf("failed to ensure encryption key directory exists: %w", err)
	}

	// Jika file belum ada, buat file dummy untuk validasi
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		lg.Info("Encryption key file does not exist, will be created during configuration")

		// Test apakah bisa membuat file di lokasi tersebut
		testFile := keyFile + ".test"
		if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
			return fmt.Errorf("cannot create encryption key file at %s: %w", keyFile, err)
		}
		os.Remove(testFile) // cleanup test file
	} else if err != nil {
		return fmt.Errorf("failed to check encryption key file: %w", err)
	} else {
		// File ada, cek apakah bisa dibaca
		if _, err := os.ReadFile(keyFile); err != nil {
			return fmt.Errorf("encryption key file is not readable: %w", err)
		}
	}

	return nil
}

// validateDiskSpace memeriksa space disk untuk direktori
func validateDiskSpace(config *mariadb_utils.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()
	lg.Debug("Validating disk space")

	directories := []string{config.DataDir, config.LogDir, config.BinlogDir}

	for _, dir := range directories {
		// Gunakan utils/disk untuk cek space (jika tersedia)
		// Minimal requirement: 1GB free space
		minSpace := int64(1024 * 1024 * 1024) // 1GB in bytes

		freeSpace, err := getDiskFreeSpace(dir)
		if err != nil {
			lg.Warn("Could not check disk space", logger.String("dir", dir), logger.Error(err))
			continue
		}

		if freeSpace < minSpace {
			return fmt.Errorf("insufficient disk space in %s: required 1GB, available %d bytes", dir, freeSpace)
		}

		lg.Debug("Disk space check passed",
			logger.String("dir", dir),
			logger.Int64("free_space_mb", freeSpace/1024/1024))
	}

	return nil
}

// validateDirectoryPermissions memeriksa permission direktori untuk user mysql
func validateDirectoryPermissions(config *mariadb_utils.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()
	lg.Debug("Validating directory permissions")

	directories := []string{config.DataDir, config.LogDir, config.BinlogDir}

	for _, dir := range directories {
		// Cek apakah user mysql bisa akses direktori
		if err := checkMySQLUserAccess(dir); err != nil {
			lg.Warn("MySQL user access check failed",
				logger.String("dir", dir),
				logger.Error(err))

			// Coba fix permission
			if err := fixDirectoryPermissions(dir); err != nil {
				return fmt.Errorf("failed to fix permissions for %s: %w", dir, err)
			}
		}
	}

	return nil
}

// Helper functions

// ensureDirectoryExists memastikan direktori ada, buat jika tidak ada
func ensureDirectoryExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// checkDirectoryWritable memeriksa apakah direktori bisa ditulis
func checkDirectoryWritable(dir string) error {
	testFile := filepath.Join(dir, ".sfdbtools_write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("directory is not writable: %w", err)
	}
	os.Remove(testFile) // cleanup
	return nil
}

// isPortInUse memeriksa apakah port sedang digunakan
func isPortInUse(port int) bool {
	// Port available = false means port is in use
	return !system.IsPortAvailable(port)
}

// getDiskFreeSpace mendapatkan free space untuk direktori
func getDiskFreeSpace(dir string) (int64, error) {
	return disk.GetFreeBytes(dir)
}

// checkMySQLUserAccess memeriksa apakah user mysql bisa akses direktori
func checkMySQLUserAccess(dir string) error {
	// Cek apakah direktori owned oleh mysql user atau accessible
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}

	// Untuk implementasi sederhana, cek apakah world-writable atau group-writable
	mode := info.Mode()
	if mode&0020 != 0 || mode&0002 != 0 { // group-writable atau world-writable
		return nil
	}

	// TODO: Cek ownership yang sebenarnya dengan syscall
	return nil
}

// fixDirectoryPermissions memperbaiki permission direktori untuk mysql
func fixDirectoryPermissions(dir string) error {
	lg, _ := logger.Get()
	lg.Info("Fixing directory permissions", logger.String("dir", dir))

	// Set permission yang aman untuk mysql directories
	if err := os.Chmod(dir, 0750); err != nil {
		return fmt.Errorf("failed to set directory permissions: %w", err)
	}

	// TODO: Chown ke mysql user jika diperlukan
	// Perlu syscall untuk mengubah ownership

	return nil
}
