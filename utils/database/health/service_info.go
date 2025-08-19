package health

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"sfDBTools/internal/logger"
)

// ServiceInfo represents MariaDB/MySQL service information
type ServiceInfo struct {
	ServiceName string `json:"service_name"`
	Status      string `json:"status"`
	IsActive    bool   `json:"is_active"`
	Uptime      string `json:"uptime"`
	ProcessID   string `json:"process_id"`
}

// GetServiceInfo retrieves MariaDB/MySQL service information from systemd
func GetServiceInfo() (*ServiceInfo, error) {
	lg, _ := logger.Get()

	info := &ServiceInfo{
		ServiceName: "mariadb",
	}

	// Get service status
	status, isActive, err := getServiceStatus()
	if err != nil {
		lg.Warn("Failed to get service status", logger.Error(err))
		info.Status = "Unknown"
		info.IsActive = false
	} else {
		info.Status = status
		info.IsActive = isActive
	}

	// Get service uptime
	uptime, err := getServiceUptime()
	if err != nil {
		lg.Warn("Failed to get service uptime", logger.Error(err))
		info.Uptime = "Unknown"
	} else {
		info.Uptime = uptime
	}

	// Get process ID
	pid, err := getServicePID()
	if err != nil {
		lg.Warn("Failed to get service PID", logger.Error(err))
		info.ProcessID = "Unknown"
	} else {
		info.ProcessID = pid
	}

	return info, nil
}

// getServiceStatus checks if MariaDB service is active
func getServiceStatus() (string, bool, error) {
	// Try mariadb first, then mysql as fallback
	services := []string{"mariadb", "mysql"}

	for _, service := range services {
		cmd := exec.Command("systemctl", "is-active", service)
		output, err := cmd.Output()
		if err == nil {
			status := strings.TrimSpace(string(output))
			isActive := status == "active"
			if isActive {
				return "✅ Active (running)", true, nil
			} else {
				return fmt.Sprintf("❌ %s", status), false, nil
			}
		}
	}

	return "Unknown", false, fmt.Errorf("unable to determine service status")
}

// getServiceUptime gets the service uptime
func getServiceUptime() (string, error) {
	services := []string{"mariadb", "mysql"}

	for _, service := range services {
		cmd := exec.Command("systemctl", "show", service, "--property=ActiveEnterTimestamp")
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		line := strings.TrimSpace(string(output))
		parts := strings.Split(line, "=")
		if len(parts) != 2 || parts[1] == "" {
			continue
		}

		// Parse the timestamp
		timestamp, err := time.Parse("Mon 2006-01-02 15:04:05 MST", parts[1])
		if err != nil {
			continue
		}

		duration := time.Since(timestamp)
		return formatUptime(duration), nil
	}

	return "", fmt.Errorf("unable to get service uptime")
}

// getServicePID gets the main process ID of the service
func getServicePID() (string, error) {
	services := []string{"mariadb", "mysql"}

	for _, service := range services {
		cmd := exec.Command("systemctl", "show", service, "--property=MainPID")
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		line := strings.TrimSpace(string(output))
		parts := strings.Split(line, "=")
		if len(parts) != 2 || parts[1] == "" || parts[1] == "0" {
			continue
		}

		return parts[1], nil
	}

	return "", fmt.Errorf("unable to get service PID")
}

// formatUptime formats duration into a human-readable string
func formatUptime(duration time.Duration) string {
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	} else {
		return fmt.Sprintf("%d minutes", minutes)
	}
}
