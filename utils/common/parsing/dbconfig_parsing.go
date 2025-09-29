package parsing

import (
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
