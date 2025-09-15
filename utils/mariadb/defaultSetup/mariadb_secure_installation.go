package defaultsetup

import (
	"fmt"
	"os/exec"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
	"time"
)

// menjalankan mariadb-secure-installation
func RunMariaDBSecureInstallation() error {
	lg, _ := logger.Get()
	ProcessManager := system.NewProcessManager()
	args := []string{}

	// Periksa apakah executable `mariadb-secure-installation` tersedia di PATH
	if _, lookErr := exec.LookPath("mariadb-secure-installation"); lookErr != nil {
		// Pesan teknis di log (hindari pengulangan teks UI)
		lg.Debug("exec.LookPath untuk mariadb-secure-installation gagal", logger.Error(lookErr))
		return fmt.Errorf("mariadb-secure-installation tidak ditemukan di $PATH: %w", lookErr)
	}

	// Jalankan secara interaktif sehingga user dapat merespon prompt
	if err := ProcessManager.ExecuteWithTimeout("mariadb-secure-installation", args, 2*time.Minute); err != nil {
		// Pesan teknis di log (hindari pengulangan teks UI)
		lg.Debug("mariadb-secure-installation execution failed", logger.Error(err))
		return fmt.Errorf("gagal menjalankan secure installation: %w", err)
	}

	lg.Info("Secure installation berhasil dijalankan")

	lg.Info("Post-installation setup selesai")
	return nil
}
