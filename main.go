package main

import (
	"fmt"
	"os"

	"sfDBTools/cmd"
	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
)

func main() {
	// Validasi config terlebih dahulu sebelum menjalankan command apapun
	// if _, err := config.LoadConfig(); err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	// 	// fmt.Fprintf(os.Stderr, "Pastikan file konfigurasi ada di /etc/sfDBTools/config/config.yaml\n")
	// 	os.Exit(1)
	// }

	cfg, err := config.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	lg, err := logger.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Logger initialization error: %v\n", err)
		os.Exit(1)
	}
	lg.Info("Starting "+cfg.General.AppName, logger.String("version", cfg.General.Version))

	if err := cmd.Execute(cfg, lg); err != nil {
		os.Exit(1)
	}
}
