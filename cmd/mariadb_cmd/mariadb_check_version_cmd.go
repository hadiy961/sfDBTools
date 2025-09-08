package mariadb_cmd

import (
	"github.com/spf13/cobra"
)

// CheckVersionCmd command for checking available MariaDB versions
var CheckVersionCmd = &cobra.Command{
	Use:   "check_version",
	Short: "Cek versi MariaDB yang tersedia",
	Long: `Menampilkan daftar versi MariaDB yang tersedia untuk instalasi.
Informasi diambil dari dokumentasi resmi MariaDB Repository Setup Script.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// return mariadb.DisplayAvailableVersions()
		return nil
	},
}

func init() {
	// No flags needed since we only have one display format
}
