package mariadb

import (
	"fmt"
	"strings"

	"sfDBTools/internal/config"
	"sfDBTools/utils/common"

	"github.com/spf13/cobra"
)

// MariaDBInstallConfig berisi konfigurasi untuk instalasi MariaDB
type MariaDBInstallConfig struct {
	Version        string // Versi MariaDB yang akan diinstall
	NonInteractive bool   // Mode non-interactive
}

// MariaDBRemoveConfig berisi konfigurasi untuk penghapusan MariaDB
type MariaDBRemoveConfig struct {
	RemoveData       bool   // Hapus data directory (/var/lib/mysql)
	RemoveConfig     bool   // Hapus file konfigurasi (/etc/mysql, /etc/my.cnf)
	RemoveRepository bool   // Hapus repository MariaDB
	RemoveUser       bool   // Hapus user mysql dari sistem
	Force            bool   // Force removal tanpa konfirmasi
	BackupData       bool   // Backup data sebelum dihapus
	BackupPath       string // Path untuk backup data
	NonInteractive   bool   // Mode non-interactive
}

// ResolveMariaDBInstallConfig membaca flags/env dan default dari config file
func ResolveMariaDBInstallConfig(cmd *cobra.Command) (*MariaDBInstallConfig, error) {
	// Baca konfigurasi dari flags dan environment variables
	version := common.GetStringFlagOrEnv(cmd, "version", "SFDBTOOLS_MARIADB_VERSION", "")
	nonInteractive := common.GetBoolFlagOrEnv(cmd, "non-interactive", "SFDBTOOLS_NON_INTERACTIVE", false)

	// Jika versi tidak ditentukan melalui flag/env, ambil dari config file
	if version == "" {
		cfg, err := config.Get()
		if err != nil {
			// Jika config tidak dapat dimuat, gunakan default hardcoded
			version = "10.6.23"
		} else {
			// Ambil dari config file
			if cfg.MariaDB.Installation.Version != "" {
				version = cfg.MariaDB.Installation.Version
			} else {
				// Fallback ke default hardcoded jika config kosong
				version = "10.6.23"
			}
		}
	}

	cfg := &MariaDBInstallConfig{
		Version:        version,
		NonInteractive: nonInteractive,
	}

	// Validasi konfigurasi basic (format saja)
	if err := validateVersionFormat(cfg.Version); err != nil {
		return nil, fmt.Errorf("format versi tidak valid: %w", err)
	}

	return cfg, nil
}

// ResolveMariaDBRemoveConfig membaca flags/env untuk konfigurasi penghapusan
func ResolveMariaDBRemoveConfig(cmd *cobra.Command) (*MariaDBRemoveConfig, error) {
	// Baca konfigurasi dari flags dan environment variables
	removeData := common.GetBoolFlagOrEnv(cmd, "remove-data", "SFDBTOOLS_REMOVE_DATA", false)
	removeConfig := common.GetBoolFlagOrEnv(cmd, "remove-config", "SFDBTOOLS_REMOVE_CONFIG", false)
	removeRepository := common.GetBoolFlagOrEnv(cmd, "remove-repository", "SFDBTOOLS_REMOVE_REPOSITORY", false)
	removeUser := common.GetBoolFlagOrEnv(cmd, "remove-user", "SFDBTOOLS_REMOVE_USER", false)
	force := common.GetBoolFlagOrEnv(cmd, "force", "SFDBTOOLS_FORCE", false)
	backupData := common.GetBoolFlagOrEnv(cmd, "backup-data", "SFDBTOOLS_BACKUP_DATA", false)
	backupPath := common.GetStringFlagOrEnv(cmd, "backup-path", "SFDBTOOLS_BACKUP_PATH", "/tmp/mariadb_backup")
	nonInteractive := common.GetBoolFlagOrEnv(cmd, "non-interactive", "SFDBTOOLS_NON_INTERACTIVE", false)

	cfg := &MariaDBRemoveConfig{
		RemoveData:       removeData,
		RemoveConfig:     removeConfig,
		RemoveRepository: removeRepository,
		RemoveUser:       removeUser,
		Force:            force,
		BackupData:       backupData,
		BackupPath:       backupPath,
		NonInteractive:   nonInteractive,
	}

	return cfg, nil
}

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
