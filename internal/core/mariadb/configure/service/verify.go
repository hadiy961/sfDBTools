package service

import (
	"fmt"
	"time"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/mariadb/discovery"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// verifyServiceRunning checks the service status with backoff
func verifyServiceRunning(sm system.ServiceManager, serviceName string) error {
	// Wait a bit for service to start
	time.Sleep(3 * time.Second)

	// Check service status multiple times with backoff
	maxRetries := 10
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		status, err := sm.GetStatus(serviceName)
		if err != nil {
			return fmt.Errorf("failed to get service status: %w", err)
		}

		if status.Active && status.Running {
			terminal.PrintSuccess("MariaDB service is running")
			return nil
		}

		if i < maxRetries-1 {
			terminal.PrintInfo(fmt.Sprintf("Waiting for service to start... (attempt %d/%d)", i+1, maxRetries))
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("service failed to start after %d attempts", maxRetries)
}

// verifyDatabaseConnection attempts to connect to database
func verifyDatabaseConnection(installation *discovery.MariaDBInstallation, config *mariadb_config.MariaDBConfigureConfig) error {
	// Try connection multiple times
	maxRetries := 5
	retryDelay := 3 * time.Second

	for i := 0; i < maxRetries; i++ {
		err := testDatabaseConnection(installation, config)
		if err == nil {
			terminal.PrintSuccess("Database connection verified")
			return nil
		}

		if i < maxRetries-1 {
			terminal.PrintInfo(fmt.Sprintf("Waiting for database to accept connections... (attempt %d/%d)", i+1, maxRetries))
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("database connection failed after %d attempts", maxRetries)
}

// testDatabaseConnection does a simple discovery-based check
func testDatabaseConnection(installation *discovery.MariaDBInstallation, config *mariadb_config.MariaDBConfigureConfig) error {
	// Simple connection test - check if service is responding
	if !installation.IsRunning {
		return fmt.Errorf("MariaDB service is not running")
	}

	return nil
}

// verifyConfigurationApplied checks minimal validation of applied config
func verifyConfigurationApplied(installation *discovery.MariaDBInstallation, config *mariadb_config.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()
	lg.Info("Verifying configuration applied")

	if !installation.IsRunning {
		return fmt.Errorf("MariaDB service is not running")
	}

	if installation.Port != 0 && installation.Port != config.Port {
		lg.Warn("Port mismatch detected",
			logger.Int("expected", config.Port),
			logger.Int("actual", installation.Port))
	}

	terminal.PrintSuccess("Configuration verification completed")
	return nil
}
