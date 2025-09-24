package menu

import (
	"sfDBTools/cmd/dbconfig_cmd"
	"sfDBTools/internal/config/model"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

func DBConfigMenu(lg *logger.Logger, cfg *model.Config) {
	terminal.Headers("Manajemen Konfigurasi DB")
	choice, err := terminal.ShowMenuAndClear("Pilih Menu : ", []string{
		"Buat Konfigurasi DB",
		"Edit Konfigurasi DB",
		"Hapus Konfigurasi DB",
		"Validasi Koneksi DB",
		"Lihat Konfigurasi DB",
		"Menu utama",
		"Keluar",
	})
	if err != nil {
		lg.Error("Menu error", logger.Error(err))
		return
	}

	switch choice {
	case 1:
		lg.Info("Selected: Buat Konfigurasi DB")
		dbconfig_cmd.GenerateCmd.Run(dbconfig_cmd.GenerateCmd, []string{})
		return
		// Panggil fungsi atau command untuk Buat Konfigurasi DB
	case 2:
		lg.Info("Selected: Edit Konfigurasi DB")
		dbconfig_cmd.EditCmd.Run(dbconfig_cmd.EditCmd, []string{})
		return
		// Panggil fungsi atau command untuk Edit Konfigurasi DB
	case 3:
		lg.Info("Selected: Hapus Konfigurasi DB")
		// Panggil fungsi atau command untuk Backup Database
	case 4:
		lg.Info("Selected: Restore Database")
		// Panggil fungsi atau command untuk Restore Database
	case 5:
		lg.Info("Exiting application")
		return
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
