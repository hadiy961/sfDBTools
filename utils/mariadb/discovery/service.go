package mariadb

import (
	"fmt"
	"os/exec"
	"strings"

	"sfDBTools/internal/logger"
)

// detectMariaDBService mendeteksi service MariaDB
func detectMariaDBService(installation *MariaDBInstallation) error {
	lg, _ := logger.Get()
	serviceNames := []string{"mariadb", "mysql", "mysqld"}
	for _, serviceName := range serviceNames {
		cmd := exec.Command("systemctl", "is-active", serviceName)
		output, err := cmd.Output()
		if err == nil {
			status := strings.TrimSpace(string(output))
			installation.ServiceName = serviceName
			installation.IsRunning = (status == "active")
			lg.Info("Ditemukan service", logger.String("service", serviceName), logger.Bool("is_running", installation.IsRunning))
			return nil
		}
	}
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err == nil {
		outputStr := string(output)
		for _, serviceName := range serviceNames {
			if strings.Contains(outputStr, serviceName) {
				installation.ServiceName = serviceName
				installation.IsRunning = true
				lg.Info("Ditemukan process service", logger.String("service", serviceName))
				return nil
			}
		}
	}
	return fmt.Errorf("service MariaDB tidak ditemukan")
}
