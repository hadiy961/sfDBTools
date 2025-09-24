package menu

import (
	"sfDBTools/cmd/mariadb_cmd"
	"sfDBTools/internal/config/model"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

func MariaDBMenu(lg *logger.Logger, cfg *model.Config) {
	terminal.Headers("Menu Instalasi MariaDB")
	choice, err := terminal.ShowMenuAndClear("Pilih Menu : ", []string{
		"Install MariaDB",
		"Hapus MariaDB",
		"Modifikasi Konfigurasi MariaDB",
		"Check Status MariaDB",
		"Check Versi (Online)",
		"Menu utama",
		"Keluar",
	})
	if err != nil {
		lg.Error("Menu error", logger.Error(err))
		return
	}

	switch choice {
	case 1:
		// Panggil fungsi atau command untuk Install MariaDB
		mariadb_cmd.InstallCmd.Run(mariadb_cmd.InstallCmd, []string{})
		return
	case 2:
		// Panggil fungsi atau command untuk Hapus MariaDB
		mariadb_cmd.RemoveCmd.Run(mariadb_cmd.RemoveCmd, []string{})
	case 3:
		lg.Info("Selected: Modifikasi Konfigurasi MariaDB")
		// Panggil fungsi atau command untuk Modifikasi Konfigurasi MariaDB
		mariadb_cmd.ConfigureMariadbCMD.Run(mariadb_cmd.ConfigureMariadbCMD, []string{})
	case 4:
		lg.Info("Selected: Check Status MariaDB")
		// Panggil fungsi atau command untuk Check Status MariaDB

	case 5:
		lg.Info("Selected: Check Versi (Online)")
		// Panggil fungsi atau command untuk Check Versi (Online)
	case 6:
		MenuUtama(lg, cfg)
		return
	case 7:
		lg.Info("Exiting application")
		return
	default:
		lg.Warn("Invalid selection")
	}
}
