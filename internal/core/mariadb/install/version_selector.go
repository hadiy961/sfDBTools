package install

import (
	"fmt"

	"sfDBTools/internal/core/mariadb/check_version"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/terminal"
)

// VersionSelector handles MariaDB version selection operations
type VersionSelector struct {
	config    *Config
	validator *VersionValidator
}

// NewVersionSelector creates a new version selector instance
func NewVersionSelector(config *Config, osInfo *common.OSInfo) *VersionSelector {
	return &VersionSelector{
		config:    config,
		validator: NewVersionValidator(osInfo),
	}
}

// SelectVersion allows user to select MariaDB version to install
func (vs *VersionSelector) SelectVersion() (string, error) {
	lg, _ := logger.Get()
	
	// First, get repository-supported versions
	supportedVersions, err := vs.validator.GetSupportedVersions()
	if err != nil {
		lg.Warn("Failed to get repository-supported versions, falling back to version check", logger.Error(err))
		return vs.selectVersionFallback()
	}
	
	if len(supportedVersions) == 0 {
		lg.Warn("No repository-supported versions found, falling back to version check")
		return vs.selectVersionFallback()
	}
	
	// Use repository-supported versions directly
	stableVersions, versionOptions := vs.prepareRepositoryVersionOptions(supportedVersions)
	
	// Handle auto-selection for non-interactive mode
	if vs.config != nil && vs.config.SkipConfirm {
		return vs.autoSelectRepositoryVersion(supportedVersions)
	}
	
	// Interactive version selection
	return vs.interactiveVersionSelection(stableVersions, versionOptions)
}

// selectVersionFallback falls back to the original method with validation
func (vs *VersionSelector) selectVersionFallback() (string, error) {
	availableVersions, err := vs.fetchAvailableVersions()
	if err != nil {
		return "", err
	}

	stableVersions, versionOptions := vs.prepareVersionOptions(availableVersions)
	if len(versionOptions) == 0 {
		terminal.PrintError("No stable versions available for installation")
		return "", fmt.Errorf("no stable versions available for installation")
	}

	var selectedVersion string
	
	// Handle auto-selection for non-interactive mode
	if vs.config != nil && vs.config.SkipConfirm {
		selectedVersion, err = vs.autoSelectVersion(availableVersions, stableVersions)
	} else {
		// Interactive version selection
		selectedVersion, err = vs.interactiveVersionSelection(stableVersions, versionOptions)
	}
	
	if err != nil {
		return "", err
	}
	
	// Validate the selected version against repository support
	return vs.validateAndOfferAlternative(selectedVersion)
}

// fetchAvailableVersions fetches available MariaDB versions
func (vs *VersionSelector) fetchAvailableVersions() (*check_version.VersionCheckResult, error) {
	spinner := terminal.NewProgressSpinnerWithStyle("Fetching available MariaDB versions...", terminal.SpinnerMinimal)
	spinner.Start()

	checkerConfig := check_version.DefaultConfig()
	checker, err := check_version.NewChecker(checkerConfig)
	if err != nil {
		spinner.StopWithError("Failed to create version checker")
		return nil, fmt.Errorf("failed to create version checker: %w", err)
	}

	result, err := checker.CheckAvailableVersions()
	if err != nil {
		spinner.StopWithError("Failed to fetch available versions")
		return nil, fmt.Errorf("failed to get available versions: %w", err)
	}

	if len(result.AvailableVersions) == 0 {
		spinner.StopWithError("No MariaDB versions available")
		return nil, fmt.Errorf("no MariaDB versions available")
	}

	spinner.StopWithSuccess("Available MariaDB versions retrieved")
	return result, nil
}

// prepareVersionOptions prepares version options for selection menu
func (vs *VersionSelector) prepareVersionOptions(result *check_version.VersionCheckResult) ([]string, []string) {
	var stableVersions []string
	var versionOptions []string

	for _, version := range result.AvailableVersions {
		if version.Type == "stable" {
			stableVersions = append(stableVersions, version.Version)
			option := fmt.Sprintf("MariaDB %s (Stable)", version.Version)
			if version.Version == result.CurrentStable {
				option += " [Recommended]"
			}
			versionOptions = append(versionOptions, option)
		}
	}

	return stableVersions, versionOptions
}

// autoSelectVersion automatically selects version for non-interactive mode
func (vs *VersionSelector) autoSelectVersion(result *check_version.VersionCheckResult, stableVersions []string) (string, error) {
	lg, _ := logger.Get()

	// Prefer current stable
	if result.CurrentStable != "" {
		terminal.PrintInfo(fmt.Sprintf("Auto-selecting recommended version: MariaDB %s", result.CurrentStable))
		terminal.PrintSuccess(fmt.Sprintf("Selected MariaDB %s for installation", result.CurrentStable))
		lg.Info("Auto-selected recommended version", logger.String("version", result.CurrentStable))
		return result.CurrentStable, nil
	}

	// Fallback to first stable
	if len(stableVersions) > 0 {
		selectedVersion := stableVersions[0]
		terminal.PrintInfo(fmt.Sprintf("Auto-selecting first stable version: MariaDB %s", selectedVersion))
		terminal.PrintSuccess(fmt.Sprintf("Selected MariaDB %s for installation", selectedVersion))
		lg.Info("Auto-selected first stable version", logger.String("version", selectedVersion))
		return selectedVersion, nil
	}

	return "", fmt.Errorf("no stable versions available for auto-selection")
}

