package validation

import (
	"fmt"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
)

func validatePort(port int) error {
	lg, _ := logger.Get()
	lg.Debug("Validating port", logger.Int("port", port))

	if port < 1024 || port > 65535 {
		return fmt.Errorf("port must be between 1024-65535, got: %d", port)
	}

	pi, err := system.CheckPortConflict(port)
	if err != nil {
		if isPortInUse(port) {
			return fmt.Errorf("port %d is already in use", port)
		}
		return nil
	}

	if pi.Status == "available" {
		return nil
	}

	proc := strings.ToLower(pi.ProcessName)
	if proc == "mysqld" || proc == "mariadbd" || strings.Contains(proc, "mysqld") || strings.Contains(proc, "mariadb") {
		lg.Info("Port digunakan oleh MariaDB, mengabaikan konflik", logger.Int("port", port), logger.String("process", pi.ProcessName))
		return nil
	}

	return fmt.Errorf("port %d is already in use by process %s (pid=%s)", port, pi.ProcessName, pi.PID)
}

func isPortInUse(port int) bool {
	return !system.IsPortAvailable(port)
}
