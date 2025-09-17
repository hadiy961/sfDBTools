package interactive

import (
	"fmt"
	"path/filepath"
	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/mariadb/discovery"
	"strconv"
)

// ConfigDefaults menyediakan fallback values untuk konfigurasi interaktif
// Task 1: Menggunakan config.yaml sebagai fallback jika template kosong
type ConfigDefaults struct {
	Installation *discovery.MariaDBInstallation
	Template     *MariaDBConfigTemplate
	AppConfig    *mariadb_config.MariaDBConfigureConfig // Optional, jika ingin gunakan config.yaml
}

// GetStringDefault mendapatkan string default dengan prioritas:
// 1. Template value (dari MariaDB config saat ini)
// 2. App config value (dari config.yaml)
// 3. Hardcoded default
func (cd *ConfigDefaults) GetStringDefault(templateKey string, hardcodedDefault string) string {
	// Priority 1: Current installation values
	if cd.Installation != nil {
		switch templateKey {
		case "datadir":
			if cd.Installation.DataDir != "" {
				return cd.Installation.DataDir
			}
		case "log_error":
			if cd.Installation.LogDir != "" {
				// return a plausible log filename inside the detected log dir
				return filepath.Join(cd.Installation.LogDir, "mysqld.log")
			}
		case "log_bin":
			if cd.Installation.BinlogDir != "" {
				return filepath.Join(cd.Installation.BinlogDir, "mysql-bin")
			}
		case "file_key_management_filename":
			if cd.Installation.EncryptionKeyFile != "" {
				return cd.Installation.EncryptionKeyFile
			}
		}
	}

	// Priority 2: Template value (dari MariaDB config saat ini)
	if cd.Template != nil && cd.Template.DefaultValues[templateKey] != "" {
		return cd.Template.DefaultValues[templateKey]
	}

	// Priority 3: App config (config.yaml)
	if cd.AppConfig != nil {
		switch templateKey {
		case "datadir":
			if cd.AppConfig.DataDir != "" {
				return cd.AppConfig.DataDir
			}
		case "log_bin":
			if cd.AppConfig.BinlogDir != "" {
				return filepath.Join(cd.AppConfig.BinlogDir, "mysql-bin")
			}
		case "file_key_management_filename":
			if cd.AppConfig.EncryptionKeyFile != "" {
				return cd.AppConfig.EncryptionKeyFile
			}
		}
	}

	// Priority 4: Hardcoded default
	return hardcodedDefault
}

// GetIntDefault mendapatkan int default dengan prioritas yang sama
func (cd *ConfigDefaults) GetIntDefault(templateKey string, hardcodedDefault int) int {
	// Priority 1: Current installation
	if cd.Installation != nil {
		switch templateKey {
		case "server_id":
			if cd.Installation.ServerID > 0 {
				return cd.Installation.ServerID
			}
		case "port":
			if cd.Installation.Port > 0 {
				return cd.Installation.Port
			}
		case "innodb-buffer-pool-instances":
			if cd.Installation.InnodbBufferPoolInstances > 0 {
				return cd.Installation.InnodbBufferPoolInstances
			}
		}
	}

	// Priority 2: AppConfig
	if cd.AppConfig != nil {
		switch templateKey {
		case "server_id":
			if cd.AppConfig.ServerID > 0 {
				return cd.AppConfig.ServerID
			}
		case "port":
			if cd.AppConfig.Port > 0 {
				return cd.AppConfig.Port
			}
		case "innodb-buffer-pool-instances":
			if cd.AppConfig.InnodbBufferPoolInstances > 0 {
				return cd.AppConfig.InnodbBufferPoolInstances
			}
		}
	}

	// Priority 3: Template value
	if cd.Template != nil && cd.Template.DefaultValues[templateKey] != "" {
		if val, err := strconv.Atoi(cd.Template.DefaultValues[templateKey]); err == nil {
			return val
		}
	}

	// Priority 4: Hardcoded default
	return hardcodedDefault
}

// GetDirectoryFromTemplate mencoba ekstrak directory dari template value
// Digunakan untuk log_error -> log directory, log_bin -> binlog directory
func (cd *ConfigDefaults) GetDirectoryFromTemplate(templateKey string) string {
	// Priority 1: Current installation directories
	if cd.Installation != nil {
		switch templateKey {
		case "log_error":
			if cd.Installation.LogDir != "" {
				return cd.Installation.LogDir
			}
		case "log_bin":
			if cd.Installation.BinlogDir != "" {
				return cd.Installation.BinlogDir
			}
		case "datadir":
			if cd.Installation.DataDir != "" {
				return cd.Installation.DataDir
			}
		}
	}

	// Priority 2: Template value (try extract dir)
	if cd.Template != nil && cd.Template.DefaultValues[templateKey] != "" {
		dir := filepath.Dir(cd.Template.DefaultValues[templateKey])
		if dir != "." && dir != "" {
			return dir
		}
	}

	// Priority 3: AppConfig values
	if cd.AppConfig != nil {
		switch templateKey {
		case "log_bin":
			if cd.AppConfig.BinlogDir != "" {
				return cd.AppConfig.BinlogDir
			}
		case "datadir":
			if cd.AppConfig.DataDir != "" {
				return cd.AppConfig.DataDir
			}
		}
	}

	return ""
}

// GetBoolDefault mendapatkan bool default dengan prioritas Current -> Template -> AppConfig -> hardcoded
func (cd *ConfigDefaults) GetBoolDefault(templateKey string, hardcodedDefault bool) bool {
	// Priority 1: Current
	if cd.Installation != nil {
		switch templateKey {
		case "innodb_encrypt_tables":
			return cd.Installation.InnodbEncryptTables
		}
	}

	// Priority 2: Template
	if cd.Template != nil {
		v := cd.Template.DefaultValues[templateKey]
		if v == "ON" || v == "1" || v == "true" || v == "True" {
			return true
		}
		if v == "OFF" || v == "0" || v == "false" || v == "False" {
			return false
		}
	}

	// Priority 3: AppConfig
	if cd.AppConfig != nil {
		switch templateKey {
		case "innodb_encrypt_tables":
			return cd.AppConfig.InnodbEncryptTables
		}
	}

	// Priority 4: hardcoded
	return hardcodedDefault
}

// InteractiveValue berisi default values dari config.yaml
type InteractiveValue struct {
	ServerID          int
	Port              int
	DataDir           string
	LogDir            string
	BinlogDir         string
	EncryptionKeyFile string
}

// Current returns a logger.Field constructed from the resolved default value
// The "def" parameter is used only to indicate the expected type and the
// hardcoded default; the actual returned value follows the priority rules
// implemented by the Get*Default methods.
func (cd *ConfigDefaults) Current(templateKey string, def interface{}) logger.Field {
	switch v := def.(type) {
	case int:
		val := cd.GetIntDefault(templateKey, v)
		return logger.Int(templateKey, val)
	case string:
		val := cd.GetStringDefault(templateKey, v)
		return logger.String(templateKey, val)
	case bool:
		val := cd.GetBoolDefault(templateKey, v)
		return logger.Bool(templateKey, val)
	default:
		// Fallback: try to represent as string
		var s string
		if v == nil {
			s = ""
		} else {
			s = fmt.Sprintf("%v", v)
		}
		val := cd.GetStringDefault(templateKey, s)
		return logger.String(templateKey, val)
	}
}
