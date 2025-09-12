package mariadb_cmd

import (
	"context"

	"sfDBTools/internal/core/mariadb/configure"
	mariadb_utils "sfDBTools/utils/mariadb"

	"github.com/spf13/cobra"
)

// ConfigureMariadbCMD configures MariaDB server with custom settings and data migration
var ConfigureMariadbCMD = &cobra.Command{
	Use:   "configure",
	Short: "Configure MariaDB server with custom settings and data migration",
	Long: `Configure MariaDB server with custom settings including:
- Data directory location (with automatic data migration)
- Log directory location  
- Binary log directory location
- InnoDB settings and encryption
- Auto-tuning based on system resources
- Port and network configuration

This command will safely migrate existing data if directories are changed.`,
	RunE: executeMariaDBConfigure,
}

func executeMariaDBConfigure(cmd *cobra.Command, args []string) error {
	// 1. Resolve config dari flags/env/file
	config, err := mariadb_utils.ResolveMariaDBConfigureConfig(cmd)
	if err != nil {
		return err
	}

	// 2. Panggil core business logic
	return configure.RunMariaDBConfigure(context.Background(), config)
}

func init() {
	// Add flags untuk MariaDB configure
	mariadb_utils.AddMariaDBConfigureFlags(ConfigureMariadbCMD)
}
