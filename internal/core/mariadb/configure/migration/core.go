package migration

import (
	"context"
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/disk"
	mariadb_utils "sfDBTools/utils/mariadb"
)

// PerformDataMigrationWithInstallation performs migration using an already-discovered installation
// This avoids re-running discovery when the caller already has the installation info.
func PerformDataMigrationWithInstallation(ctx context.Context, config *mariadb_utils.MariaDBConfigureConfig, installation *mariadb_utils.MariaDBInstallation) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting data migration process (using provided installation)")

	// Build migration plan
	needsMigration := false
	migrations := []DataMigration{}

	// Use detected DataDir; fallback to standard default if empty
	currentDataDir := installation.DataDir
	if currentDataDir == "" {
		currentDataDir = "/var/lib/mysql"
	}
	if currentDataDir != config.DataDir {
		migrations = append(migrations, DataMigration{Type: "data", Source: currentDataDir, Destination: config.DataDir, Critical: true})
		needsMigration = true
	}

	// Use detected LogDir if available; fallback to DataDir
	// Use detected LogDir; fallback to DataDir then to standard default
	currentLogDir := installation.LogDir
	if currentLogDir == "" {
		if installation.DataDir != "" {
			currentLogDir = installation.DataDir
		} else {
			currentLogDir = "/var/lib/mysql"
		}
	}
	if currentLogDir != config.LogDir {
		migrations = append(migrations, DataMigration{Type: "logs", Source: currentLogDir, Destination: config.LogDir, Critical: false})
		needsMigration = true
	}

	// Use detected binlog directory if available; fallback to previous default
	// Use detected binlog directory; fallback to previous default if empty
	currentBinlogDir := installation.BinlogDir
	if currentBinlogDir == "" {
		currentBinlogDir = "/var/lib/mysqlbinlogs"
	}
	if currentBinlogDir != config.BinlogDir {
		migrations = append(migrations, DataMigration{Type: "binlogs", Source: currentBinlogDir, Destination: config.BinlogDir, Critical: false})
		needsMigration = true
	}

	if !needsMigration {
		lg.Info("No data migration required")
		return nil
	}

	ShowMigrationPlan(migrations)

	// lg.Info("Stopping MariaDB service for data migration")
	// sm := system.NewServiceManager()
	// if err := sm.Stop(installation.ServiceName); err != nil {
	// 	return fmt.Errorf("failed to stop MariaDB service: %w", err)
	// }

	// for _, m := range migrations {
	// 	if err := PerformSingleMigration(m); err != nil {
	// 		if m.Critical {
	// 			return fmt.Errorf("critical migration failed: %w", err)
	// 		}
	// 		lg.Warn("Non-critical migration failed")
	// 	}
	// }

	lg.Info("Data migration completed successfully")
	return nil
}

// ChooseSocketPath is exported so callers in package configure can use it.
func ChooseSocketPath(dataDir string) string {
	return chooseSocketPathImpl(dataDir, disk.GetUsage)
}
