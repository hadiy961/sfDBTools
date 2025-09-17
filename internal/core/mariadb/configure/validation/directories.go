package validation

import (
	"fmt"

	"sfDBTools/internal/logger"
	dir "sfDBTools/utils/fs/dir"
	mariadb_config "sfDBTools/utils/mariadb/config"
)

func validateDirectories(config *mariadb_config.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()
	directories := map[string]string{
		"data-dir":   config.DataDir,
		"binlog-dir": config.BinlogDir,
	}

	for name, p := range directories {
		lg.Debug("Validating directory", logger.String("type", name), logger.String("path", p))
		// Ensure directory exists and is writable
		if err := dir.Ensure(p); err != nil {
			return fmt.Errorf("failed to ensure %s exists: %w", name, err)
		}
	}

	if config.DataDir == config.BinlogDir {
		return fmt.Errorf("data-dir and binlog-dir cannot be the same: %s", config.DataDir)
	}

	return nil
}
