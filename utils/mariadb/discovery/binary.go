package mariadb

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"sfDBTools/internal/logger"
)

// detectMariaDBBinary mendeteksi binary MariaDB/MySQL
func detectMariaDBBinary(installation *MariaDBInstallation) error {
	binaries := []string{"mariadb", "mysql", "mysqld", "mariadbd"}
	for _, binary := range binaries {
		path, err := exec.LookPath(binary)
		if err == nil {
			installation.BinaryPath = path
			installation.IsInstalled = true
			return nil
		}
	}
	standardPaths := []string{"/usr/bin/mariadb", "/usr/bin/mysql", "/usr/sbin/mysqld", "/usr/sbin/mariadbd", "/usr/local/bin/mariadb", "/usr/local/bin/mysql"}
	for _, path := range standardPaths {
		if _, err := os.Stat(path); err == nil {
			installation.BinaryPath = path
			installation.IsInstalled = true
			return nil
		}
	}
	return fmt.Errorf("binary MariaDB/MySQL tidak ditemukan")
}

// detectMariaDBVersion mendeteksi versi MariaDB
func detectMariaDBVersion(installation *MariaDBInstallation) error {
	lg, _ := logger.Get()
	cmd := exec.Command(installation.BinaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("gagal menjalankan %s --version: %w", installation.BinaryPath, err)
	}
	version := parseVersionFromOutput(string(output))
	if version != "" {
		installation.Version = version
		lg.Info("Terdeteksi versi MariaDB", logger.String("version", version))
		return nil
	}
	return fmt.Errorf("gagal parsing versi dari output: %s", string(output))
}

// parseVersionFromOutput parsing versi dari output command
func parseVersionFromOutput(output string) string {
	patterns := []string{`mariadb\s+Ver\s+(\d+\.\d+\.\d+)`, `MariaDB\s+(\d+\.\d+\.\d+)`, `mysql\s+Ver\s+(\d+\.\d+\.\d+).*MariaDB`, `(\d+\.\d+\.\d+)-MariaDB`}
	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		matches := re.FindStringSubmatch(output)
		if len(matches) >= 2 {
			return matches[1]
		}
	}
	return ""
}
