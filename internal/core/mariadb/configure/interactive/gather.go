package interactive

import (
	mariadb_utils "sfDBTools/utils/mariadb"
)

// GatherServerID mengumpulkan server ID - Task 2: modular function
func GatherServerID(config *mariadb_utils.MariaDBConfigureConfig, collector *InputCollector) error {
	appDefaults := collector.Defaults.GetAppConfigDefaults()

	serverID, err := collector.CollectInt(
		"Server ID for replication",
		config.ServerID,
		"server_id",
		appDefaults.ServerID,
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
func GatherPort(config *mariadb_utils.MariaDBConfigureConfig, collector *InputCollector) error {
	appDefaults := collector.Defaults.GetAppConfigDefaults()

	port, err := collector.CollectInt(
		"MariaDB port",
		config.Port,
		"port",
		appDefaults.Port,
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
func GatherDataDirectory(config *mariadb_utils.MariaDBConfigureConfig, collector *InputCollector) error {
	appDefaults := collector.Defaults.GetAppConfigDefaults()

	dataDir, err := collector.CollectString(
		"Data directory path",
		config.DataDir,
		"datadir",
		appDefaults.DataDir,
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
func GatherLogDirectory(config *mariadb_utils.MariaDBConfigureConfig, collector *InputCollector) error {
	appDefaults := collector.Defaults.GetAppConfigDefaults()

	logDir, err := collector.CollectDirectory(
		"Log directory path",
		config.LogDir,
		"log_error", // akan extract directory dari log_error path
		appDefaults.LogDir,
		"/var/lib/mysql",
	)
	if err != nil {
		return err
	}

	config.LogDir = logDir
	return nil
}

// GatherBinlogDirectory mengumpulkan binlog directory - Task 2: modular function
func GatherBinlogDirectory(config *mariadb_utils.MariaDBConfigureConfig, collector *InputCollector) error {
	appDefaults := collector.Defaults.GetAppConfigDefaults()

	binlogDir, err := collector.CollectDirectory(
		"Binary log directory path",
		config.BinlogDir,
		"log_bin", // akan extract directory dari log_bin path
		appDefaults.BinlogDir,
		"/var/lib/mysqlbinlogs",
	)
	if err != nil {
		return err
	}

	config.BinlogDir = binlogDir
	return nil
}

// GatherEncryptionSettings mengumpulkan pengaturan enkripsi - Task 2: modular function
func GatherEncryptionSettings(config *mariadb_utils.MariaDBConfigureConfig, collector *InputCollector) error {
	appDefaults := collector.Defaults.GetAppConfigDefaults()

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
			appDefaults.EncryptionKeyFile,
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

// GatherBufferPoolSettings mengumpulkan pengaturan buffer pool - Task 2: modular function
func GatherBufferPoolSettings(config *mariadb_utils.MariaDBConfigureConfig, collector *InputCollector) error {
	// Buffer pool size
	bufferPoolSize, err := collector.CollectString(
		"InnoDB buffer pool size (e.g., 1G, 512M)",
		config.InnodbBufferPoolSize,
		"innodb_buffer_pool_size",
		"", // no app config for this
		"128M",
		ValidateMemorySize,
	)
	if err != nil {
		return err
	}
	config.InnodbBufferPoolSize = bufferPoolSize

	// Buffer pool instances
	bufferPoolInstances, err := collector.CollectInt(
		"InnoDB buffer pool instances",
		config.InnodbBufferPoolInstances,
		"innodb_buffer_pool_instances",
		0, // no app config for this
		8,
		ValidateBufferPoolInstances,
	)
	if err != nil {
		return err
	}
	config.InnodbBufferPoolInstances = bufferPoolInstances

	return nil
}
