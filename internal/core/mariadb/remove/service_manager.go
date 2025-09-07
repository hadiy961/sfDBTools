package remove

import (
	"fmt"
	"os/exec"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// ServiceManager handles MariaDB service operations during removal
type ServiceManager struct {
	svcManager system.ServiceManager
}

// NewServiceManager creates a new service manager for removal operations
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		svcManager: system.NewServiceManager(),
	}
}

// StopAndDisableServices checks, stops and disables MariaDB services
func (sm *ServiceManager) StopAndDisableServices() {
	lg, _ := logger.Get()
	services := []string{"mariadb", "mysql"}

	for _, svcName := range services {
		if sm.serviceExists(svcName) {
			terminal.PrintInfo(fmt.Sprintf("Found %s service", svcName))
			sm.stopService(svcName, lg)
			sm.disableService(svcName, lg)
		} else {
			terminal.PrintInfo(fmt.Sprintf("%s service not found, skipping", svcName))
		}
	}
}

// serviceExists checks if a service unit file exists in the system
func (sm *ServiceManager) serviceExists(serviceName string) bool {
	cmd := exec.Command("systemctl", "cat", serviceName)
	err := cmd.Run()
	return err == nil
}

// stopService stops a MariaDB service if it's running
func (sm *ServiceManager) stopService(svcName string, lg *logger.Logger) {
	if sm.svcManager.IsActive(svcName) {
		terminal.PrintInfo(fmt.Sprintf("Stopping %s service...", svcName))
		if err := sm.svcManager.Stop(svcName); err != nil {
			lg.Warn("Failed to stop service", logger.String("service", svcName), logger.Error(err))
			terminal.PrintWarning(fmt.Sprintf("⚠️  Failed to stop %s service: %v", svcName, err))
		} else {
			terminal.PrintSuccess(fmt.Sprintf("✅ Stopped %s service", svcName))
		}
	} else {
		terminal.PrintInfo(fmt.Sprintf("%s service is not running", svcName))
	}
}

// disableService disables a MariaDB service if it's enabled
func (sm *ServiceManager) disableService(svcName string, lg *logger.Logger) {
	if sm.svcManager.IsEnabled(svcName) {
		terminal.PrintInfo(fmt.Sprintf("Disabling %s service...", svcName))
		if err := sm.svcManager.Disable(svcName); err != nil {
			lg.Warn("Failed to disable service", logger.String("service", svcName), logger.Error(err))
			terminal.PrintWarning(fmt.Sprintf("⚠️  Failed to disable %s service: %v", svcName, err))
		} else {
			terminal.PrintSuccess(fmt.Sprintf("✅ Disabled %s service", svcName))
		}
	} else {
		terminal.PrintInfo(fmt.Sprintf("%s service is not enabled", svcName))
	}
}
