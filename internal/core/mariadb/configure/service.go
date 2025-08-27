package configure

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// ServiceManager handles MariaDB service operations
type ServiceManager struct {
	serviceName string
	settings    *MariaDBSettings
}

// NewServiceManager creates a new service manager
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		serviceName: "mariadb",
		settings:    nil,
	}
}

// NewServiceManagerWithSettings creates a new service manager with MariaDB settings
func NewServiceManagerWithSettings(settings *MariaDBSettings) *ServiceManager {
	return &ServiceManager{
		serviceName: "mariadb",
		settings:    settings,
	}
}

// StopService stops the MariaDB service
func (s *ServiceManager) StopService() error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Stopping MariaDB service...")

	cmd := exec.Command("systemctl", "stop", s.serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to stop MariaDB service",
			logger.Error(err),
			logger.String("output", string(output)))
		return fmt.Errorf("failed to stop MariaDB service: %w", err)
	}

	lg.Info("MariaDB service stopped successfully")
	terminal.PrintSuccess("MariaDB service stopped")
	return nil
}

// StartService starts the MariaDB service with retry mechanism
func (s *ServiceManager) StartService() error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Starting MariaDB service...")

	// Try up to 3 times with increasing delay
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		cmd := exec.Command("systemctl", "start", s.serviceName)
		output, err := cmd.CombinedOutput()

		if err == nil {
			lg.Info("MariaDB service started successfully")
			terminal.PrintSuccess("MariaDB service started")
			return nil
		}

		lg.Warn("MariaDB service start attempt failed",
			logger.Int("attempt", attempt),
			logger.Int("max_retries", maxRetries),
			logger.Error(err),
			logger.String("output", string(output)))

		if attempt == maxRetries {
			// Final attempt failed, provide diagnostic information
			terminal.PrintError("MariaDB failed to start. Collecting diagnostic information...")
			s.provideDiagnosticInfo()

			lg.Error("Failed to start MariaDB service after retries",
				logger.Error(err),
				logger.String("output", string(output)))
			return fmt.Errorf("failed to start MariaDB service: %w", err)
		}

		// Wait before retry (2, 4, 6 seconds)
		time.Sleep(time.Duration(attempt*2) * time.Second)
		terminal.PrintWarning(fmt.Sprintf("Retrying MariaDB start (attempt %d/%d)...", attempt+1, maxRetries))
	}

	return fmt.Errorf("unexpected error in service start retry loop")
}

// EnableService enables the MariaDB service on boot
func (s *ServiceManager) EnableService() error {
	lg, _ := logger.Get()

	cmd := exec.Command("systemctl", "enable", s.serviceName)
	if err := cmd.Run(); err != nil {
		lg.Error("Failed to enable MariaDB service", logger.Error(err))
		return fmt.Errorf("failed to enable MariaDB service: %w", err)
	}

	lg.Info("MariaDB service enabled on boot")
	terminal.PrintSuccess("MariaDB service enabled on boot")
	return nil
}

// GetServiceStatus gets the current status of MariaDB service
func (s *ServiceManager) GetServiceStatus() error {
	lg, _ := logger.Get()

	cmd := exec.Command("systemctl", "status", s.serviceName)
	output, err := cmd.Output()
	if err != nil {
		lg.Error("Failed to get service status", logger.Error(err))
		return fmt.Errorf("failed to get service status: %w", err)
	}

	lg.Info("MariaDB service status", logger.String("status", string(output)))
	terminal.PrintInfo(fmt.Sprintf("MariaDB service status:\n%s", string(output)))
	return nil
}

// IsServiceRunning checks if MariaDB service is running
func (s *ServiceManager) IsServiceRunning() bool {
	cmd := exec.Command("systemctl", "is-active", s.serviceName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return string(output) == "active\n"
}

// provideDiagnosticInfo provides diagnostic information when service fails to start
func (s *ServiceManager) provideDiagnosticInfo() {
	lg, _ := logger.Get()

	// Show service status
	terminal.PrintWarning("Service Status:")
	if cmd := exec.Command("systemctl", "status", s.serviceName, "--no-pager", "-l"); cmd != nil {
		if output, err := cmd.CombinedOutput(); err == nil {
			terminal.PrintInfo(string(output))
		}
	}

	// Show recent journal logs
	terminal.PrintWarning("Recent Service Logs:")
	if cmd := exec.Command("journalctl", "-u", s.serviceName, "--no-pager", "-n", "20"); cmd != nil {
		if output, err := cmd.CombinedOutput(); err == nil {
			terminal.PrintInfo(string(output))
		}
	}

	// Check for MariaDB error logs using dynamic paths
	if s.settings != nil && s.settings.LogDir != "" {
		// Try error log in configured log directory
		errorLogPath := filepath.Join(s.settings.LogDir, "mysql_error.log")
		if cmd := exec.Command("test", "-f", errorLogPath); cmd.Run() == nil {
			terminal.PrintWarning(fmt.Sprintf("Recent errors from %s:", errorLogPath))
			if tailCmd := exec.Command("tail", "-20", errorLogPath); tailCmd != nil {
				if output, err := tailCmd.CombinedOutput(); err == nil && len(output) > 0 {
					terminal.PrintInfo(string(output))
				} else {
					terminal.PrintInfo("Error log file exists but is empty or unreadable")
				}
			}
		} else {
			terminal.PrintWarning(fmt.Sprintf("Error log not found at expected location: %s", errorLogPath))
		}
	}

	lg.Info("Diagnostic information provided for MariaDB startup failure")
}
