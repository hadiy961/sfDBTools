package validation

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/disk"
	mariadb_config "sfDBTools/utils/mariadb/config"
)

func validateDiskSpace(config *mariadb_config.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()
	lg.Debug("Validating disk space")

	directories := []string{config.DataDir, config.BinlogDir}

	for _, dir := range directories {
		minSpace := int64(1024 * 1024 * 1024)

		freeSpace, err := disk.GetFreeBytes(dir)
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
