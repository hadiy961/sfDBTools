package remove

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/mariadb"
	"sfDBTools/utils/terminal"
)

// handleDataBackup melakukan backup data sebelum dihapus jika diminta
func handleDataBackup(cfg *mariadb.MariaDBRemoveConfig, deps *Dependencies) error {
	if !cfg.BackupData {
		return nil
	}

	lg, _ := logger.Get()
	terminal.SafePrintln("💾 Membuat backup data MariaDB...")

	// Deteksi direktori custom dari konfigurasi
	mariadbConfig, err := detectCustomDirectories()
	if err != nil {
		lg.Warn("Gagal deteksi direktori custom untuk backup, menggunakan default", logger.Error(err))
		// Fallback ke backup direktori default
		return backupDefaultDataDirectory(cfg, deps)
	}

	// Backup semua direktori custom yang terdeteksi
	return backupCustomDataDirectories(cfg, deps, mariadbConfig)
}

// backupDefaultDataDirectory backup direktori data default
func backupDefaultDataDirectory(cfg *mariadb.MariaDBRemoveConfig, deps *Dependencies) error {
	dataDir := "/var/lib/mysql"

	// Cek apakah data directory ada
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		terminal.SafePrintln("   ℹ Direktori data tidak ditemukan, skip backup")
		return nil
	}

	// Buat backup directory dengan timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupDir := filepath.Join(cfg.BackupPath, "mariadb_backup_"+timestamp)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori backup: %w", err)
	}

	terminal.SafePrintln("   📁 Direktori backup: " + backupDir)

	// Copy data directory
	if err := copyDirectory(deps, dataDir, filepath.Join(backupDir, "mysql")); err != nil {
		return fmt.Errorf("gagal backup data: %w", err)
	}

	terminal.SafePrintln("   ✓ Backup data berhasil")
	return nil
}

// backupCustomDataDirectories backup semua direktori custom yang terdeteksi
func backupCustomDataDirectories(cfg *mariadb.MariaDBRemoveConfig, deps *Dependencies, config *MariaDBConfig) error {
	lg, _ := logger.Get()

	// Buat backup directory dengan timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupDir := filepath.Join(cfg.BackupPath, "mariadb_backup_"+timestamp)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori backup: %w", err)
	}

	terminal.SafePrintln("   📁 Direktori backup: " + backupDir)

	// Backup direktori data utama
	if _, err := os.Stat(config.DataDir); err == nil {
		destDir := filepath.Join(backupDir, "data")
		if err := copyDirectory(deps, config.DataDir, destDir); err != nil {
			return fmt.Errorf("gagal backup data directory: %w", err)
		}
		terminal.SafePrintln("   ✓ Backup data directory: " + config.DataDir)
	}

	// Backup direktori InnoDB jika berbeda
	if config.InnoDBDir != "" && config.InnoDBDir != config.DataDir {
		if _, err := os.Stat(config.InnoDBDir); err == nil {
			destDir := filepath.Join(backupDir, "innodb")
			if err := copyDirectory(deps, config.InnoDBDir, destDir); err != nil {
				lg.Warn("Gagal backup InnoDB directory", logger.Error(err))
			} else {
				terminal.SafePrintln("   ✓ Backup InnoDB directory: " + config.InnoDBDir)
			}
		}
	}

	// Backup direktori binlog jika berbeda
	if config.BinlogDir != "" && config.BinlogDir != config.DataDir {
		if _, err := os.Stat(config.BinlogDir); err == nil {
			destDir := filepath.Join(backupDir, "binlogs")
			if err := copyDirectory(deps, config.BinlogDir, destDir); err != nil {
				lg.Warn("Gagal backup binlog directory", logger.Error(err))
			} else {
				terminal.SafePrintln("   ✓ Backup binlog directory: " + config.BinlogDir)
			}
		}
	}

	// Backup file log individual
	logFiles := getAllCustomFiles(config)
	if len(logFiles) > 0 {
		logBackupDir := filepath.Join(backupDir, "logs")
		if err := os.MkdirAll(logBackupDir, 0755); err == nil {
			for _, logFile := range logFiles {
				if _, err := os.Stat(logFile); err == nil {
					destFile := filepath.Join(logBackupDir, filepath.Base(logFile))
					if err := copyFile(logFile, destFile); err != nil {
						lg.Warn("Gagal backup log file", logger.String("file", logFile), logger.Error(err))
					} else {
						terminal.SafePrintln("   ✓ Backup log file: " + logFile)
					}
				}
			}
		}
	}

	terminal.SafePrintln("   ✓ Backup data berhasil")
	lg.Info("Backup data MariaDB berhasil", logger.String("backup_path", backupDir))

	return nil
}

// copyFile menyalin file individual
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// removeDataAndConfig menghapus data dan konfigurasi jika diminta
func removeDataAndConfig(cfg *mariadb.MariaDBRemoveConfig, deps *Dependencies) error {
	lg, _ := logger.Get()

	if cfg.RemoveData {
		// Deteksi direktori custom dari konfigurasi
		mariadbConfig, err := detectCustomDirectories()
		if err != nil {
			lg.Warn("Gagal deteksi direktori custom, menggunakan default", logger.Error(err))
			// Fallback ke penghapusan direktori default
			if err := removeDefaultDataDirectories(); err != nil {
				return fmt.Errorf("gagal menghapus data directory: %w", err)
			}
		} else {
			// Hapus direktori berdasarkan konfigurasi yang terdeteksi
			if err := removeCustomDataDirectories(mariadbConfig); err != nil {
				return fmt.Errorf("gagal menghapus data directory custom: %w", err)
			}
		}
		lg.Info("Data directory MariaDB berhasil dihapus")
	}

	if cfg.RemoveConfig {
		if err := removeConfigFiles(); err != nil {
			return fmt.Errorf("gagal menghapus file konfigurasi: %w", err)
		}
		lg.Info("File konfigurasi MariaDB berhasil dihapus")
	}

	return nil
}

