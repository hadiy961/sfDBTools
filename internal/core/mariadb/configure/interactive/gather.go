package interactive

import (
	"path/filepath"
	mariadb_config "sfDBTools/utils/mariadb/config"
)

// GatherServerID mengumpulkan server ID - Task 2: modular function
func GatherServerID(config *mariadb_config.MariaDBConfigureConfig, collector *InputCollector) error {

	serverID, err := collector.CollectInt(
		"Server ID for replication",
		collector.Defaults.GetIntDefault("server_id", 1),
		"server_id",
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
		collector.Defaults.GetIntDefault("port", 3306),
		"port",
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
		collector.Defaults.GetStringDefault("datadir", "/var/lib/mysql"),
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
		filepath.Dir(collector.Defaults.GetStringDefault("log_error", "/var/lib/mysql/")),
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
		filepath.Dir(collector.Defaults.GetStringDefault("log_bin", "/var/lib/mysql/mysql-bin")),
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

	// Gather encryption enabled/disabled using prioritized defaults
	defaultEncrypt := collector.Defaults.GetBoolDefault("innodb_encrypt_tables", config.InnodbEncryptTables)
	// If current config already set, prefer that
	if config.InnodbEncryptTables {
		defaultEncrypt = config.InnodbEncryptTables
	}

	config.InnodbEncryptTables = collector.CollectBool("Enable table encryption?", defaultEncrypt)

	// If encryption enabled, gather key file using prioritized defaults
	if config.InnodbEncryptTables {
		keyFileDefault := collector.Defaults.GetStringDefault("file_key_management_filename", "/var/lib/mysql/encryption/keyfile")
		// If the config already has a value prefer it
		if config.EncryptionKeyFile != "" {
			keyFileDefault = config.EncryptionKeyFile
		}

		keyFile, err := collector.CollectString(
			"Encryption key file path",
			"",
			"file_key_management_filename",
			keyFileDefault,
			ValidateAbsolutePath,
		)
		if err != nil {
			return err
		}

		config.EncryptionKeyFile = keyFile
	}

	return nil
}
