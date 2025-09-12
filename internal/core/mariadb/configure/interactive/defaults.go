package interactive

import (
	"path/filepath"
	"strconv"

	"sfDBTools/internal/config/model"
)

// ConfigDefaults menyediakan fallback values untuk konfigurasi interaktif
// Task 1: Menggunakan config.yaml sebagai fallback jika template kosong
type ConfigDefaults struct {
	AppConfig *model.Config
	Template  *MariaDBConfigTemplate
}

// GetStringDefault mendapatkan string default dengan prioritas:
// 1. Template value (dari MariaDB config saat ini)
// 2. App config value (dari config.yaml)
// 3. Hardcoded default
func (cd *ConfigDefaults) GetStringDefault(templateKey string, appConfigValue string, hardcodedDefault string) string {
	// Priority 1: Template value (konfigurasi MariaDB saat ini)
	if cd.Template != nil && cd.Template.DefaultValues[templateKey] != "" {
		return cd.Template.DefaultValues[templateKey]
	}

	// Priority 2: App config value (dari config.yaml)
	if appConfigValue != "" {
		return appConfigValue
	}

	// Priority 3: Hardcoded default
	return hardcodedDefault
}

// GetIntDefault mendapatkan int default dengan prioritas yang sama
func (cd *ConfigDefaults) GetIntDefault(templateKey string, appConfigValue int, hardcodedDefault int) int {
	// Priority 1: Template value
	if cd.Template != nil && cd.Template.DefaultValues[templateKey] != "" {
		if val, err := strconv.Atoi(cd.Template.DefaultValues[templateKey]); err == nil {
			return val
		}
	}

	// Priority 2: App config value
	if appConfigValue > 0 {
		return appConfigValue
	}

	// Priority 3: Hardcoded default
	return hardcodedDefault
}

// GetDirectoryFromTemplate mencoba ekstrak directory dari template value
// Digunakan untuk log_error -> log directory, log_bin -> binlog directory
func (cd *ConfigDefaults) GetDirectoryFromTemplate(templateKey string) string {
	if cd.Template != nil && cd.Template.DefaultValues[templateKey] != "" {
		dir := filepath.Dir(cd.Template.DefaultValues[templateKey])
		if dir != "." && dir != "" {
			return dir
		}
	}
	return ""
}

// GetAppConfigDefaults mengambil default values dari config.yaml
func (cd *ConfigDefaults) GetAppConfigDefaults() AppConfigDefaults {
	defaults := AppConfigDefaults{}

	if cd.AppConfig != nil {
		defaults.ServerID = 1 // Server ID tidak ada di config.yaml, gunakan default
		defaults.Port = cd.AppConfig.MariaDB.Port
		defaults.DataDir = cd.AppConfig.MariaDB.DataDir
		defaults.LogDir = cd.AppConfig.MariaDB.LogDir
		defaults.BinlogDir = cd.AppConfig.MariaDB.BinlogDir
		defaults.EncryptionKeyFile = cd.AppConfig.ConfigDir.MariaDBKey
	}

	return defaults
}

// AppConfigDefaults berisi default values dari config.yaml
type AppConfigDefaults struct {
	ServerID          int
	Port              int
	DataDir           string
	LogDir            string
	BinlogDir         string
	EncryptionKeyFile string
}
