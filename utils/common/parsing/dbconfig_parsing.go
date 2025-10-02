package parsing

import (
	"fmt"
	defaultconfig "sfDBTools/internal/config/default_config"
	"sfDBTools/utils/common"
	"sfDBTools/utils/common/structs"

	"github.com/spf13/cobra"
)

func ParseDBConfigFlags(cmd *cobra.Command) (DBConfig *structs.DBConfig, err error) {

	// Initialize DBConfig so we don't dereference a nil pointer
	DBConfig = &structs.DBConfig{}

	// Gunakan helper yang sudah ada di utils/common
	DBConfig.FileInfo.Path = common.GetStringFlagOrEnv(cmd, "file", "SFDB_CONFIG_FILE", "")
	DBConfig.ForceDelete = common.GetBoolFlagOrEnv(cmd, "force", "SFDB_FORCE_DELETE", false)
	DBConfig.DeleteAll = common.GetBoolFlagOrEnv(cmd, "all", "SFDB_DELETE_ALL", false)
	DBConfig.ConfigName = common.GetStringFlagOrEnv(cmd, "name", "SFDB_CONFIG_NAME", "")
	DBConfig.ConnectionOptions.Host = common.GetStringFlagOrEnv(cmd, "host", "SFDB_DB_HOST", "localhost")
	DBConfig.ConnectionOptions.Port = common.GetIntFlagOrEnv(cmd, "port", "SFDB_DB_PORT", 3306)
	DBConfig.ConnectionOptions.User = common.GetStringFlagOrEnv(cmd, "user", "SFDB_DB_USER", "")
	DBConfig.EncryptionConfig.EncryptionPassword = common.GetStringFlagOrEnv(cmd, "encryption-password", "SFDB_ENCRYPTION_PASSWORD", "")
	DBConfig.ConnectionOptions.Password = common.GetStringFlagOrEnv(cmd, "password", "SFDB_DB_PASSWORD", "")
	return DBConfig, nil
}

func ParseDBConfigGenerate(cmd *cobra.Command) (GenerateDefault *structs.DBConfigGenerateOptions, err error) {
	// 1. Dapatkan nilai default dari Configuration.
	// Nilai ini akan menjadi fallback terakhir.
	GenerateDefault, err = defaultconfig.GetDBConfigGenerateDefaults()
	if err != nil {
		// Kembalikan error agar caller (seperti fungsi Run Cobra) yang menanganinya.
		return nil, fmt.Errorf("failed to load general backup defaults from config: %w", err)
	}

	// 2. Parse flags dinamis ke dalam struct menggunakan refleksi.
	if err := DynamicParseFlags(cmd, GenerateDefault); err != nil {
		return nil, fmt.Errorf("failed to dynamically parse general backup flags: %w", err)
	}

	return GenerateDefault, nil
}
