package mariadb

import (
	"fmt"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
)

// detectMariaDBService mendeteksi service MariaDB
func detectMariaDBService(installation *MariaDBInstallation) error {
	lg, _ := logger.Get()
	sm := system.NewServiceManager()
	pm := system.NewProcessManager()

	serviceNames := []string{"mariadb", "mysql", "mysqld"}

	// Prefer checking via ServiceManager (systemctl)
	for _, serviceName := range serviceNames {
		if sm.IsActive(serviceName) {
			installation.ServiceName = serviceName
			installation.IsRunning = true
			lg.Info("Ditemukan service", logger.String("service", serviceName), logger.Bool("is_running", installation.IsRunning))
			return nil
		}
	}

	// Fallback 1: try pgrep exact-match for common daemon names (less likely to false-positive)
	exactNames := []string{"mysqld", "mariadbd", "mysql"}
	for _, name := range exactNames {
		if out, err := pm.ExecuteWithOutput("pgrep", []string{"-x", name}); err == nil && strings.TrimSpace(out) != "" {
			installation.ServiceName = name
			installation.IsRunning = true
			lg.Info("Ditemukan process service (pgrep)", logger.String("service", name))
			return nil
		}
	}

	return fmt.Errorf("service MariaDB tidak ditemukan")
}
