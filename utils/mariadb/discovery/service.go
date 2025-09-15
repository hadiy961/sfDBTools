package mariadb

import (
	"fmt"
	"regexp"
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

	// Fallback 2: parse `ps aux` but match only in the COMMAND column using word-boundary regex
	output, err := pm.ExecuteWithOutput("ps", []string{"aux"})
	if err == nil {
		lines := strings.Split(output, "\n")
		// regex to match whole word (word-boundary) of service name
		for _, serviceName := range serviceNames {
			re := regexp.MustCompile("\\b" + regexp.QuoteMeta(serviceName) + "\\b")
			for _, line := range lines {
				// skip empty lines
				if strings.TrimSpace(line) == "" {
					continue
				}
				// ps aux: the COMMAND column starts at the 11th field (index 10)
				fields := strings.Fields(line)
				if len(fields) < 11 {
					// fallback: check whole line
					if re.MatchString(line) {
						installation.ServiceName = serviceName
						installation.IsRunning = true
						lg.Debug("Matched ps line (fallback)", logger.String("line", line))
						lg.Info("Ditemukan process service", logger.String("service", serviceName))
						return nil
					}
					continue
				}
				cmd := strings.Join(fields[10:], " ")
				if re.MatchString(cmd) {
					installation.ServiceName = serviceName
					installation.IsRunning = true
					lg.Debug("Matched ps COMMAND", logger.String("command", cmd))
					lg.Info("Ditemukan process service", logger.String("service", serviceName))
					return nil
				}
			}
		}
	}

	return fmt.Errorf("service MariaDB tidak ditemukan")
}
