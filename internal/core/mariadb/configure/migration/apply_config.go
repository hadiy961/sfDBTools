package migration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"sfDBTools/internal/core/mariadb/configure/template"
	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
)

func ApplyConfiguration(ctx context.Context, config *mariadb_config.MariaDBConfigureConfig, tpl *template.MariaDBConfigTemplate) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Applying MariaDB configuration")

	backupPath, err := tpl.BackupCurrentConfig(config.BackupDir)
	if err != nil {
		return fmt.Errorf("failed to backup current config: %w", err)
	}
	lg.Info("Current configuration backed up", logger.String("backup_path", backupPath))

	configValues := buildConfigValues(config)
	newConfig, err := tpl.GenerateConfigFromTemplate(configValues)
	if err != nil {
		return fmt.Errorf("failed to generate config from template: %w", err)
	}

	if err := writeConfiguration(tpl.CurrentPath, newConfig); err != nil {
		return fmt.Errorf("failed to write new configuration: %w", err)
	}

	lg.Info("MariaDB configuration applied successfully", logger.String("config_path", tpl.CurrentPath))
	return nil
}

func buildConfigValues(config *mariadb_config.MariaDBConfigureConfig) map[string]string {
	values := make(map[string]string)

	values["server_id"] = fmt.Sprintf("%d", config.ServerID)
	values["port"] = fmt.Sprintf("%d", config.Port)
	values["datadir"] = config.DataDir
	// values["socket"] = chooseSocketPathImpl(config.DataDir, disk.GetUsage)
	values["log_bin"] = filepath.Join(config.BinlogDir, "mysql-bin")
	values["log_error"] = filepath.Join(config.LogDir, "mysql_error.log")
	values["slow_query_log_file"] = filepath.Join(config.LogDir, "mysql_slow.log")
	values["innodb_data_home_dir"] = config.DataDir
	values["innodb_log_group_home_dir"] = config.DataDir
	values["innodb_buffer_pool_size"] = config.InnodbBufferPoolSize
	values["innodb_buffer_pool_instances"] = fmt.Sprintf("%d", config.InnodbBufferPoolInstances)

	if config.InnodbEncryptTables {
		values["innodb_encrypt_tables"] = "ON"
		values["file_key_management_encryption_key_file"] = config.EncryptionKeyFile
		values["file_key_management_encryption_algorithm"] = "AES_CTR"
	} else {
		values["innodb_encrypt_tables"] = "OFF"
	}

	return values
}

// writeConfiguration writes the provided content to the given path
func writeConfiguration(configPath, content string) error {
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}
