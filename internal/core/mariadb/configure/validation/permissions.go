package validation

import (
	"fmt"

	"sfDBTools/internal/logger"
	dir "sfDBTools/utils/fs/dir"
	mariadb_config "sfDBTools/utils/mariadb/config"
)

// Note: we intentionally use the fs/dir.Manager for file/permission operations so
// behavior and fallbacks are centralized in utils/fs.
func validateDirectoryPermissions(config *mariadb_config.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()
	lg.Debug("Validating directory permissions")

	directories := []string{config.DataDir, config.BinlogDir}

	for _, d := range directories {
		// Ensure directory exists
		if err := dir.Ensure(d); err != nil {
			return fmt.Errorf("failed to ensure directory %s: %w", d, err)
		}

		// Ensure permissions/ownership (attempt to set mysql:mysql, mode 0750)
		if err := dir.EnsureWithPermissions(d, 0750, "mysql", "mysql"); err != nil {
			lg.Warn("Failed to ensure permissions via fs helpers", logger.String("dir", d), logger.Error(err))
			return fmt.Errorf("failed to fix permissions for %s: %w", d, err)
		}
	}

	return nil
}
