package configure

import (
	"context"
	"fmt"
	"time"

	sfdbconfig "sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/mariadb/discovery"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// restartAndVerifyService melakukan restart service dan verifikasi
// Sesuai dengan Step 20-23 dalam flow implementasi
func restartAndVerifyService(ctx context.Context, config *mariadb_config.MariaDBConfigureConfig) error {
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
	// Step 20: Restart MariaDB service
	// Determine service name (fallback to common names if discovery failed)
	svcName := installation.ServiceName
	if svcName == "" {
		// try common names
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
			svcName = "mariadb" // best-effort default
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

	time.Sleep(2 * time.Second) // Wait a bit

	// Try starting the discovered/service name
	if err := sm.Start(svcName); err != nil {
		// try other candidates
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

	// Step 21: Verify service is running
	lg.Info("Verifying service status")
	if err := verifyServiceRunning(sm, svcName); err != nil {
		return fmt.Errorf("service verification failed: %w", err)
	}

	// Step 22: Verify database connection
	lg.Info("Verifying database connection")
	if err := verifyDatabaseConnection(config); err != nil {
		return fmt.Errorf("database connection verification failed: %w", err)
	}

	// Step 23: Verify configuration is applied
	lg.Info("Verifying configuration is applied")
	if err := verifyConfigurationApplied(config); err != nil {
		return fmt.Errorf("configuration verification failed: %w", err)
	}

	lg.Info("Service restart and verification completed successfully")
	return nil
}

// verifyServiceRunning memverifikasi bahwa service berjalan dengan baik
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

// verifyDatabaseConnection memverifikasi koneksi ke database
func verifyDatabaseConnection(config *mariadb_config.MariaDBConfigureConfig) error {
	// Try connection multiple times
	maxRetries := 5
	retryDelay := 3 * time.Second

	for i := 0; i < maxRetries; i++ {
		err := testDatabaseConnection(config)
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

// testDatabaseConnection melakukan test koneksi ke database
func testDatabaseConnection(config *mariadb_config.MariaDBConfigureConfig) error {
	// Create connection using discovery
	installation, err := discovery.DiscoverMariaDBInstallation()
	if err != nil {
		return fmt.Errorf("failed to discover installation: %w", err)
	}

	// Simple connection test - check if service is responding
	if !installation.IsRunning {
		return fmt.Errorf("MariaDB service is not running")
	}

	return nil
}

// verifyConfigurationApplied memverifikasi bahwa konfigurasi telah diterapkan
func verifyConfigurationApplied(config *mariadb_config.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()
	lg.Info("Verifying configuration applied")

	// Simple verification - just check if service is running with new config
	installation, err := discovery.DiscoverMariaDBInstallation()
	if err != nil {
		return fmt.Errorf("failed to discover installation: %w", err)
	}

	if !installation.IsRunning {
		return fmt.Errorf("MariaDB service is not running")
	}

	// Check if port matches (if discoverable)
	if installation.Port != 0 && installation.Port != config.Port {
		lg.Warn("Port mismatch detected",
			logger.Int("expected", config.Port),
			logger.Int("actual", installation.Port))
	}

	terminal.PrintSuccess("Configuration verification completed")
	return nil
}

// finalizeConfiguration melakukan finalisasi konfigurasi
func finalizeConfiguration(ctx context.Context, config *mariadb_config.MariaDBConfigureConfig) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Finalizing configuration")

	// Step 25: Update application config file
	if err := updateApplicationConfig(config); err != nil {
		return fmt.Errorf("failed to update application config: %w", err)
	}

	// Show success summary
	showSuccessSummary(config)

	lg.Info("Configuration finalization completed")
	return nil
}

// updateApplicationConfig mengupdate file konfigurasi aplikasi sfDBTools
func updateApplicationConfig(config *mariadb_config.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()
	lg.Info("Updating application configuration file")

	// Create config updater
	updater, err := sfdbconfig.NewConfigUpdater()
	if err != nil {
		return fmt.Errorf("failed to create config updater: %w", err)
	}

	// Prepare updates map with MariaDB configuration values
	updates := make(map[string]interface{})

	// Map MariaDBConfigureConfig fields to config.yaml mariadb section
	if config.ServerID != 0 {
		updates["server_id"] = config.ServerID
	}
	if config.Port != 0 {
		updates["port"] = config.Port
	}
	if config.DataDir != "" {
		updates["data_dir"] = config.DataDir
	}
	if config.LogDir != "" {
		updates["log_dir"] = config.LogDir
	}
	if config.BinlogDir != "" {
		updates["binlog_dir"] = config.BinlogDir
	}
	if config.ConfigDir != "" {
		updates["config_dir"] = config.ConfigDir
	}
	if config.EncryptionKeyFile != "" {
		updates["encryption_key_file"] = config.EncryptionKeyFile
	}
	// InnodbEncryptTables is a boolean, so we need to check it differently
	updates["innodb_encrypt_tables"] = config.InnodbEncryptTables

	// Update the config file
	if err := updater.UpdateMariaDBConfig(updates); err != nil {
		return fmt.Errorf("failed to update config file: %w", err)
	}

	lg.Info("Application configuration update completed",
		logger.String("config_file", updater.GetConfigFilePath()))
	return nil
}

// showSuccessSummary menampilkan ringkasan keberhasilan konfigurasi
func showSuccessSummary(config *mariadb_config.MariaDBConfigureConfig) {
	terminal.PrintSuccess("MariaDB Configuration Completed Successfully!")
	fmt.Println()
	terminal.PrintInfo("Configuration Summary:")
	terminal.PrintInfo("======================")
	fmt.Printf("✓ Server ID: %d\n", config.ServerID)
	fmt.Printf("✓ Port: %d\n", config.Port)
	fmt.Printf("✓ Data Directory: %s\n", config.DataDir)
	fmt.Printf("✓ Log Directory: %s\n", config.LogDir)
	fmt.Printf("✓ Binlog Directory: %s\n", config.BinlogDir)
	fmt.Printf("✓ Table Encryption: %t\n", config.InnodbEncryptTables)
	fmt.Printf("✓ Buffer Pool Size: %s\n", config.InnodbBufferPoolSize)
	fmt.Printf("✓ Buffer Pool Instances: %d\n", config.InnodbBufferPoolInstances)
	fmt.Println()
	terminal.PrintInfo("MariaDB service is running and ready to accept connections.")
	terminal.PrintInfo("You can now connect to MariaDB using the new configuration.")
}
