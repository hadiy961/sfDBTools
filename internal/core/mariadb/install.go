package mariadb

import (
	"fmt"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/mariadb"
	"time"
)

// InstallMariaDB installs MariaDB with custom configuration
func InstallMariaDB(options mariadb.InstallOptions) (*mariadb.InstallResult, error) {
	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	startTime := time.Now()
	result := &mariadb.InstallResult{
		Success:   false,
		Version:   options.Version,
		Port:      options.Port,
		DataDir:   options.DataDir,
		LogDir:    options.LogDir,
		BinlogDir: options.BinlogDir,
	}

	lg.Info("Starting MariaDB installation process",
		logger.String("version", options.Version),
		logger.Int("port", options.Port),
		logger.Bool("force", options.Force))

	// Step 1: Validate version
	if !mariadb.IsValidVersionWithConnectivityCheck(options.Version, false) {
		result.Duration = time.Since(startTime)
		return result, fmt.Errorf("unsupported MariaDB version: %s", options.Version)
	}

	// Step 3: Detect OS (without connectivity check since it's already verified)
	lg.Info("Detecting operating system and distribution")
	osInfo, err := mariadb.DetectOS()
	if err != nil {
		result.Duration = time.Since(startTime)
		return result, fmt.Errorf("failed to detect operating system: %w", err)
	}

	result.OperatingSystem = osInfo.ID
	result.Distribution = fmt.Sprintf("%s %s", osInfo.Name, osInfo.Version)

	lg.Info("OS detected",
		logger.String("os", osInfo.ID),
		logger.String("distribution", result.Distribution))

	// Step 4: Validate version for this OS
	if err := mariadb.ValidateVersionForOS(options.Version, osInfo); err != nil {
		result.Duration = time.Since(startTime)
		lg.Warn("Version validation failed",
			logger.Error(err),
			logger.String("recommended_version", mariadb.GetRecommendedVersion(osInfo)))
		return result, fmt.Errorf("version validation failed: %w", err)
	}

	// Step 5: Check for existing installation
	if !options.Force {
		if serviceInfo, err := mariadb.GetServiceInfo(); err == nil && serviceInfo.Status != "not-found" {
			result.Duration = time.Since(startTime)
			return result, fmt.Errorf("MariaDB service already exists (status: %s). Use --force to override", serviceInfo.Status)
		}
	}

	// Step 6: Perform installation
	lg.Info("Performing MariaDB installation")
	installResult, err := mariadb.InstallMariaDB(options, osInfo)
	if err != nil {
		result.Duration = time.Since(startTime)
		return result, fmt.Errorf("installation failed: %w", err)
	}

	// Copy results from install function
	result.Success = installResult.Success
	result.ServiceStatus = installResult.ServiceStatus
	result.Duration = time.Since(startTime)

	if result.Success {
		lg.Info("MariaDB installation completed successfully",
			logger.String("duration", result.Duration.String()),
			logger.String("version", result.Version),
			logger.String("service_status", result.ServiceStatus))
	}

	return result, nil
}
