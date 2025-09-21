package mariadb_cmd

import (
	"github.com/spf13/cobra"
)

// Check command for checking installed MariaDB version
var Check = &cobra.Command{
	Use:   "check",
	Short: "Cek versi MariaDB yang terpasang",
	Long: `Menampilkan versi MariaDB yang terpasang saat ini.
Informasi diambil dari sistem yang sedang berjalan.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// return mariadb.DisplayInstalledVersion()
		return nil
	},
}
