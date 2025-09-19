package defaultsetup

import (
	"fmt"
	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
	"time"
)

// membuat database default untuk client_code
func CreateDefaultDatabase() error {
	lg, _ := logger.Get()
	ProcessManager := system.NewProcessManager()
	lg.Info("Membuat database default untuk client_code")
	// Ambil client_code dari konfigurasi aplikasi
	conf, confErr := config.Get()
	clientCode := "demo"
	if confErr == nil && conf != nil {
		if conf.General.ClientCode != "" {
			clientCode = conf.General.ClientCode
		}
	} else {
		lg.Debug("Gagal membaca konfigurasi, menggunakan client_code default 'demo'", logger.Error(confErr))
	}

	databaseSQL := "-- ---------------------------------------------------------------------\n"
	databaseSQL += "-- LANGKAH 2: BUAT DATABASE UNTUK KLIEN '" + clientCode + "' (Gunakan IF NOT EXISTS)\n"
	databaseSQL += "-- Ini membuat skrip bisa dijalankan berulang kali tanpa error.\n"
	databaseSQL += "-- ---------------------------------------------------------------------\n\n"
	databaseSQL += "CREATE DATABASE IF NOT EXISTS `dbsf_nbc_" + clientCode + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;\n"
	databaseSQL += "CREATE DATABASE IF NOT EXISTS `dbsf_nbc_" + clientCode + "_dmart` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;\n"
	databaseSQL += "CREATE DATABASE IF NOT EXISTS `dbsf_nbc_" + clientCode + "_temp` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;\n"
	databaseSQL += "CREATE DATABASE IF NOT EXISTS `dbsf_nbc_" + clientCode + "_archive` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;\n"
	databaseSQL += "CREATE DATABASE IF NOT EXISTS `dbsf_nbc_" + clientCode + "_secondary_training` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;\n"
	databaseSQL += "CREATE DATABASE IF NOT EXISTS `dbsf_nbc_" + clientCode + "_secondary_training_dmart` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;\n"
	databaseSQL += "DROP DATABASE test;\n"
	databaseSQL += "DELETE FROM mysql.user WHERE user = '';\n"
	databaseSQL += "FLUSH PRIVILEGES;\n"

	// Jalankan skrip SQL via mysql client
	args := []string{"-e", databaseSQL}

	if err := ProcessManager.ExecuteWithTimeout("mysql", args, 60*time.Second); err != nil {
		lg.Debug("Gagal menjalankan skrip pembuatan database default", logger.Error(err))
		return fmt.Errorf("gagal membuat database default: %w", err)
	}

	lg.Info("Database default berhasil dibuat untuk client " + clientCode)
	return nil
}
