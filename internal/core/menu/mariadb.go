package menu

import (
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
		lg.Info("Selected: Install MariaDB")
		// Panggil fungsi atau command untuk Install MariaDB
	case 2:
		lg.Info("Selected: Hapus MariaDB")
		// Panggil fungsi atau command untuk Hapus MariaDB
	case 3:
		lg.Info("Selected: Modifikasi Konfigurasi MariaDB")
		// Panggil fungsi atau command untuk Modifikasi Konfigurasi MariaDB
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
