package service

import (
	"fmt"

	sfdbconfig "sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/terminal"
)

// FinalizeConfiguration performs finalization steps (update app config + summary)
func FinalizeConfiguration(config *mariadb_config.MariaDBConfigureConfig) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Finalizing configuration")

	// Update application config file
	if err := updateApplicationConfig(config); err != nil {
		return fmt.Errorf("failed to update application config: %w", err)
	}

	// Show success summary
	showSuccessSummary(config)

	lg.Info("Configuration finalization completed")
	return nil
}

// updateApplicationConfig updates sfDBTools config.yaml mariadb section
func updateApplicationConfig(config *mariadb_config.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()
	lg.Info("Updating application configuration file")

	// Create config updater
	updater, err := sfdbconfig.NewConfigUpdater()
	if err != nil {
		return fmt.Errorf("failed to create config updater: %w", err)
	}

	updates := make(map[string]interface{})

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
	updates["innodb_encrypt_tables"] = config.InnodbEncryptTables

	if err := updater.UpdateMariaDBConfig(updates); err != nil {
		return fmt.Errorf("failed to update config file: %w", err)
	}

	lg.Info("Application configuration update completed",
		logger.String("config_file", updater.GetConfigFilePath()))
	return nil
}

// showSuccessSummary prints a short summary to the terminal
func showSuccessSummary(config *mariadb_config.MariaDBConfigureConfig) {
	terminal.PrintSuccess("MariaDB Configuration Completed Successfully!")
	println()
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
	println()
	terminal.PrintInfo("MariaDB service is running and ready to accept connections.")
	terminal.PrintInfo("You can now connect to MariaDB using the new configuration.")
}
