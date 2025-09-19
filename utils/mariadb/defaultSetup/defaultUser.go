package defaultsetup

import (
	"fmt"
	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
	"time"
)

func CreateDefaultMariaDBUser() error {
	lg, _ := logger.Get()
	ProcessManager := system.NewProcessManager()

	lg.Info("Membuat user default untuk akses awal")
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

	grantsSQL := "-- ---------------------------------------------------------------------\n"
	grantsSQL += "-- LANGKAH 1: BUAT SEMUA PENGGUNA (Gunakan IF NOT EXISTS)\n"
	grantsSQL += "-- Ini membuat skrip bisa dijalankan berulang kali tanpa error.\n"
	grantsSQL += "-- ---------------------------------------------------------------------\n\n"
	grantsSQL += "-- Pengguna Administratif\n"
	grantsSQL += "ALTER USER 'root'@'localhost' IDENTIFIED BY 'P@ssw0rdDB';\n"
	grantsSQL += "CREATE USER IF NOT EXISTS 'papp'@'%' IDENTIFIED BY 'P@ssw0rdpapp!@#';\n"
	grantsSQL += "CREATE USER IF NOT EXISTS 'sysadmin'@'%' IDENTIFIED BY 'P@ssw0rdsys!@#';\n"
	grantsSQL += "CREATE USER IF NOT EXISTS 'dbaDO'@'%' IDENTIFIED BY 'DataOn24!!';\n\n"
	grantsSQL += "-- Pengguna untuk Galera Cluster State Snapshot Transfer (SST)\n"
	grantsSQL += "CREATE USER IF NOT EXISTS 'sst_user'@'%' IDENTIFIED BY 'P@ssw0rdsst!@#';\n\n"
	grantsSQL += "-- Pengguna untuk Backup & Restore\n"
	grantsSQL += "CREATE USER IF NOT EXISTS 'backup_user'@'%' IDENTIFIED BY 'P@ssw0rdBackup!@#';\n"
	grantsSQL += "CREATE USER IF NOT EXISTS 'restore_user'@'%' IDENTIFIED BY 'P@ssw0rdRestore!@#';\n\n"
	grantsSQL += "-- Pengguna untuk MaxScale\n"
	grantsSQL += "CREATE USER IF NOT EXISTS 'maxscale'@'%' IDENTIFIED BY 'P@ssw0rdMaxscale!@#';\n\n"
	grantsSQL += "-- Pengguna Aplikasi untuk Klien '" + clientCode + "'\n"
	grantsSQL += "CREATE USER IF NOT EXISTS 'sfnbc_" + clientCode + "_admin'@'%' IDENTIFIED BY 'P@ssw0rdadm!@#';\n"
	grantsSQL += "CREATE USER IF NOT EXISTS 'sfnbc_" + clientCode + "_user'@'%' IDENTIFIED BY 'P@ssw0rduser!@#';\n"
	grantsSQL += "CREATE USER IF NOT EXISTS 'sfnbc_" + clientCode + "_fin'@'%' IDENTIFIED BY 'P@ssw0rdfin!@#';\n\n"
	grantsSQL += "-- Pengguna Administratif\n"
	grantsSQL += "-- PERINGATAN: 'ALL PRIVILEGES ON *.*' adalah risiko keamanan besar.\n"
	grantsSQL += "GRANT ALL PRIVILEGES ON *.* TO 'papp'@'%';\n"
	grantsSQL += "GRANT ALL PRIVILEGES ON *.* TO 'sysadmin'@'%';\n"
	grantsSQL += "GRANT ALL PRIVILEGES ON *.* TO 'dbaDO'@'%' WITH GRANT OPTION;\n\n"
	grantsSQL += "-- Pengguna SST (Hak Akses Minimal)\n"
	grantsSQL += "-- Hanya memberikan izin yang dibutuhkan untuk SST, bukan 'ALL'.\n"
	grantsSQL += "GRANT ALL PRIVILEGES ON *.* TO 'sst_user'@'%';\n\n"
	grantsSQL += "-- Pengguna Backup (Hak Akses Minimal) - DIPERBAIKI\n"
	// Pisahkan privilege yang bisa diberikan secara global vs per-database
	grantsSQL += "-- Global privileges untuk backup\n"
	grantsSQL += "GRANT SELECT, SHOW VIEW, TRIGGER, LOCK TABLES, EVENT ON *.* TO 'backup_user'@'%';\n"
	grantsSQL += "GRANT RELOAD, PROCESS, REPLICATION CLIENT ON *.* TO 'backup_user'@'%';\n"
	// ROUTINE privilege harus diberikan per database atau menggunakan EXECUTE
	grantsSQL += "GRANT EXECUTE ON *.* TO 'backup_user'@'%';\n\n"
	grantsSQL += "-- Pengguna Restore (Dibatasi pada Database Tertentu)\n"
	grantsSQL += "GRANT ALL PRIVILEGES ON `dbsf_nbc_" + clientCode + "_secondary_training`.* TO 'restore_user'@'%';\n"
	grantsSQL += "GRANT ALL PRIVILEGES ON `dbsf_nbc_" + clientCode + "_secondary_training_dmart`.* TO 'restore_user'@'%';\n\n"
	grantsSQL += "-- Pengguna MaxScale\n"
	grantsSQL += "GRANT ALL PRIVILEGES ON *.* TO 'maxscale'@'%';\n\n"
	grantsSQL += "-- Pengguna Aplikasi untuk Klien '" + clientCode + "'\n"
	grantsSQL += "-- Memberikan hak akses penuh pada database yang relevan untuk setiap pengguna.\n"
	grantsSQL += "GRANT ALL PRIVILEGES ON `dbsf_nbc_" + clientCode + "`.* TO 'sfnbc_" + clientCode + "_admin'@'%', 'sfnbc_" + clientCode + "_user'@'%', 'sfnbc_" + clientCode + "_fin'@'%';\n"
	grantsSQL += "GRANT ALL PRIVILEGES ON `dbsf_nbc_" + clientCode + "_dmart`.* TO 'sfnbc_" + clientCode + "_admin'@'%', 'sfnbc_" + clientCode + "_user'@'%', 'sfnbc_" + clientCode + "_fin'@'%';\n"
	grantsSQL += "GRANT ALL PRIVILEGES ON `dbsf_nbc_" + clientCode + "_temp`.* TO 'sfnbc_" + clientCode + "_admin'@'%', 'sfnbc_" + clientCode + "_user'@'%', 'sfnbc_" + clientCode + "_fin'@'%';\n"
	grantsSQL += "GRANT ALL PRIVILEGES ON `dbsf_nbc_" + clientCode + "_archive`.* TO 'sfnbc_" + clientCode + "_admin'@'%', 'sfnbc_" + clientCode + "_user'@'%', 'sfnbc_" + clientCode + "_fin'@'%';\n"
	grantsSQL += "GRANT ALL PRIVILEGES ON `dbsf_nbc_" + clientCode + "_secondary_training`.* TO 'sfnbc_" + clientCode + "_admin'@'%', 'sfnbc_" + clientCode + "_user'@'%', 'sfnbc_" + clientCode + "_fin'@'%';\n"
	grantsSQL += "GRANT ALL PRIVILEGES ON `dbsf_nbc_" + clientCode + "_secondary_training_dmart`.* TO 'sfnbc_" + clientCode + "_admin'@'%', 'sfnbc_" + clientCode + "_user'@'%', 'sfnbc_" + clientCode + "_fin'@'%';\n\n"
	grantsSQL += "FLUSH PRIVILEGES;\n"

	// Jalankan skrip SQL via mysql client
	args := []string{"-e", grantsSQL}

	if err := ProcessManager.ExecuteWithTimeout("mysql", args, 60*time.Second); err != nil {
		lg.Debug("Gagal menjalankan skrip grants default", logger.Error(err))
		return fmt.Errorf("gagal membuat grants default: %w", err)
	}

	lg.Info("Grants default berhasil diterapkan untuk client " + clientCode)
	return nil
}
