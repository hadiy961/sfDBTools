package remove

import (
	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// stopMariaDBService menghentikan dan mendisable service MariaDB
func stopMariaDBService(deps *Dependencies) error {
	lg, _ := logger.Get()

	terminal.SafePrintln("🛑 Menghentikan service MariaDB...")

	serviceName := "mariadb"

	// Cek apakah service aktif
	if deps.ServiceManager.IsActive(serviceName) {
		// Stop service
		if err := deps.ServiceManager.Stop(serviceName); err != nil {
			lg.Warn("Gagal menghentikan service MariaDB", logger.Error(err))
			// Tidak return error karena mungkin service sudah mati
		} else {
			terminal.SafePrintln("   ✓ Service MariaDB dihentikan")
		}
	} else {
		terminal.SafePrintln("   ℹ Service MariaDB sudah tidak aktif")
	}

	// Disable service agar tidak auto-start
	if err := deps.ServiceManager.Disable(serviceName); err != nil {
		lg.Warn("Gagal mendisable service MariaDB", logger.Error(err))
		// Tidak return error karena mungkin service sudah disabled
	} else {
		terminal.SafePrintln("   ✓ Service MariaDB didisable")
	}

	// Kill proses yang mungkin masih berjalan
	if err := killMariaDBProcesses(deps); err != nil {
		lg.Warn("Gagal kill proses MariaDB", logger.Error(err))
		// Tidak return error karena proses mungkin sudah mati
	}

	lg.Info("Service MariaDB berhasil dihentikan dan didisable")
	return nil
}

// killMariaDBProcesses membunuh proses MariaDB yang mungkin masih berjalan
func killMariaDBProcesses(deps *Dependencies) error {
	terminal.SafePrintln("   🔫 Menghentikan proses MariaDB yang masih berjalan...")

	// Cari proses mysqld
	processes := []string{"mysqld", "mariadbd", "mysql"}

	for _, process := range processes {
		// Gunakan pkill untuk menghentikan proses
		err := deps.ProcessManager.Execute("pkill", []string{"-f", process})
		if err != nil {
			// Proses mungkin tidak ada, tidak masalah
			continue
		}
	}

	terminal.SafePrintln("   ✓ Proses MariaDB dihentikan")
	return nil
}
