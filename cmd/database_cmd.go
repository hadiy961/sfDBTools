package cmd

import (
	database_cmd "sfDBTools/cmd/database_cmd"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

// DatabaseCmd root untuk operasi manajemen database (non-backup)
var DatabaseCmd = &cobra.Command{
	Use:   "database",
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
	rootCmd.AddCommand(DatabaseCmd)
	DatabaseCmd.AddCommand(database_cmd.DatabaseDropCmd)
}