// interactiveVersionSelection handles interactive version selection
func (vs *VersionSelector) interactiveVersionSelection(stableVersions []string, versionOptions []string) (string, error) {
	lg, _ := logger.Get()

	// Show version selection menu
	terminal.ClearAndShowHeader("MariaDB Version Selection")
	terminal.PrintInfo("Select a MariaDB version to install:")

	selected, err := terminal.ShowMenuAndClear("Available Versions", versionOptions)
	if err != nil {
		return "", fmt.Errorf("version selection failed: %w", err)
	}

	selectedVersion := stableVersions[selected-1]

	lg.Info("User selected version", logger.String("version", selectedVersion))
	terminal.PrintSuccess(fmt.Sprintf("Selected MariaDB %s for installation", selectedVersion))

	return selectedVersion, nil
}

// prepareRepositoryVersionOptions prepares version options from repository-supported versions
func (vs *VersionSelector) prepareRepositoryVersionOptions(supportedVersions []string) ([]string, []string) {
	var versionOptions []string

	for _, version := range supportedVersions {
		option := fmt.Sprintf("MariaDB %s (Repository Supported)", version)
		versionOptions = append(versionOptions, option)
	}

	return supportedVersions, versionOptions
}

// autoSelectRepositoryVersion automatically selects version from repository-supported versions
func (vs *VersionSelector) autoSelectRepositoryVersion(supportedVersions []string) (string, error) {
	lg, _ := logger.Get()
	
	if len(supportedVersions) == 0 {
		return "", fmt.Errorf("no repository-supported versions available for auto-selection")
	}
	
	// Select the latest version (last in the list)
	selectedVersion := vs.validator.getLatestVersion(supportedVersions)
	
	terminal.PrintInfo(fmt.Sprintf("Auto-selecting latest repository-supported version: MariaDB %s", selectedVersion))
	terminal.PrintSuccess(fmt.Sprintf("Selected MariaDB %s for installation", selectedVersion))
	lg.Info("Auto-selected latest repository-supported version", logger.String("version", selectedVersion))
	
	return selectedVersion, nil
}

// validateAndOfferAlternative validates version and offers alternatives if not supported
func (vs *VersionSelector) validateAndOfferAlternative(requestedVersion string) (string, error) {
	lg, _ := logger.Get()
	
	// Check if the version is supported
	isSupported, err := vs.validator.IsVersionSupported(requestedVersion)
	if err != nil {
		lg.Warn("Failed to validate version support", 
			logger.String("version", requestedVersion),
			logger.Error(err))
		return requestedVersion, nil // Continue with original version on validation error
	}
	
	if isSupported {
		return requestedVersion, nil
	}
	
	// Version not supported, find best match
	lg.Warn("Requested version not supported by repository", 
		logger.String("requested", requestedVersion))
	
	bestMatch, err := vs.validator.FindBestMatch(requestedVersion)
	if err != nil {
		return "", fmt.Errorf("failed to find alternative version: %w", err)
	}
	
	// In interactive mode, ask user for confirmation
	if vs.config == nil || !vs.config.SkipConfirm {
		terminal.PrintWarning(fmt.Sprintf("⚠️ MariaDB %s is not supported by the official repository", requestedVersion))
		terminal.PrintInfo(fmt.Sprintf("📋 Closest supported version is: MariaDB %s", bestMatch))
		
		confirmed, err := terminal.ConfirmAndClear(fmt.Sprintf("Would you like to install MariaDB %s instead?", bestMatch))
		if err != nil {
			return "", fmt.Errorf("failed to get user confirmation: %w", err)
		}
		
		if !confirmed {
			return "", fmt.Errorf("installation cancelled by user")
		}
	} else {
		// Auto-confirm in non-interactive mode
		terminal.PrintWarning(fmt.Sprintf("⚠️ MariaDB %s is not supported by the official repository", requestedVersion))
		terminal.PrintInfo(fmt.Sprintf("📋 Auto-selecting closest supported version: MariaDB %s", bestMatch))
	}
	
	terminal.PrintSuccess(fmt.Sprintf("Selected MariaDB %s for installation", bestMatch))
	lg.Info("Selected alternative version", 
		logger.String("requested", requestedVersion),
		logger.String("selected", bestMatch))
	
	return bestMatch, nil
}
