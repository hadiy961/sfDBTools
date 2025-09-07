package install

import (
	"fmt"
	"time"

	"sfDBTools/internal/logger"
)

// InstallationValidator handles validation operations for installation
type InstallationValidator struct{}

// NewInstallationValidator creates a new installation validator instance
func NewInstallationValidator() *InstallationValidator {
	return &InstallationValidator{}
}

// CreateInstallResult creates a standardized installation result
func (iv *InstallationValidator) CreateInstallResult(success bool, message string, startTime time.Time) *InstallResult {
	return &InstallResult{
		Success:     success,
		Message:     message,
		InstalledAt: time.Now(),
		Duration:    time.Since(startTime),
	}
}

// CreateErrorResult creates an error result with proper formatting
func (iv *InstallationValidator) CreateErrorResult(err error, operation string, startTime time.Time) *InstallResult {
	message := fmt.Sprintf("%s failed: %v", operation, err)
	return iv.CreateInstallResult(false, message, startTime)
}

// CreateSuccessResult creates a success result with additional data
func (iv *InstallationValidator) CreateSuccessResult(version string, packagesCount int, serviceStatus string, startTime time.Time) *InstallResult {
	result := iv.CreateInstallResult(true, "MariaDB installed successfully", startTime)
	result.Version = version
	result.PackagesCount = packagesCount
	result.ServiceStatus = serviceStatus
	return result
}

// LogInstallationStart logs the start of installation process
func (iv *InstallationValidator) LogInstallationStart() error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}
	lg.Info("Starting MariaDB installation process")
	return nil
}

// LogInstallationSuccess logs successful installation completion
func (iv *InstallationValidator) LogInstallationSuccess(version string, duration time.Duration) {
	if lg, err := logger.Get(); err == nil {
		lg.Info("MariaDB installation completed successfully",
			logger.String("version", version),
			logger.String("duration", duration.String()))
	}
}

// ValidateConfig validates the installation configuration
func (iv *InstallationValidator) ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("installation config cannot be nil")
	}
	// Add more configuration validation as needed
	return nil
}
