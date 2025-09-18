package system

import (
	"fmt"
	"os/exec"
	"strings"
)

// ServiceManager interface provides abstraction for service management operations
type ServiceManager interface {
	Stop(name string) error
	Start(name string) error
	Restart(name string) error
	Reload(name string) error
	Disable(name string) error
	Enable(name string) error
	IsActive(name string) bool
	IsEnabled(name string) bool
	GetStatus(name string) (ServiceStatus, error)
}

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Name    string
	Active  bool
	Enabled bool
	Running bool
}

// serviceManager implements ServiceManager interface using systemctl
type serviceManager struct{}

// NewServiceManager creates a new service manager
func NewServiceManager() ServiceManager {
	return &serviceManager{}
}

// Stop stops a service
func (sm *serviceManager) Stop(name string) error {
	cmd := exec.Command("systemctl", "stop", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop service %s: %w\nOutput: %s", name, err, string(output))
	}
	return nil
}

// Start starts a service
func (sm *serviceManager) Start(name string) error {
	cmd := exec.Command("systemctl", "start", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start service %s: %w\nOutput: %s", name, err, string(output))
	}
	return nil
}

// Restart service
func (sm *serviceManager) Restart(name string) error {
	cmd := exec.Command("systemctl", "restart", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart service %s: %w\nOutput: %s", name, err, string(output))
	}
	return nil
}

// Reload reloads a service
func (sm *serviceManager) Reload(name string) error {
	cmd := exec.Command("systemctl", "reload", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reload service %s: %w\nOutput: %s", name, err, string(output))
	}
	return nil
}

// Disable disables a service
func (sm *serviceManager) Disable(name string) error {
	cmd := exec.Command("systemctl", "disable", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to disable service %s: %w\nOutput: %s", name, err, string(output))
	}
	return nil
}

// Enable enables a service
func (sm *serviceManager) Enable(name string) error {
	cmd := exec.Command("systemctl", "enable", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to enable service %s: %w\nOutput: %s", name, err, string(output))
	}
	return nil
}

// IsActive checks if a service is currently active/running
func (sm *serviceManager) IsActive(name string) bool {
	cmd := exec.Command("systemctl", "is-active", name)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "active"
}

// IsEnabled checks if a service is enabled
func (sm *serviceManager) IsEnabled(name string) bool {
	cmd := exec.Command("systemctl", "is-enabled", name)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	status := strings.TrimSpace(string(output))
	return status == "enabled"
}

// GetStatus returns comprehensive status information for a service
func (sm *serviceManager) GetStatus(name string) (ServiceStatus, error) {
	status := ServiceStatus{
		Name:    name,
		Active:  sm.IsActive(name),
		Enabled: sm.IsEnabled(name),
		Running: sm.IsActive(name), // For systemd, active generally means running
	}
	return status, nil
}
