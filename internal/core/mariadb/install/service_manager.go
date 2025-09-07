package install

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// ServiceConfigManager handles MariaDB service configuration operations
type ServiceConfigManager struct {
	svcManager system.ServiceManager
}

// NewServiceConfigManager creates a new service configuration manager instance
func NewServiceConfigManager(svcManager system.ServiceManager) *ServiceConfigManager {
	return &ServiceConfigManager{
		svcManager: svcManager,
	}
}

// ConfigureService starts and enables MariaDB service
func (scm *ServiceConfigManager) ConfigureService() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProcessingSpinner("Configuring MariaDB service...")
	spinner.Start()

	serviceName := "mariadb"

	// Start the service
	if err := scm.startService(spinner, serviceName); err != nil {
		spinner.StopWithError("Failed to start service")
		return fmt.Errorf("failed to start service: %w", err)
	}

	// Enable the service
	if err := scm.enableService(spinner, serviceName); err != nil {
		spinner.StopWithError("Failed to enable service")
		return fmt.Errorf("failed to enable service: %w", err)
	}

	lg.Info("Service configured successfully", logger.String("service", serviceName))
	spinner.StopWithSuccess("Service configuration completed")

	return nil
}

// VerifyInstallation verifies the installation was successful
func (scm *ServiceConfigManager) VerifyInstallation() (string, error) {
	lg, _ := logger.Get()

	spinner := terminal.NewLoadingSpinner("Verifying installation...")
	spinner.Start()

	serviceName := "mariadb"

	// Check service status
	spinner.UpdateMessage("Checking service status...")
	status, err := scm.svcManager.GetStatus(serviceName)
	if err != nil {
		spinner.StopWithError("Failed to get service status")
		return "", fmt.Errorf("failed to get service status: %w", err)
	}

	if !status.Active || !status.Enabled {
		statusMsg := fmt.Sprintf("Service issues - Active: %v, Enabled: %v", status.Active, status.Enabled)
		spinner.StopWithWarning("Service is not properly configured")
		return statusMsg, fmt.Errorf("service is not properly configured")
	}

	lg.Info("Installation verification completed",
		logger.Bool("active", status.Active),
		logger.Bool("enabled", status.Enabled))

	spinner.StopWithSuccess("Installation verification completed")

	return "Active and Enabled", nil
}

// startService starts the MariaDB service
func (scm *ServiceConfigManager) startService(spinner *terminal.ProgressSpinner, serviceName string) error {
	spinner.UpdateMessage("Starting MariaDB service...")
	return scm.svcManager.Start(serviceName)
}

// enableService enables the MariaDB service
func (scm *ServiceConfigManager) enableService(spinner *terminal.ProgressSpinner, serviceName string) error {
	spinner.UpdateMessage("Enabling MariaDB service...")
	return scm.svcManager.Enable(serviceName)
}

// GetServiceName returns the service name used for MariaDB
func (scm *ServiceConfigManager) GetServiceName() string {
	return "mariadb"
}
