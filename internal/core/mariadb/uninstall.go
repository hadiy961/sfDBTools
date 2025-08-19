package mariadb

import (
	"fmt"
	"time"

	"sfDBTools/internal/logger"
	mariadb_utils "sfDBTools/utils/mariadb"
)

// UninstallMariaDB completely uninstalls MariaDB from the system
func UninstallMariaDB(options mariadb_utils.UninstallOptions) (*mariadb_utils.UninstallResult, error) {
	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	startTime := time.Now()
	result := &mariadb_utils.UninstallResult{
		Success: false,
	}

	lg.Info("Starting MariaDB uninstall process",
		logger.Bool("force", options.Force),
		logger.Bool("keep_data", options.KeepData),
		logger.Bool("keep_config", options.KeepConfig))

	// Step 1: Detect OS
	lg.Info("Detecting operating system and distribution")
	osInfo, err := mariadb_utils.DetectOS()
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to detect OS: %v", err))
		result.Duration = time.Since(startTime)
		return result, fmt.Errorf("failed to detect operating system: %w", err)
	}

	result.OperatingSystem = osInfo.ID
	result.Distribution = fmt.Sprintf("%s %s", osInfo.Name, osInfo.Version)

	lg.Info("OS detected",
		logger.String("os", osInfo.ID),
		logger.String("distribution", result.Distribution))

	// Step 2: Check current service status
	lg.Info("Checking MariaDB service status")
	serviceInfo, err := mariadb_utils.GetServiceInfo()
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to get service info: %v", err))
	} else {
		result.ServiceStatus = serviceInfo.Status
		lg.Info("Service status", logger.String("status", serviceInfo.Status))

		// If service is not found, check if any packages are installed
		if serviceInfo.Status == "not-found" {
			lg.Info("MariaDB service not found, checking for installed packages")
			packagesCount, _, err := mariadb_utils.GetInstalledPackageCount(osInfo)
			if err != nil {
				lg.Warn("Failed to check installed packages", logger.Error(err))
			} else if packagesCount == 0 {
				lg.Info("No MariaDB packages found, system appears to be clean")
				result.Success = true
				result.ServiceStatus = "not installed"
				result.Duration = time.Since(startTime)

				lg.Info("MariaDB uninstall completed - system was already clean",
					logger.Bool("success", result.Success),
					logger.String("duration", result.Duration.String()))

				return result, nil
			} else {
				lg.Info("Found MariaDB packages without service, continuing cleanup",
					logger.Int("packages", packagesCount))
			}
		}
	}

	// Step 3: Stop and disable services
	lg.Info("Stopping MariaDB service")
	if err := mariadb_utils.StopService(); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to stop service: %v", err))
	}

	lg.Info("Disabling MariaDB service")
	if err := mariadb_utils.DisableService(); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to disable service: %v", err))
	}

	// Step 4a: Mask and cleanup systemd services
	lg.Info("Cleaning up systemd services")
	if err := mariadb_utils.MaskAndRemoveServices(); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to cleanup systemd services: %v", err))
	}

	// Step 4: Remove packages
	lg.Info("Removing MariaDB packages")
	packagesCount, _, err := mariadb_utils.RemovePackages(osInfo)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Package removal issues: %v", err))
	}
	result.PackagesRemoved = packagesCount

	lg.Info("Package removal completed", logger.Int("packages", packagesCount))

	// Step 5: Cleanup directories
	lg.Info("Cleaning up directories and configuration files")
	removedDirs, err := mariadb_utils.CleanupDirectories(options.KeepData, options.KeepConfig)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Directory cleanup issues: %v", err))
	}
	result.DirectoriesRemoved = removedDirs

	lg.Info("Directory cleanup completed", logger.Int("removed", len(removedDirs)))

	// Step 6: Remove repositories
	lg.Info("Removing MariaDB repositories")
	removedRepos, err := mariadb_utils.CleanupRepositories(osInfo)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Repository cleanup issues: %v", err))
	}
	result.RepositoriesRemoved = removedRepos

	lg.Info("Repository cleanup completed", logger.Int("removed", len(removedRepos)))

	// Step 7: Verification
	lg.Info("Verifying MariaDB uninstall")
	verifySuccess, verifyWarnings, err := mariadb_utils.VerifyUninstall(osInfo)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Verification error: %v", err))
	}

	if len(verifyWarnings) > 0 {
		result.Warnings = append(result.Warnings, verifyWarnings...)
	}

	// Determine final status
	result.Duration = time.Since(startTime)
	result.Success = verifySuccess && len(result.Errors) == 0

	if result.Success {
		result.ServiceStatus = "completely removed"
	}

	lg.Info("MariaDB uninstall completed",
		logger.Bool("success", result.Success),
		logger.String("duration", result.Duration.String()),
		logger.Int("warnings", len(result.Warnings)),
		logger.Int("errors", len(result.Errors)))

	return result, nil
}
