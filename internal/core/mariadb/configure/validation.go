package configure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/disk"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/system"
)

// validateSystemRequirements melakukan validasi sistem dan input user
// Sesuai dengan Step 7-11 dalam flow implementasi
func validateSystemRequirements(ctx context.Context, config *mariadb_config.MariaDBConfigureConfig) error {
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
func validateDirectories(config *mariadb_config.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()

	directories := map[string]string{
		"data-dir":   config.DataDir,
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
	if config.DataDir == config.BinlogDir {
		return fmt.Errorf("data-dir and binlog-dir cannot be the same: %s", config.DataDir)
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
	pi, err := system.CheckPortConflict(port)
	if err != nil {
		// jika gagal mendapatkan info process, fallback ke simple check
		if isPortInUse(port) {
			return fmt.Errorf("port %d is already in use", port)
		}
		return nil
	}

	// Jika tersedia, ok
	if pi.Status == "available" {
		return nil
	}

	// Jika port sedang digunakan, periksa apakah listener adalah MariaDB (mysqld/mariadbd)
	proc := strings.ToLower(pi.ProcessName)
	if proc == "mysqld" || proc == "mariadbd" || strings.Contains(proc, "mysqld") || strings.Contains(proc, "mariadb") {
		lg.Info("Port digunakan oleh MariaDB, mengabaikan konflik", logger.Int("port", port), logger.String("process", pi.ProcessName))
		return nil
	}

	return fmt.Errorf("port %d is already in use by process %s (pid=%s)", port, pi.ProcessName, pi.PID)

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
func validateDiskSpace(config *mariadb_config.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()
	lg.Debug("Validating disk space")

	directories := []string{config.DataDir, config.BinlogDir}

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
func validateDirectoryPermissions(config *mariadb_config.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()
	lg.Debug("Validating directory permissions")

	directories := []string{config.DataDir, config.BinlogDir}

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

	// Untuk implementasi, cek ownership dan permission
	mode := info.Mode()
	if mode&0020 != 0 || mode&0002 != 0 { // group-writable atau world-writable
		return nil
	}

	// Periksa owner uid/gid dari file stat
	statT, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		// Jika tidak bisa mendapatkan stat info, anggap tidak accessible
		return fmt.Errorf("cannot determine owner of %s", dir)
	}

	// Jika owner adalah mysql (uid 992 biasanya, but we check by name later in fix), return nil
	// We'll treat non-root/non-mysql ownership as not accessible to mysql
	if statT.Uid == 0 {
		// owned by root, likely not writable by mysql
		return fmt.Errorf("directory %s owned by root (uid 0)", dir)
	}

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

	// Chown ke mysql:mysql - coba lookup user mysql uid/gid via system call
	// Default to UID/GID commonly used by distributions if lookup not available
	mysqlUID := uint32(992)
	mysqlGID := uint32(991)

	// Attempt to get uid/gid from passwd entry if available
	if pw, err := os.ReadFile("/etc/passwd"); err == nil {
		// simple parse: look for line starting with "mysql:"
		lines := strings.Split(string(pw), "\n")
		for _, l := range lines {
			if strings.HasPrefix(l, "mysql:") {
				fields := strings.Split(l, ":")
				if len(fields) >= 4 {
					// fields[2]=uid, fields[3]=gid
					var uid, gid int
					fmt.Sscanf(fields[2], "%d", &uid)
					fmt.Sscanf(fields[3], "%d", &gid)
					if uid >= 0 && gid >= 0 {
						mysqlUID = uint32(uid)
						mysqlGID = uint32(gid)
					}
				}
				break
			}
		}
	}

	if err := os.Chown(dir, int(mysqlUID), int(mysqlGID)); err != nil {
		return fmt.Errorf("failed to chown directory to mysql:mysql: %w", err)
	}

	return nil
}
