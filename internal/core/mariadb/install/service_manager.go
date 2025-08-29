package install

import (
	"fmt"
	"os/exec"
)

// ServiceManager handles MariaDB service operations
type ServiceManager struct{}

// NewServiceManager creates a new service manager
func NewServiceManager() *ServiceManager {
	return &ServiceManager{}
}

// StartMariaDBService starts the MariaDB service
func (s *ServiceManager) StartMariaDBService() error {
	services := []string{"mariadb", "mysql"}

	for _, service := range services {
		cmd := exec.Command("systemctl", "start", service)
		if err := cmd.Run(); err == nil {
			return nil // Successfully started
		}
	}

	return fmt.Errorf("failed to start any MariaDB service")
}

// EnableMariaDBService enables MariaDB service on boot
func (s *ServiceManager) EnableMariaDBService() error {
	services := []string{"mariadb", "mysql"}

	for _, service := range services {
		cmd := exec.Command("systemctl", "enable", service)
		if err := cmd.Run(); err == nil {
			return nil // Successfully enabled
		}
	}

	return fmt.Errorf("failed to enable any MariaDB service")
}

// StopMariaDBService stops the MariaDB service
func (s *ServiceManager) StopMariaDBService() error {
	services := []string{"mariadb", "mysql", "mysqld"}

	for _, service := range services {
		cmd := exec.Command("systemctl", "stop", service)
		if err := cmd.Run(); err == nil {
			return nil // Successfully stopped
		}
	}

	return fmt.Errorf("failed to stop any MariaDB service")
}
