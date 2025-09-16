package mariadb

import (
	"sfDBTools/utils/common"

	"github.com/spf13/cobra"
)

// ResolveMariaDBRemoveConfig membaca flags/env untuk konfigurasi penghapusan
func ResolveMariaDBRemoveConfig(cmd *cobra.Command) (*MariaDBRemoveConfig, error) {
	// Baca konfigurasi dari flags dan environment variables
	removeData := common.GetBoolFlagOrEnv(cmd, "remove-data", "SFDBTOOLS_REMOVE_DATA", false)
	removeConfig := common.GetBoolFlagOrEnv(cmd, "remove-config", "SFDBTOOLS_REMOVE_CONFIG", false)
	removeRepository := common.GetBoolFlagOrEnv(cmd, "remove-repository", "SFDBTOOLS_REMOVE_REPOSITORY", false)
	removeUser := common.GetBoolFlagOrEnv(cmd, "remove-user", "SFDBTOOLS_REMOVE_USER", false)
	force := common.GetBoolFlagOrEnv(cmd, "force", "SFDBTOOLS_FORCE", false)
	backupData := common.GetBoolFlagOrEnv(cmd, "backup-data", "SFDBTOOLS_BACKUP_DATA", false)
	backupPath := common.GetStringFlagOrEnv(cmd, "backup-path", "SFDBTOOLS_BACKUP_PATH", "/tmp/mariadb_backup")
	nonInteractive := common.GetBoolFlagOrEnv(cmd, "non-interactive", "SFDBTOOLS_NON_INTERACTIVE", false)

	cfg := &MariaDBRemoveConfig{
		RemoveData:       removeData,
		RemoveConfig:     removeConfig,
		RemoveRepository: removeRepository,
		RemoveUser:       removeUser,
		Force:            force,
		BackupData:       backupData,
		BackupPath:       backupPath,
		NonInteractive:   nonInteractive,
	}

	return cfg, nil
}
