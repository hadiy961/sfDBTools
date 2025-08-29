package install

import (
	"fmt"

	"sfDBTools/internal/core/mariadb/check_version"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// VersionManager handles version fetching and management
type VersionManager struct {
	versionService  *check_version.VersionService
	versionSelector *VersionSelector
	selectedVersion *SelectableVersion
}

// NewVersionManager creates a new version manager
func NewVersionManager() *VersionManager {
	return &VersionManager{}
}

// FetchAvailableVersions retrieves available MariaDB versions
func (v *VersionManager) FetchAvailableVersions() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Fetching available MariaDB versions...")
	spinner.Start()

	// Create version service
	versionConfig := check_version.DefaultCheckVersionConfig()
	v.versionService = check_version.NewVersionService(versionConfig)

	// Fetch versions
	versions, err := v.versionService.FetchAvailableVersions()
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to fetch MariaDB versions: %w", err)
	}

	if len(versions) == 0 {
		spinner.Stop()
		return fmt.Errorf("no MariaDB versions available for installation")
	}

	spinner.Stop()
	terminal.PrintSuccess(fmt.Sprintf("Found %d available MariaDB versions", len(versions)))

	// Convert to selectable versions
	selectableVersions := ConvertVersionInfo(versions)
	v.versionSelector = NewVersionSelector(selectableVersions)

	lg.Info("Available versions fetched successfully", logger.Int("count", len(versions)))
	return nil
}

// SelectVersion handles version selection
func (v *VersionManager) SelectVersion(autoConfirm bool, version string) (*SelectableVersion, error) {
	lg, _ := logger.Get()

	terminal.PrintInfo("Please select a MariaDB version to install:")

	selectedVersion, err := v.versionSelector.SelectVersion(autoConfirm, version)
	if err != nil {
		return nil, fmt.Errorf("version selection failed: %w", err)
	}

	v.selectedVersion = selectedVersion

	terminal.PrintSuccess(fmt.Sprintf("Selected MariaDB version: %s (%s)",
		selectedVersion.Version, selectedVersion.LatestVersion))

	lg.Info("Version selected",
		logger.String("major_version", selectedVersion.Version),
		logger.String("latest_version", selectedVersion.LatestVersion))

	return selectedVersion, nil
}

// GetSelectedVersion returns the currently selected version
func (v *VersionManager) GetSelectedVersion() *SelectableVersion {
	return v.selectedVersion
}
