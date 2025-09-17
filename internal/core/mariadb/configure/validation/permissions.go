package validation

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
)

func validateDirectoryPermissions(config *mariadb_config.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()
	lg.Debug("Validating directory permissions")

	directories := []string{config.DataDir, config.BinlogDir}

	for _, dir := range directories {
		if err := checkMySQLUserAccess(dir); err != nil {
			lg.Warn("MySQL user access check failed", logger.String("dir", dir), logger.Error(err))

			if err := fixDirectoryPermissions(dir); err != nil {
				return fmt.Errorf("failed to fix permissions for %s: %w", dir, err)
			}
		}
	}

	return nil
}

// checkMySQLUserAccess checks whether the directory is writable/accessible by mysql
func checkMySQLUserAccess(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}

	mode := info.Mode()
	if mode&0020 != 0 || mode&0002 != 0 {
		return nil
	}

	statT, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("cannot determine owner of %s", dir)
	}

	if statT.Uid == 0 {
		return fmt.Errorf("directory %s owned by root (uid 0)", dir)
	}

	return nil
}

// fixDirectoryPermissions attempts to set appropriate ownership/permissions
func fixDirectoryPermissions(dir string) error {
	lg, _ := logger.Get()
	lg.Info("Fixing directory permissions", logger.String("dir", dir))

	if err := os.Chmod(dir, 0750); err != nil {
		return fmt.Errorf("failed to set directory permissions: %w", err)
	}

	mysqlUID := uint32(992)
	mysqlGID := uint32(991)

	if pw, err := os.ReadFile("/etc/passwd"); err == nil {
		lines := strings.Split(string(pw), "\n")
		for _, l := range lines {
			if strings.HasPrefix(l, "mysql:") {
				fields := strings.Split(l, ":")
				if len(fields) >= 4 {
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
