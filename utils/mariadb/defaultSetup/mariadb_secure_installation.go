package defaultsetup

import (
	"fmt"
	"os/exec"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
	"time"
)

type Dependencies struct {
	PackageManager system.PackageManager
	ProcessManager system.ProcessManager
	ServiceManager system.ServiceManager
}

// menjalankan mariadb-secure-installation
func RunMariaDBSecureInstallation(deps *Dependencies) error {
	lg, _ := logger.Get()
	args := []string{}

	// Beberapa distro/versi menyediakan tool dengan nama berbeda. Prioritaskan
	// `mysql_secure_installation` (umum pada MariaDB/MySQL) lalu fallback ke
	// `mariadb-secure-installation`.
	candidates := []string{"mysql_secure_installation", "mariadb-secure-installation"}
	var bin string
	for _, c := range candidates {
		if _, err := exec.LookPath(c); err == nil {
			bin = c
			break
		} else {
			// log the specific error for debugging
			lg.Debug("exec.LookPath gagal untuk kandidat", logger.String("candidate", c), logger.Error(err))
		}
	}

	if bin == "" {
		return fmt.Errorf("tidak ditemukan executable secure-installation (cari %v) di $PATH", candidates)
	}

	// Jalankan secara interaktif sehingga user dapat merespon prompt. Gunakan
	// ExecuteInteractiveWithTimeout agar stdin/stdout terhubung ke terminal.
	timeout := 5 * time.Minute
	if err := deps.ProcessManager.ExecuteInteractiveWithTimeout(bin, args, timeout); err != nil {
		lg.Debug("secure-installation execution failed", logger.String("binary", bin), logger.Error(err))
		return fmt.Errorf("gagal menjalankan %s: %w", bin, err)
	}

	lg.Info("Secure installation berhasil dijalankan")

	lg.Info("Post-installation setup selesai")
	return nil
}
