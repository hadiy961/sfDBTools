package menu

import (
	"fmt"
	"sfDBTools/internal/config/model"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

func MenuUtama(lg *logger.Logger, cfg *model.Config) {
	terminal.Headers("Menu Utama")
	choice, err := terminal.ShowMenuAndClear("Pilih Menu : ", []string{
		"Menu Konfigurasi DB",
		"Menu Instalasi MariDB",
		"Menu Backup",
		"Menu Restore",
		"Menu Backup & Restore",
		"Keluar",
	})
	if err != nil {
		lg.Error("Menu error", logger.Error(err))
		return
	}

	switch choice {
	case 1:
		DBConfigMenu(lg, cfg)
		return
	case 2:
		MariaDBMenu(lg, cfg)
		return
	case 3:
		lg.Info("Selected: Menu Backup")
		fmt.Println("Fungsi Backup Database belum diimplementasikan.")
		terminal.WaitForEnterWithMessage("Tekan Enter untuk kembali ke menu utama...")
		MenuUtama(lg, cfg)
		return
	case 4:
		lg.Info("Selected: Menu Restore")
		fmt.Println("Fungsi Restore Database belum diimplementasikan.")
		terminal.WaitForEnterWithMessage("Tekan Enter untuk kembali ke menu utama...")
		MenuUtama(lg, cfg)
		return
	case 5:
		lg.Info("Selected: Menu Backup & Restore")
		fmt.Println("Fungsi Backup & Restore Database belum diimplementasikan.")
		terminal.WaitForEnterWithMessage("Tekan Enter untuk kembali ke menu utama...")
		MenuUtama(lg, cfg)
		return
	case 6:
		lg.Info("Exiting application")
		return
	default:
		lg.Warn("Invalid selection")
	}
}
