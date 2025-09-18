package validation

import (
	"context"
	"fmt"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
)

// ValidateSystemRequirements performs the high-level validation steps.
func ValidateSystemRequirements(ctx context.Context, config *mariadb_config.MariaDBConfigureConfig) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	// Respect context cancellation early
	if ctx != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	lg.Info("Starting system requirements validation")

	if err := validateDirectories(config); err != nil {
		return fmt.Errorf("directory validation failed: %w", err)
	}

	if err := validatePort(config.Port); err != nil {
		return fmt.Errorf("port validation failed: %w", err)
	}

	if config.InnodbEncryptTables {
		if err := validateEncryptionKeyFile(config.EncryptionKeyFile); err != nil {
			return fmt.Errorf("encryption key file validation failed: %w", err)
		}
	}

	if err := validateDiskSpace(config); err != nil {
		return fmt.Errorf("disk space validation failed: %w", err)
	}

	if err := validateDirectoryPermissions(config); err != nil {
		return fmt.Errorf("directory permissions validation failed: %w", err)
	}

	lg.Info("All system requirements validation passed")
	return nil
}
