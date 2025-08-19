package mariadb

import (
	"fmt"
	"os/exec"
	"strings"

	"sfDBTools/internal/logger"
)

// GetServiceInfo gets information about MariaDB service
func GetServiceInfo() (*ServiceInfo, error) {
	lg, _ := logger.Get()

	// Try mariadb service first, then mysql
	services := []string{"mariadb", "mysql"}

	for _, service := range services {
		info, err := getServiceStatus(service)
		if err == nil {
			lg.Debug("Service found",
				logger.String("service", service),
				logger.String("status", info.Status))
			return info, nil
		}
	}

	// No service found
	return &ServiceInfo{
		Name:    "",
		Status:  "not-found",
		Enabled: false,
		Active:  false,
		Running: false,
	}, nil
}

// getServiceStatus gets status of a specific service
func getServiceStatus(serviceName string) (*ServiceInfo, error) {
	info := &ServiceInfo{Name: serviceName}

	// First check if service unit exists
	cmd := exec.Command("systemctl", "list-unit-files", serviceName+".service")
	output, err := cmd.Output()
	if err != nil || !strings.Contains(string(output), serviceName+".service") {
		// Service unit file doesn't exist
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// Check service status
	cmd = exec.Command("systemctl", "is-active", serviceName)
	output, err = cmd.Output()
	if err != nil {
		// Service exists but might be inactive/failed
		info.Status = "inactive"
		info.Active = false
		info.Running = false
	} else {
		status := strings.TrimSpace(string(output))
		info.Status = status
		info.Active = (status == "active")
		info.Running = (status == "active")
	}

	// Check if service is enabled
	cmd = exec.Command("systemctl", "is-enabled", serviceName)
	output, err = cmd.Output()
	if err == nil {
		enabled := strings.TrimSpace(string(output))
		info.Enabled = (enabled == "enabled")
	}

	return info, nil
}

// StopService stops MariaDB service
func StopService() error {
	lg, _ := logger.Get()

	services := []string{"mariadb", "mysql"}
	stopped := false

	for _, service := range services {
		lg.Debug("Attempting to stop service", logger.String("service", service))

		cmd := exec.Command("systemctl", "stop", service)
		err := cmd.Run()
		if err == nil {
			lg.Info("Service stopped", logger.String("service", service))
			stopped = true
		} else {
			lg.Debug("Failed to stop service",
				logger.String("service", service),
				logger.Error(err))
		}
	}

	if !stopped {
		lg.Warn("No MariaDB services were stopped (might not be running)")
	}

	return nil
}

// DisableService disables MariaDB service from auto-start
func DisableService() error {
	lg, _ := logger.Get()

	services := []string{"mariadb", "mysql"}
	disabled := false

	for _, service := range services {
		lg.Debug("Attempting to disable service", logger.String("service", service))

		cmd := exec.Command("systemctl", "disable", service)
		err := cmd.Run()
		if err == nil {
			lg.Info("Service disabled", logger.String("service", service))
			disabled = true
		} else {
			lg.Debug("Failed to disable service",
				logger.String("service", service),
				logger.Error(err))
		}
	}

	if !disabled {
		lg.Warn("No MariaDB services were disabled")
	}

	return nil
}

// MaskAndRemoveServices masks and attempts to remove MariaDB systemd service files
func MaskAndRemoveServices() error {
	lg, _ := logger.Get()

	services := []string{"mariadb", "mysql"}

	// First mask the services to prevent them from being started
	for _, service := range services {
		lg.Debug("Attempting to mask service", logger.String("service", service))

		cmd := exec.Command("systemctl", "mask", service)
		err := cmd.Run()
		if err == nil {
			lg.Info("Service masked", logger.String("service", service))
		} else {
			lg.Debug("Failed to mask service",
				logger.String("service", service),
				logger.Error(err))
		}
	}

	// Reload systemd to ensure changes take effect
	lg.Debug("Reloading systemd daemon")
	cmd := exec.Command("systemctl", "daemon-reload")
	if err := cmd.Run(); err != nil {
		lg.Debug("Failed to reload systemd daemon", logger.Error(err))
	} else {
		lg.Debug("Systemd daemon reloaded")
	}

	// Reset failed units
	lg.Debug("Resetting failed systemd units")
	for _, service := range services {
		cmd := exec.Command("systemctl", "reset-failed", service)
		if err := cmd.Run(); err != nil {
			lg.Debug("Failed to reset failed unit",
				logger.String("service", service),
				logger.Error(err))
		}
	}

	return nil
}

// IsServiceRunning checks if any MariaDB service is running
func IsServiceRunning() (bool, string) {
	services := []string{"mariadb", "mysql"}

	for _, service := range services {
		cmd := exec.Command("systemctl", "is-active", service)
		output, err := cmd.Output()
		if err == nil {
			status := strings.TrimSpace(string(output))
			if status == "active" {
				return true, service
			}
		}
	}

	return false, ""
}

// GetRunningProcesses gets running MariaDB/MySQL processes
func GetRunningProcesses() ([]string, error) {
	var processes []string

	// Check for mysqld processes
	cmd := exec.Command("pgrep", "-f", "mysqld")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		pids := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, pid := range pids {
			if pid != "" {
				processes = append(processes, fmt.Sprintf("mysqld (PID: %s)", pid))
			}
		}
	}

	return processes, nil
}
