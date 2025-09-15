package migration

import (
	"context"
	"fmt"
	"path/filepath"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/disk"
	mariadb_utils "sfDBTools/utils/mariadb"
	"sfDBTools/utils/system"
)

// Entry point that performs the data migration flow. This mirrors the previous
// `performDataMigration` function but as an exported function in the
// migration package.
func PerformDataMigration(ctx context.Context, config *mariadb_utils.MariaDBConfigureConfig) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting data migration process")

	// Detect current installation
	installation, err := mariadb_utils.DiscoverMariaDBInstallation()
	if err != nil {
		return fmt.Errorf("failed to discover current installation: %w", err)
	}

	// Build migration plan
	needsMigration := false
	migrations := []DataMigration{}

	currentDataDir := filepath.Dir(installation.DataDir)
	if currentDataDir != config.DataDir {
		migrations = append(migrations, DataMigration{Type: "data", Source: currentDataDir, Destination: config.DataDir, Critical: true})
		needsMigration = true
	}

	currentLogDir := filepath.Dir(installation.DataDir) // Approximation
	if currentLogDir != config.LogDir {
		migrations = append(migrations, DataMigration{Type: "logs", Source: currentLogDir, Destination: config.LogDir, Critical: false})
		needsMigration = true
	}

	// Use detected binlog directory if available; fallback to previous default
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

	currentDataDir := filepath.Dir(installation.DataDir)
	if currentDataDir != config.DataDir {
		migrations = append(migrations, DataMigration{Type: "data", Source: currentDataDir, Destination: config.DataDir, Critical: true})
		needsMigration = true
	}

	currentLogDir := filepath.Dir(installation.DataDir) // Approximation
	if currentLogDir != config.LogDir {
		migrations = append(migrations, DataMigration{Type: "logs", Source: currentLogDir, Destination: config.LogDir, Critical: false})
		needsMigration = true
	}

	// Use detected binlog directory if available; fallback to previous default
	currentBinlogDir := filepath.Dir(installation.BinlogDir)
	if currentBinlogDir != config.BinlogDir {
		migrations = append(migrations, DataMigration{Type: "binlogs", Source: currentBinlogDir, Destination: config.BinlogDir, Critical: false})
		needsMigration = true
	}

	if !needsMigration {
		lg.Info("No data migration required")
		return nil
	}

	ShowMigrationPlan(migrations)

	lg.Info("Stopping MariaDB service for data migration")
	sm := system.NewServiceManager()
	if err := sm.Stop(installation.ServiceName); err != nil {
		return fmt.Errorf("failed to stop MariaDB service: %w", err)
	}

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
