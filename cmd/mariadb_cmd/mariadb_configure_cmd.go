package mariadb_cmd

import (
	"context"

	"sfDBTools/internal/core/mariadb/configure"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/terminal"

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
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeMariaDBConfigure(cmd, args); err != nil {
			terminal.PrintError("Instalasi MariaDB gagal")
			terminal.WaitForEnterWithMessage("Tekan Enter untuk melanjutkan...")
			// Jangan panggil os.Exit di sini; biarkan Cobra menangani exit code
		} else {
			terminal.PrintSuccess("Instalasi MariaDB selesai")
			terminal.WaitForEnterWithMessage("Tekan Enter untuk melanjutkan...")
			return
		}
	},
}

func executeMariaDBConfigure(cmd *cobra.Command, args []string) error {
	// 1. Resolve config dari flags/env/file
	config, err := mariadb_config.ResolveMariaDBConfigureConfig(cmd)
	if err != nil {
		return err
	}

	// 2. Panggil core business logic
	return configure.RunMariaDBConfigure(context.Background(), config)
}

func init() {
	// Add flags untuk MariaDB configure
	mariadb_config.AddMariaDBConfigureFlags(ConfigureMariadbCMD)
}
