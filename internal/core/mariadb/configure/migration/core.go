package migration

import (
	"context"
	"fmt"
	"path/filepath"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/mariadb/discovery"
	"sfDBTools/utils/system"
)

// PerformDataMigrationWithInstallation performs migration using an already-discovered installation
// This avoids re-running discovery when the caller already has the installation info.
func PerformDataMigrationWithInstallation(ctx context.Context, config *mariadb_config.MariaDBConfigureConfig, installation *discovery.MariaDBInstallation) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting data migration process (using provided installation)")

	// Build migration plan
	needsMigration := false
	migrations := []DataMigration{}

	// Clean paths to avoid false positives due to trailing slashes or relative segments
	currentDataDir := filepath.Clean(installation.DataDir)
	targetDataDir := filepath.Clean(config.DataDir)
	if currentDataDir != targetDataDir {
		migrations = append(migrations, DataMigration{Type: "data", Source: currentDataDir, Destination: targetDataDir, Critical: true})
		needsMigration = true
	}

	currentLogDir := filepath.Clean(installation.LogDir)
	targetLogDir := filepath.Clean(config.LogDir)
	if currentLogDir != targetLogDir {
		migrations = append(migrations, DataMigration{Type: "logs", Source: currentLogDir, Destination: targetLogDir, Critical: false})
		needsMigration = true
	}

	currentBinlogDir := filepath.Clean(installation.BinlogDir)
	targetBinlogDir := filepath.Clean(config.BinlogDir)
	if currentBinlogDir != targetBinlogDir {
		migrations = append(migrations, DataMigration{Type: "binlogs", Source: currentBinlogDir, Destination: targetBinlogDir, Critical: false})
		needsMigration = true
	}

	if !needsMigration {
		lg.Info("No data migration required")
		return nil
	}

	// Log the planned migrations for visibility
	for _, m := range migrations {
		lg.Info(fmt.Sprintf("Planned migration: type=%s, source=%s, destination=%s, critical=%v", m.Type, m.Source, m.Destination, m.Critical))
	}

	lg.Info("Stopping MariaDB service for data migration")
	sm := system.NewServiceManager()

	// If service name is empty, continue but warn
	if installation.ServiceName == "" {
		lg.Warn("installation.ServiceName is empty; skipping service stop/start")
	} else {
		if err := sm.Stop(installation.ServiceName); err != nil {
			return fmt.Errorf("failed to stop MariaDB service: %w", err)
		}

		// Ensure we attempt to start the service again when this function returns.
		defer func() {
			if err := sm.Start(installation.ServiceName); err != nil {
				lg.Warn(fmt.Sprintf("failed to start MariaDB service after migration: %v", err))
			} else {
				lg.Info("MariaDB service started after migration")
			}
		}()
	}

	for _, m := range migrations {
		if err := PerformSingleMigration(m); err != nil {
			if m.Critical {
				return fmt.Errorf("critical migration failed: %w", err)
			}
			lg.Warn(fmt.Sprintf("Non-critical migration failed for %s -> %s: %v", m.Source, m.Destination, err))
		}
	}

	lg.Info("Data migration completed successfully")
	return nil
}
