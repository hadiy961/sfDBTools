package remove

import (
	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// stopMariaDBService menghentikan dan mendisable service MariaDB
func stopMariaDBService(deps *Dependencies) error {
	lg, _ := logger.Get()

	terminal.PrintSubHeader("🛑 Menghentikan & mendisable service MariaDB...")

	serviceName := "mariadb"

	// Cek apakah service aktif
	if deps.ServiceManager.IsActive(serviceName) {
		// Stop service
		if err := deps.ServiceManager.Stop(serviceName); err != nil {
			lg.Warn("Gagal menghentikan service MariaDB", logger.Error(err))
			warn("Gagal menghentikan service MariaDB, melanjutkan cleanup")
		} else {
			success("Service MariaDB dihentikan")
		}
	} else {
		info("Service MariaDB sudah tidak aktif")
	}

	// Disable service agar tidak auto-start
	if err := deps.ServiceManager.Disable(serviceName); err != nil {
		lg.Warn("Gagal mendisable service MariaDB", logger.Error(err))
		warn("Gagal mendisable service MariaDB")
	} else {
		success("Service MariaDB didisable")
	}

	// Kill proses yang mungkin masih berjalan
	if err := killMariaDBProcesses(deps); err != nil {
		lg.Warn("Gagal kill proses MariaDB", logger.Error(err))
		warn("Gagal menghentikan beberapa proses MariaDB")
	}

	return nil
}

// killMariaDBProcesses membunuh proses MariaDB yang mungkin masih berjalan
func killMariaDBProcesses(deps *Dependencies) error {

	terminal.PrintSubHeader("🔫 Menghentikan proses MariaDB yang masih berjalan...")

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

	success("Proses MariaDB dihentikan")
	return nil
}