// removeDefaultDataDirectories menghapus direktori data default MariaDB
func removeDefaultDataDirectories() error {
	terminal.SafePrintln("🗑️  Menghapus data directory MariaDB (default)...")

	dataDirs := []string{
		"/var/lib/mysql",
		"/var/lib/mysql-files",
		"/var/lib/mysql-keyring",
	}

	for _, dir := range dataDirs {
		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				continue // Directory tidak ada, skip
			}
			return fmt.Errorf("tidak dapat mengakses direktori %s: %w", dir, err)
		}

		terminal.SafePrintln("   🗂️  Menghapus: " + dir)
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("gagal menghapus direktori %s: %w", dir, err)
		}
		terminal.SafePrintln("   ✓ Dihapus: " + dir)
	}

	return nil
}

// removeCustomDataDirectories menghapus direktori berdasarkan konfigurasi yang terdeteksi
func removeCustomDataDirectories(config *MariaDBConfig) error {
	terminal.SafePrintln("🗑️  Menghapus data directory MariaDB (custom)...")

	// Dapatkan semua direktori yang perlu dihapus
	customDirs := getAllCustomDirectories(config)
	customFiles := getAllCustomFiles(config)

	// Hapus file-file individual terlebih dahulu
	for _, file := range customFiles {
		if _, err := os.Stat(file); err != nil {
			if os.IsNotExist(err) {
				continue // File tidak ada, skip
			}
			continue // Skip error, tidak critical untuk file individual
		}

		terminal.SafePrintln("   📄 Menghapus file: " + file)
		if err := os.Remove(file); err != nil {
			terminal.SafePrintln("   ⚠️  Gagal menghapus file: " + file)
		} else {
			terminal.SafePrintln("   ✓ Dihapus file: " + file)
		}
	}

	// Hapus direktori
	for _, dir := range customDirs {
		// Validasi keamanan direktori
		if err := validateDirectoryForRemoval(dir); err != nil {
			terminal.SafePrintln("   ⚠️  Skip direktori tidak aman: " + dir + " (" + err.Error() + ")")
			continue
		}

		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				continue // Directory tidak ada, skip
			}
			return fmt.Errorf("tidak dapat mengakses direktori %s: %w", dir, err)
		}

		terminal.SafePrintln("   🗂️  Menghapus direktori: " + dir)
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("gagal menghapus direktori %s: %w", dir, err)
		}
		terminal.SafePrintln("   ✓ Dihapus direktori: " + dir)
	}

	return nil
}

// removeConfigFiles menghapus file-file konfigurasi MariaDB
func removeConfigFiles() error {
	terminal.SafePrintln("🗑️  Menghapus file konfigurasi MariaDB...")

	configPaths := []string{
		"/etc/mysql",
		"/etc/my.cnf",
		"/etc/mysql/my.cnf",
		"/etc/my.cnf.d",
		"/etc/mariadb",
		"/root/.my.cnf",
		"/home/*/.my.cnf", // Note: ini memerlukan expansion shell
	}

	for _, path := range configPaths {
		if path == "/home/*/.my.cnf" {
			// Handle wildcard path secara manual
			if err := removeUserConfigFiles(); err != nil {
				// Log warning tapi tidak return error
				terminal.SafePrintln("   ⚠️  Gagal menghapus beberapa file config user")
			}
			continue
		}

		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				continue // File/directory tidak ada, skip
			}
			return fmt.Errorf("tidak dapat mengakses %s: %w", path, err)
		}

		terminal.SafePrintln("   📄 Menghapus: " + path)
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("gagal menghapus %s: %w", path, err)
		}
		terminal.SafePrintln("   ✓ Dihapus: " + path)
	}

	return nil
}

// removeUserConfigFiles menghapus file konfigurasi user .my.cnf
func removeUserConfigFiles() error {
	homePattern := "/home/*/.my.cnf"

	// Cari semua file .my.cnf di home directory
	matches, err := filepath.Glob(homePattern)
	if err != nil {
		return err
	}

	for _, match := range matches {
		if err := os.Remove(match); err != nil {
			// Log tapi tidak return error
			terminal.SafePrintln("   ⚠️  Gagal menghapus: " + match)
		} else {
			terminal.SafePrintln("   ✓ Dihapus: " + match)
		}
	}

	return nil
}

// copyDirectory melakukan copy rekursif directory
func copyDirectory(deps *Dependencies, src, dst string) error {
	// Gunakan rsync untuk copy yang efisien
	args := []string{"-av", src + "/", dst + "/"}

	if err := deps.ProcessManager.Execute("rsync", args); err != nil {
		// Fallback ke cp jika rsync tidak tersedia
		cpArgs := []string{"-r", src, dst}
		if err := deps.ProcessManager.Execute("cp", cpArgs); err != nil {
			return fmt.Errorf("gagal copy directory: %w", err)
		}
	}

	return nil
}
