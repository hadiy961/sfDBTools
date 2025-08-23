package templates

import (
	"fmt"
	"sfDBTools/internal/config"
)

// MariaDBServerTemplate contains the MariaDB server configuration template
// This template is based on production-ready settings with placeholders for customization
const MariaDBServerTemplate = `[server]
log_warnings                                    = 1
server_id                                       = {{.ServerID}}
gtid-domain-id                                  = 1
gtid_ignore_duplicates                          = ON
gtid_strict_mode                                = 1
rpl_semi_sync_master_enabled                    = ON
rpl_semi_sync_slave_enabled                     = ON
rpl_semi_sync_master_wait_point                 = AFTER_SYNC
slave-skip-errors                               = 1062,1032

[mysqld]
thread_handling                                 = pool-of-threads
plugin-load-add                                 = file_key_management
file_key_management_encryption_algorithm        = AES_CTR
file_key_management_filename                    = {{.KeyFilePath}}
innodb-encrypt-tables
log_bin                                         = {{.BinLogPath}}
datadir                                         = {{.DataDir}}
lower_case_table_names                          = 1
sql-mode                                        = "PIPES_AS_CONCAT"
skip-host-cache
skip-name-resolve
log-slave-updates                               = 1
query_cache_size                                = 0 #query cache is not supported with wsrep
query_cache_type                                = 0 #query cache is not supported with wsrep

# LIMIT #
net_buffer_length                               = 16384
max_allowed_packet                              = 1G
expire_logs_days                                = 3
max_connections                                 = {{.MaxConnections}}
max_connect_errors                              = 1000
wait_timeout                                    = 40
interactive_timeout                             = 40
max_statement_time                              = 900
open-files-limit                                = 393210

# INNODB #
default_storage_engine                          = InnoDB
innodb_data_home_dir                            = {{.DataDir}}
innodb_log_group_home_dir                       = {{.DataDir}}
innodb_file_per_table                           = 1
innodb_log_file_size                            = 2G
innodb_autoinc_lock_mode                        = 2
innodb_flush_log_at_trx_commit                  = 1
innodb_doublewrite                              = 1
binlog_format                                   = ROW
log_bin_trust_function_creators                 = 1

log_error                                       = {{.DataDir}}/mysql_error.log
slow_query_log                                  = 1
slow_query_log_file                             = {{.DataDir}}/mysql_slow.log
long_query_time                                 = 2
log_slow_verbosity                              = query_plan,explain

port                                            = {{.Port}}  #custom - EDIT TERLEBIH DAHULU
bind-address                                    = {{.BindAddress}}

performance_schema                              = ON

innodb_buffer_pool_size                         = {{.BufferPoolSize}}
innodb_buffer_pool_instances                    = 8
innodb_buffer_pool_chunk_size                   = 128M
thread_cache_size                               = 256
join_buffer_size                                = 1M
key_buffer_size                                 = 128M
max_heap_table_size                             = 512M
tmp_table_size                                  = 512M
table_open_cache                                = 2000
table_definition_cache                          = 400
innodb_flush_method                             = O_DIRECT
`

// MariaDBConfigParams holds the template parameters
type MariaDBConfigParams struct {
	ServerID       string
	KeyFilePath    string
	BinLogPath     string
	DataDir        string
	MaxConnections string
	Port           string
	BindAddress    string
	BufferPoolSize string
}

// GetDefaultParams returns default configuration parameters from config file
func GetDefaultParams() (MariaDBConfigParams, error) {
	// Load configuration from config.yaml
	cfg, err := config.LoadConfig()
	if err != nil {
		// Return default values if config loading fails
		return MariaDBConfigParams{
			ServerID:       "TES-1",
			KeyFilePath:    "/var/lib/mysql/key_maria_nbc.txt",
			BinLogPath:     "/var/lib/mysqlbinlogs/mysql-bin",
			DataDir:        "/var/lib/mysql",
			MaxConnections: "10000",
			Port:           "3306",
			BindAddress:    "0.0.0.0",
			BufferPoolSize: "10G",
		}, err
	}

	// Extract parameters from configuration
	params := MariaDBConfigParams{
		ServerID:       cfg.General.ClientCode + "-1", // Use client_code from config
		KeyFilePath:    cfg.ConfigDir.MariaDBKey,      // Default path, could be made configurable
		BinLogPath:     cfg.MariaDB.Installation.BinlogDir + "/mysql-bin",
		DataDir:        cfg.MariaDB.Installation.DataDir,
		MaxConnections: "10000", // Default value, could be made configurable
		Port:           fmt.Sprintf("%d", cfg.MariaDB.Installation.Port),
		BindAddress:    "0.0.0.0", // Default value, could be made configurable
		BufferPoolSize: "10G",     // Default value, could be made configurable
	}

	return params, nil
}
