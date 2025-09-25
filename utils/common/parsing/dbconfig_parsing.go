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
	DBConfig.FileInfo.Path = common.GetStringFlagOrEnv(cmd, "file", "SFDBTOOLS_CONFIG_FILE", "")
	DBConfig.ForceDelete = common.GetBoolFlagOrEnv(cmd, "force", "SFDBTOOLS_FORCE_DELETE", false)
	DBConfig.DeleteAll = common.GetBoolFlagOrEnv(cmd, "all", "SFDBTOOLS_DELETE_ALL", false)
	DBConfig.ConfigName = common.GetStringFlagOrEnv(cmd, "name", "SFDBTOOLS_CONFIG_NAME", "")
	DBConfig.ConnectionOptions.Host = common.GetStringFlagOrEnv(cmd, "host", "SFDBTOOLS_DB_HOST", "localhost")
	DBConfig.ConnectionOptions.Port = common.GetIntFlagOrEnv(cmd, "port", "SFDBTOOLS_DB_PORT", 3306)
	DBConfig.ConnectionOptions.User = common.GetStringFlagOrEnv(cmd, "user", "SFDBTOOLS_DB_USER", "")
	DBConfig.AutoMode = common.GetBoolFlagOrEnv(cmd, "auto", "SFDBTOOLS_AUTO_MODE", false)

	return DBConfig, nil
}
