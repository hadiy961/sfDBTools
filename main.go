package main

import (
	"fmt"
	"os"

	"sfDBTools/cmd"
	"sfDBTools/internal/config"
)

func main() {
	// Validasi config terlebih dahulu sebelum menjalankan command apapun
	if _, err := config.LoadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Pastikan file konfigurasi ada di /etc/sfDBTools/config/config.yaml\n")
		os.Exit(1)
	}

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
