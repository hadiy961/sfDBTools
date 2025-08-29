package cmd

import (
	mariadb_cmd "sfDBTools/cmd/mariadb_cmd"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

// MariaDBCmd root untuk operasi manajemen database (non-backup)
var MariaDBCmd = &cobra.Command{
	Use:   "mariadb",
	Short: "Perintah manajemen database (drop, dsb)",
	Long:  "Kumpulan subcommand untuk operasi administrasi database yang bersifat destruktif atau manajerial.",
	Run: func(cmd *cobra.Command, args []string) {
		lg, _ := logger.Get()
		lg.Info("Menjalankan perintah database (menampilkan help)")
		cmd.Help()
	},
	Annotations: map[string]string{
		"command":  "database",
		"category": "administration",
	},
}

func init() {
	rootCmd.AddCommand(MariaDBCmd)
	MariaDBCmd.AddCommand(mariadb_cmd.CheckVersionCmd)
	MariaDBCmd.AddCommand(mariadb_cmd.ConfigureMariadbCMD)
	MariaDBCmd.AddCommand(mariadb_cmd.InstallCmd)
	MariaDBCmd.AddCommand(mariadb_cmd.RemoveCmd)
}
