package interactive

import (
	mariadb_config "sfDBTools/utils/mariadb/config"
)

// GatherServerID mengumpulkan server ID - Task 2: modular function
func GatherServerID(config *mariadb_config.MariaDBConfigureConfig, collector *InputCollector) error {

	serverID, err := collector.CollectInt(
		"Server ID for replication",
		config.ServerID,
		"server_id",
		1,
		ValidateServerIDRange,
	)

	if err != nil {
		return err
	}

	config.ServerID = serverID
	return nil
}

// GatherPort mengumpulkan port konfigurasi - Task 2: modular function
func GatherPort(config *mariadb_config.MariaDBConfigureConfig, collector *InputCollector) error {

	port, err := collector.CollectInt(
		"MariaDB port",
		config.Port,
		"port",
		3306,
		ValidatePortRange,
	)
	if err != nil {
		return err
	}

	config.Port = port
	return nil
}

// GatherDataDirectory mengumpulkan data directory - Task 2: modular function
func GatherDataDirectory(config *mariadb_config.MariaDBConfigureConfig, collector *InputCollector) error {

	dataDir, err := collector.CollectString(
		"Data directory path",
		config.DataDir,
		"datadir",
		"/var/lib/mysql",
		ValidateAbsolutePath,
	)
	if err != nil {
		return err
	}

	config.DataDir = dataDir
	return nil
}

// GatherLogDirectory mengumpulkan log directory - Task 2: modular function
func GatherLogDirectory(config *mariadb_config.MariaDBConfigureConfig, collector *InputCollector) error {

	logDir, err := collector.CollectDirectory(
		"Log directory path",
		config.LogDir,
		"log_error", // akan extract directory dari log_error path,
		"/var/lib/mysql",
	)
	if err != nil {
		return err
	}

	config.LogDir = logDir
	return nil
}

// GatherBinlogDirectory mengumpulkan binlog directory - Task 2: modular function
func GatherBinlogDirectory(config *mariadb_config.MariaDBConfigureConfig, collector *InputCollector) error {

	binlogDir, err := collector.CollectDirectory(
		"Binary log directory path",
		config.BinlogDir,
		"log_bin", // akan extract directory dari log_bin path
		"/var/lib/mysqlbinlogs",
	)
	if err != nil {
		return err
	}

	config.BinlogDir = binlogDir
	return nil
}

// GatherEncryptionSettings mengumpulkan pengaturan enkripsi - Task 2: modular function
func GatherEncryptionSettings(config *mariadb_config.MariaDBConfigureConfig, collector *InputCollector) error {

	// Gather encryption enabled/disabled
	defaultEncrypt := config.InnodbEncryptTables
	if !defaultEncrypt && collector.Defaults.Template != nil && collector.Defaults.Template.DefaultValues["innodb-encrypt-tables"] == "ON" {
		defaultEncrypt = true
	}

	config.InnodbEncryptTables = collector.CollectBool("Enable table encryption?", defaultEncrypt)

	// If encryption enabled, gather key file
	if config.InnodbEncryptTables {
		keyFile, err := collector.CollectString(
			"Encryption key file path",
			config.EncryptionKeyFile,
			"file_key_management_filename",
			"/var/lib/mysql/encryption/keyfile",
			ValidateAbsolutePath,
		)
		if err != nil {
			return err
		}

		config.EncryptionKeyFile = keyFile
	}

	return nil
}
