package mariadb

import (
	"fmt"
	"os"

	"sfDBTools/internal/logger"
)

// detectConfigFiles mendeteksi file konfigurasi MariaDB
func detectConfigFiles(installation *MariaDBInstallation) error {
	lg, _ := logger.Get()
	configPaths := []string{
		"/etc/my.cnf.d/server.cnf",
		"/etc/my.cnf.d/50-server.cnf",
		"/etc/my.cnf.d/mariadb-server.cnf",
		"/etc/mysql/mariadb.conf.d/50-server.cnf",
		"/etc/mysql/conf.d/mysql.cnf",
	}
	configPaths = append(configPaths, "/etc/sfDBTools/server.cnf")
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			installation.ConfigPaths = append(installation.ConfigPaths, path)
		}
	}
	if len(installation.ConfigPaths) == 0 {
		return fmt.Errorf("tidak ditemukan file konfigurasi MariaDB")
	}
	lg.Info("Ditemukan file konfigurasi",
		logger.Int("count", len(installation.ConfigPaths)),
		logger.Strings("paths", installation.ConfigPaths))
	return nil
}
