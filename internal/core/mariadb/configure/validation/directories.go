package validation

import (
	"fmt"
	"os"
	"path/filepath"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
)

func validateDirectories(config *mariadb_config.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()

	directories := map[string]string{
		"data-dir":   config.DataDir,
		"binlog-dir": config.BinlogDir,
	}

	for name, dir := range directories {
		lg.Debug("Validating directory", logger.String("type", name), logger.String("path", dir))

		if !filepath.IsAbs(dir) {
			return fmt.Errorf("%s must be absolute path: %s", name, dir)
		}

		if err := ensureDirectoryExists(dir); err != nil {
			return fmt.Errorf("failed to ensure %s exists: %w", name, err)
		}

		if err := checkDirectoryWritable(dir); err != nil {
			return fmt.Errorf("%s is not writable: %w", name, err)
		}
	}

	if config.DataDir == config.BinlogDir {
		return fmt.Errorf("data-dir and binlog-dir cannot be the same: %s", config.DataDir)
	}

	return nil
}

// ensureDirectoryExists ensures directory exists and creates it if absent.
func ensureDirectoryExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// checkDirectoryWritable writes a small test file to verify writability.
func checkDirectoryWritable(dir string) error {
	testFile := filepath.Join(dir, ".sfdbtools_write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("directory is not writable: %w", err)
	}
	_ = os.Remove(testFile)
	return nil
}
