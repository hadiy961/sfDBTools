package service

import (
	"context"
	"fmt"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/mariadb/discovery"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// RestartAndVerifyService performs restart and verification steps (steps 20-23)
func RestartAndVerifyService(ctx context.Context, config *mariadb_config.MariaDBConfigureConfig, installation *discovery.MariaDBInstallation) error {
	_ = ctx
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting service restart and verification")

	sm := system.NewServiceManager()
	// Determine service name (fallback to common names if discovery failed)
	svcName := installation.ServiceName

	lg.Info("Restarting MariaDB service", logger.String("service", svcName))
	terminal.PrintInfo("Restarting MariaDB service...")

	//Restart Mariadb service
	if err := sm.Restart(svcName); err != nil {
		return fmt.Errorf("failed to restart service %s: %w", svcName, err)
	}

	// Verify service running
	lg.Info("Verifying service status")
	if err := verifyServiceRunning(sm, svcName); err != nil {
		return fmt.Errorf("service verification failed: %w", err)
	}

	// Verify database connection
	lg.Info("Verifying database connection")
	if err := verifyDatabaseConnection(installation, config); err != nil {
		return fmt.Errorf("database connection verification failed: %w", err)
	}

	// Verify configuration applied
	lg.Info("Verifying configuration is applied")
	if err := verifyConfigurationApplied(installation, config); err != nil {
		return fmt.Errorf("configuration verification failed: %w", err)
	}

	lg.Info("Service restart and verification completed successfully")
	return nil
}
