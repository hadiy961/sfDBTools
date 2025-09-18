package mariadb

import (
	"fmt"

	"sfDBTools/internal/config"
	"sfDBTools/utils/common"
	"sfDBTools/utils/database"
	"sfDBTools/utils/mariadb/discovery"

	"github.com/spf13/cobra"
)

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
			if cfg.MariaDB.Version != "" {
				version = cfg.MariaDB.Version
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

// CreateDatabaseConfigFromInstallation creates a basic database.Config from installation info
func CreateDatabaseConfigFromInstallation(installation *discovery.MariaDBInstallation) *database.Config {
	if installation == nil {
		return nil
	}
	return &database.Config{
		Host:     "localhost",
		Port:     installation.Port,
		User:     "root",
		Password: "",
		DBName:   "",
	}
}
