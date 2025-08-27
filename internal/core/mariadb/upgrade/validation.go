package upgrade

import (
	"fmt"
	"strconv"
	"strings"

	"sfDBTools/internal/core/mariadb/check_version"
	"sfDBTools/internal/core/mariadb/remove"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
)

// ValidationService handles upgrade validation
type ValidationService struct {
	versionService   *check_version.VersionService
	detectionService *remove.DetectionService
	osInfo           *common.OSInfo
}

// NewValidationService creates a new validation service
func NewValidationService() *ValidationService {
	versionConfig := check_version.DefaultCheckVersionConfig()
	versionService := check_version.NewVersionService(versionConfig)

	// Detect OS info for detection service
	detector := common.NewOSDetector()
	osInfo, _ := detector.DetectOS()

	detectionService := remove.NewDetectionService(osInfo)

	return &ValidationService{
		versionService:   versionService,
		detectionService: detectionService,
		osInfo:           osInfo,
	}
}

// ValidateUpgrade validates if upgrade is possible and safe
func (v *ValidationService) ValidateUpgrade(config *UpgradeConfig) (*ValidationResult, error) {
	lg, _ := logger.Get()

	result := &ValidationResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		Suggestions: []string{},
	}

	lg.Info("Starting upgrade validation")

	// 1. Check if MariaDB is installed
	current, err := v.detectCurrentInstallation()
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to detect current installation: %v", err))
		return result, nil
	}

	if !current.IsInstalled {
		result.Valid = false
		result.Errors = append(result.Errors, "MariaDB is not installed - nothing to upgrade")
		return result, nil
	}

	// 2. Check available versions
	availableVersions, err := v.versionService.FetchAvailableVersions()
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to fetch available versions: %v", err))
		return result, nil
	}

	// 3. Validate target version
	if config.TargetVersion != "" {
		if !v.isVersionAvailable(config.TargetVersion, availableVersions) {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Target version %s is not available", config.TargetVersion))
		}
	}

	// 4. Check upgrade compatibility
	upgradeType := v.determineUpgradeType(current.Version, config.TargetVersion, availableVersions)

	switch upgradeType {
	case UpgradeTypeInvalid:
		if !config.ForceUpgrade {
			result.Valid = false
			result.Errors = append(result.Errors, "Invalid upgrade path (use --force-upgrade to override)")
		} else {
			result.Warnings = append(result.Warnings, "Forced upgrade may cause data incompatibility")
		}
	case UpgradeTypeNone:
		result.Valid = false
		result.Errors = append(result.Errors, "Already running the latest available version")
	case UpgradeTypeMajor:
		result.Warnings = append(result.Warnings, "Major version upgrade requires careful planning")
		result.Suggestions = append(result.Suggestions, "Consider testing upgrade on a copy of your data first")
	}

	// 5. Check system resources
	if err := v.validateSystemResources(); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("System resource warning: %v", err))
	}

	// 6. Check disk space
	if err := v.validateDiskSpace(current.DataDirectory); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Insufficient disk space: %v", err))
	}

	lg.Info("Upgrade validation completed",
		logger.Bool("valid", result.Valid),
		logger.Int("errors", len(result.Errors)),
		logger.Int("warnings", len(result.Warnings)))

	return result, nil
}

// detectCurrentInstallation detects current MariaDB installation
func (v *ValidationService) detectCurrentInstallation() (*CurrentInstallation, error) {
	installation, err := v.detectionService.DetectInstallation()
	if err != nil {
		return nil, fmt.Errorf("failed to detect installation: %w", err)
	}

	current := &CurrentInstallation{
		IsInstalled:    installation.IsInstalled,
		Version:        installation.Version,
		PackageName:    installation.PackageName,
		DataDirectory:  installation.ActualDataDir, // Use ActualDataDir
		ServiceName:    installation.ServiceName,
		ServiceRunning: installation.ServiceActive, // Use ServiceActive
		ServiceEnabled: installation.ServiceEnabled,
		ConfigFiles:    installation.ConfigFiles,
	}

	return current, nil
}

// isVersionAvailable checks if version is in available versions list
func (v *ValidationService) isVersionAvailable(targetVersion string, availableVersions []check_version.VersionInfo) bool {
	for _, version := range availableVersions {
		if version.Version == targetVersion || version.LatestVersion == targetVersion {
			return true
		}
	}
	return false
}

// determineUpgradeType determines the type of upgrade needed
func (v *ValidationService) determineUpgradeType(currentVersion, targetVersion string, availableVersions []check_version.VersionInfo) UpgradeType {
	// If no target version specified, find latest
	if targetVersion == "" {
		// Find latest available version
		latest := v.findLatestVersion(availableVersions)
		if latest == "" {
			return UpgradeTypeInvalid
		}
		targetVersion = latest
	}

	// Parse current version
	currentMajor, currentMinor := v.parseVersion(currentVersion)
	targetMajor, targetMinor := v.parseVersion(targetVersion)

	// Compare versions
	if targetMajor < currentMajor {
		return UpgradeTypeInvalid // Downgrade
	}

	if targetMajor > currentMajor {
		return UpgradeTypeMajor // Major upgrade
	}

	// Same major version
	if targetMinor < currentMinor {
		return UpgradeTypeInvalid // Downgrade
	}

	if targetMinor > currentMinor {
		return UpgradeTypeMinor // Minor upgrade
	}

	// Check patch level (simplified)
	if targetVersion != currentVersion {
		return UpgradeTypePatch // Patch upgrade
	}

	return UpgradeTypeNone // Same version
}

// parseVersion extracts major and minor version numbers
func (v *ValidationService) parseVersion(version string) (int, int) {
	// Remove any package suffix (e.g., "10.6.23-1.el9.x86_64" -> "10.6.23")
	cleanVersion := strings.Split(version, "-")[0]

	// Handle MariaDB prefix (e.g., "11.8.3-MariaDB" -> "11.8.3")
	if strings.Contains(cleanVersion, "-") {
		cleanVersion = strings.Split(cleanVersion, "-")[0]
	}

	parts := strings.Split(cleanVersion, ".")

	major, minor := 0, 0

	if len(parts) >= 1 {
		if val, err := strconv.Atoi(parts[0]); err == nil {
			major = val
		}
	}

	if len(parts) >= 2 {
		if val, err := strconv.Atoi(parts[1]); err == nil {
			minor = val
		}
	}

	return major, minor
}

// findLatestVersion finds the latest version from available versions
func (v *ValidationService) findLatestVersion(availableVersions []check_version.VersionInfo) string {
	if len(availableVersions) == 0 {
		return ""
	}

	// Return the latest version of the first (newest) available version series
	return availableVersions[0].LatestVersion
}

// validateSystemResources checks if system has enough resources
func (v *ValidationService) validateSystemResources() error {
	// Basic system resource validation
	// This could be expanded to check memory, CPU, etc.
	return nil
}

// validateDiskSpace checks if there's enough disk space for upgrade
func (v *ValidationService) validateDiskSpace(dataDir string) error {
	// Check disk space in data directory
	// This is a simplified check - in reality, you'd want to check actual usage
	return nil
}

// GetCurrentInstallation returns current installation details
func (v *ValidationService) GetCurrentInstallation() (*CurrentInstallation, error) {
	return v.detectCurrentInstallation()
}
