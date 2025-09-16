package mariadb

import (
	"fmt"
	"path/filepath"
	"strings"
)

// validateVersionFormat melakukan validasi sederhana format versi
func validateVersionFormat(version string) error {
	// Versi harus berupa angka dan titik, misalnya: 10.6, 10.6.23, 11.4
	if len(version) == 0 {
		return fmt.Errorf("versi tidak boleh kosong")
	}

	// Cek apakah mengandung karakter yang valid
	for _, char := range version {
		if char != '.' && (char < '0' || char > '9') {
			return fmt.Errorf("karakter tidak valid dalam versi: %c", char)
		}
	}

	// Minimal harus ada satu titik untuk major.minor
	if !strings.Contains(version, ".") {
		return fmt.Errorf("format versi harus berupa major.minor (contoh: 10.6)")
	}

	return nil
}

// validateConfigureInput melakukan validasi input untuk MariaDB configure
func validateConfigureInput(cfg *MariaDBConfigureConfig) error {
	// Server ID validation
	if cfg.ServerID <= 0 || cfg.ServerID > 4294967295 {
		return fmt.Errorf("server_id harus antara 1 dan 4294967295, diberikan: %d", cfg.ServerID)
	}

	// Port validation
	if cfg.Port < 1024 || cfg.Port > 65535 {
		return fmt.Errorf("port harus antara 1024 dan 65535, diberikan: %d", cfg.Port)
	}

	// Directory validation - harus absolute path
	dirs := map[string]string{
		"data-dir":   cfg.DataDir,
		"log-dir":    cfg.LogDir,
		"binlog-dir": cfg.BinlogDir,
	}

	for name, dir := range dirs {
		if !filepath.IsAbs(dir) {
			return fmt.Errorf("direktori %s harus absolute path: %s", name, dir)
		}
	}

	// Directories must be different
	if cfg.DataDir == cfg.BinlogDir {
		return fmt.Errorf("data-dir dan binlog-dir tidak boleh sama: %s", cfg.DataDir)
	}
	if cfg.LogDir == cfg.BinlogDir {
		return fmt.Errorf("log-dir dan binlog-dir tidak boleh sama: %s", cfg.LogDir)
	}

	// Encryption key file validation (jika encryption enabled)
	if cfg.InnodbEncryptTables && cfg.EncryptionKeyFile != "" {
		if !filepath.IsAbs(cfg.EncryptionKeyFile) {
			return fmt.Errorf("encryption-key-file harus absolute path: %s", cfg.EncryptionKeyFile)
		}
	}

	return nil
}
