package remove

import (
	"fmt"
	"os"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// removeMariaDBRepository menghapus repository MariaDB jika diminta
func removeMariaDBRepository(cfg *mariadb_config.MariaDBRemoveConfig, deps *Dependencies) error {
	if !cfg.RemoveRepository {
		return nil
	}

	lg, _ := logger.Get()
	terminal.SafePrintln("🗑️  Menghapus repository MariaDB...")

	osInfo, err := system.DetectOS()
	if err != nil {
		return fmt.Errorf("gagal deteksi OS: %w", err)
	}

	switch osInfo.PackageType {
	case "deb":
		if err := removeDebianRepository(deps); err != nil {
			return fmt.Errorf("gagal menghapus repository Debian: %w", err)
		}
	case "rpm":
		if err := removeRPMRepository(deps); err != nil {
			return fmt.Errorf("gagal menghapus repository RPM: %w", err)
		}
	default:
		terminal.SafePrintln("   ℹ Repository removal tidak didukung untuk OS ini")
		return nil
	}

	lg.Info("Repository MariaDB berhasil dihapus")
	return nil
}

// removeDebianRepository menghapus repository MariaDB untuk Debian/Ubuntu
func removeDebianRepository(deps *Dependencies) error {
	repoFiles := []string{
		"/etc/apt/sources.list.d/mariadb.list",
		"/etc/apt/sources.list.d/MariaDB.list",
		"/etc/apt/trusted.gpg.d/mariadb.gpg",
		"/etc/apt/trusted.gpg.d/MariaDB.gpg",
	}

	for _, file := range repoFiles {
		if _, err := os.Stat(file); err != nil {
			if os.IsNotExist(err) {
				continue // File tidak ada, skip
			}
			return fmt.Errorf("tidak dapat mengakses %s: %w", file, err)
		}

		terminal.SafePrintln("   📄 Menghapus: " + file)
		if err := os.Remove(file); err != nil {
			return fmt.Errorf("gagal menghapus %s: %w", file, err)
		}
		terminal.SafePrintln("   ✓ Dihapus: " + file)
	}

	// Update package cache setelah menghapus repository
	terminal.SafePrintln("   🔄 Mengupdate package cache...")
	if err := deps.PackageManager.UpdateCache(); err != nil {
		// Log warning tapi tidak return error
		terminal.SafePrintln("   ⚠️  Gagal update package cache")
	} else {
		terminal.SafePrintln("   ✓ Package cache diupdate")
	}

	return nil
}

// removeRPMRepository menghapus repository MariaDB untuk CentOS/RHEL/Rocky
func removeRPMRepository(deps *Dependencies) error {
	repoFiles := []string{
		"/etc/yum.repos.d/MariaDB.repo",
		"/etc/yum.repos.d/mariadb.repo",
	}

	for _, file := range repoFiles {
		if _, err := os.Stat(file); err != nil {
			if os.IsNotExist(err) {
				continue // File tidak ada, skip
			}
			return fmt.Errorf("tidak dapat mengakses %s: %w", file, err)
		}

		terminal.SafePrintln("   📄 Menghapus: " + file)
		if err := os.Remove(file); err != nil {
			return fmt.Errorf("gagal menghapus %s: %w", file, err)
		}
		terminal.SafePrintln("   ✓ Dihapus: " + file)
	}

	// Clean package cache
	terminal.SafePrintln("   🔄 Membersihkan package cache...")
	if err := deps.ProcessManager.Execute("yum", []string{"clean", "all"}); err != nil {
		// Try dnf if yum fails
		if err := deps.ProcessManager.Execute("dnf", []string{"clean", "all"}); err != nil {
			// Log warning tapi tidak return error
			terminal.SafePrintln("   ⚠️  Gagal membersihkan package cache")
		} else {
			terminal.SafePrintln("   ✓ Package cache dibersihkan")
		}
	} else {
		terminal.SafePrintln("   ✓ Package cache dibersihkan")
	}

	return nil
}
