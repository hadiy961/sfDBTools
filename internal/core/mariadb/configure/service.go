package configure

import (
	"fmt"
	"os/exec"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// ServiceManager handles MariaDB service operations
type ServiceManager struct {
	serviceName string
}

// NewServiceManager creates a new service manager
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		serviceName: "mariadb",
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

// StartService starts the MariaDB service
func (s *ServiceManager) StartService() error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Starting MariaDB service...")

	cmd := exec.Command("systemctl", "start", s.serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to start MariaDB service",
			logger.Error(err),
			logger.String("output", string(output)))
		return fmt.Errorf("failed to start MariaDB service: %w", err)
	}

	lg.Info("MariaDB service started successfully")
	terminal.PrintSuccess("MariaDB service started")
	return nil
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
