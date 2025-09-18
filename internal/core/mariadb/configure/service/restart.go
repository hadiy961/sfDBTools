package service

import (
	"context"
	"fmt"
	"time"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/mariadb/discovery"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// RestartAndVerifyService performs restart and verification steps (steps 20-23)
func RestartAndVerifyService(ctx context.Context, config *mariadb_config.MariaDBConfigureConfig) error {
	_ = ctx
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting service restart and verification")

	// Get current installation info
	installation, err := discovery.DiscoverMariaDBInstallation()
	if err != nil {
		return fmt.Errorf("failed to discover MariaDB installation: %w", err)
	}

	sm := system.NewServiceManager()
	// Determine service name (fallback to common names if discovery failed)
	svcName := installation.ServiceName
	if svcName == "" {
		candidates := []string{"mariadb", "mysql", "mysqld"}
		found := ""
		for _, c := range candidates {
			if sm.IsActive(c) || sm.IsEnabled(c) {
				found = c
				break
			}
		}
		if found == "" {
			lg.Warn("Service name not discovered; attempting common service names", logger.Strings("candidates", []string{"mariadb", "mysql", "mysqld"}))
			svcName = "mariadb"
		} else {
			svcName = found
		}
	}

	lg.Info("Restarting MariaDB service", logger.String("service", svcName))
	terminal.PrintInfo("Restarting MariaDB service...")

	// Try stopping using candidate names until one succeeds
	stopErr := fmt.Errorf("no stop attempted")
	tried := map[string]struct{}{}
	candidates := []string{svcName, "mariadb", "mysql", "mysqld"}
	for _, name := range candidates {
		if name == "" {
			continue
		}
		if _, ok := tried[name]; ok {
			continue
		}
		tried[name] = struct{}{}
		if err := sm.Stop(name); err == nil {
			stopErr = nil
			svcName = name
			break
		} else {
			stopErr = err
			lg.Debug("Stop attempt failed for candidate service", logger.String("candidate", name), logger.Error(err))
		}
	}
	if stopErr != nil {
		return fmt.Errorf("failed to stop MariaDB service: %w", stopErr)
	}

	time.Sleep(2 * time.Second)

	// Try starting the discovered/service name
	if err := sm.Start(svcName); err != nil {
		startErr := err
		for _, name := range []string{"mariadb", "mysql", "mysqld"} {
			if name == svcName {
				continue
			}
			if err2 := sm.Start(name); err2 == nil {
				startErr = nil
				svcName = name
				break
			} else {
				lg.Debug("Start attempt failed for candidate service", logger.String("candidate", name), logger.Error(err2))
			}
		}
		if startErr != nil {
			return fmt.Errorf("failed to start MariaDB service: %w", startErr)
		}
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
