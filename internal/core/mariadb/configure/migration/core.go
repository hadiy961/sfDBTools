package migration

import (
	"context"
	"fmt"

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

	currentDataDir := installation.DataDir
	if currentDataDir != config.DataDir {
		migrations = append(migrations, DataMigration{Type: "data", Source: currentDataDir, Destination: config.DataDir, Critical: true})
		needsMigration = true
	}

	currentLogDir := installation.LogDir
	if currentLogDir != config.LogDir {
		migrations = append(migrations, DataMigration{Type: "logs", Source: currentLogDir, Destination: config.LogDir, Critical: false})
		needsMigration = true
	}

	currentBinlogDir := installation.BinlogDir
	if currentBinlogDir != config.BinlogDir {
		migrations = append(migrations, DataMigration{Type: "binlogs", Source: currentBinlogDir, Destination: config.BinlogDir, Critical: false})
		needsMigration = true
	}

	if !needsMigration {
		lg.Info("No data migration required")
		return nil
	}

	// ShowMigrationPlan(migrations)

	lg.Info("Stopping MariaDB service for data migration")
	sm := system.NewServiceManager()
	if err := sm.Stop(installation.ServiceName); err != nil {
		return fmt.Errorf("failed to stop MariaDB service: %w", err)
	}

	for _, m := range migrations {
		if err := PerformSingleMigration(m); err != nil {
			if m.Critical {
				return fmt.Errorf("critical migration failed: %w", err)
			}
			lg.Warn("Non-critical migration failed")
		}
	}

	lg.Info("Data migration completed successfully")
	return nil
}
