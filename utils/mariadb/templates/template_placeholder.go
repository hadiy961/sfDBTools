package templates

import (
	"strings"
)

func ParsePlaceholders(template *MariaDBConfigTemplate) error {
	lines := strings.Split(template.Content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				if strings.HasPrefix(value, "{{") && strings.HasSuffix(value, "}}") {
					placeholder := strings.Trim(value, "{}")
					template.Placeholders[key] = placeholder
				} else {
					template.DefaultValues[key] = value
				}
			}
		}
	}
	return nil
}

func SetDefaultValues(template *MariaDBConfigTemplate) {
	defaults := map[string]string{
		"server_id": "1",
		"file_key_management_encryption_algorithm": "AES_CTR",
		"file_key_management_filename":             "/var/lib/mysql/encryption/keyfile",
		"innodb-encrypt-tables":                    "ON",
		"log_bin":                                  "/var/lib/mysqlbinlogs/mysql-bin",
		"datadir":                                  "/var/lib/mysql",
		"socket":                                   "/var/lib/mysql/mysql.sock",
		"port":                                     "3306",
		"innodb_buffer_pool_size":                  "128M",
		"innodb_data_home_dir":                     "/var/lib/mysql",
		"innodb_log_group_home_dir":                "/var/lib/mysql",
		"log_error":                                "/var/lib/mysql/mysql_error.log",
		"slow_query_log_file":                      "/var/lib/mysql/mysql_slow.log",
		"innodb_buffer_pool_instances":             "8",
	}
	for key, value := range defaults {
		if _, exists := template.DefaultValues[key]; !exists {
			template.DefaultValues[key] = value
		}
	}
}
