package install

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// VersionSelector handles version selection from available versions
type VersionSelector struct {
	versions []SelectableVersion
}

// NewVersionSelector creates a new version selector
func NewVersionSelector(versions []SelectableVersion) *VersionSelector {
	return &VersionSelector{versions: versions}
}

// SelectVersion prompts user to select a MariaDB version
func (v *VersionSelector) SelectVersion(autoConfirm bool, defaultVersion string) (*SelectableVersion, error) {
	lg, _ := logger.Get()

	if len(v.versions) == 0 {
		return nil, fmt.Errorf("no MariaDB versions available for installation")
	}

	// If auto-confirm is enabled and default version is provided
	if autoConfirm && defaultVersion != "" {
		for _, version := range v.versions {
			if version.Version == defaultVersion {
				lg.Info("Auto-selected version", logger.String("version", defaultVersion))
				return &version, nil
			}
		}
		return nil, fmt.Errorf("specified default version %s not found in available versions", defaultVersion)
	}

	// Display available versions
	v.displayVersionTable()

	// If auto-confirm without specific version, select latest stable
	if autoConfirm {
		latest := v.getLatestStableVersion()
		lg.Info("Auto-selected latest stable version", logger.String("version", latest.Version))
		return latest, nil
	}

	// Interactive selection
	return v.promptForSelection()
}

// displayVersionTable displays available MariaDB versions in a table
func (v *VersionSelector) displayVersionTable() {
	terminal.PrintInfo("Available MariaDB versions:")

	headers := []string{"No", "Version", "Latest", "Support Type", "EOL Date"}
	rows := make([][]string, len(v.versions))

	for i, version := range v.versions {
		rows[i] = []string{
			strconv.Itoa(version.Index),
			version.Version,
			version.LatestVersion,
			version.SupportType,
			version.EOL,
		}
	}

	terminal.FormatTable(headers, rows)
}

// promptForSelection prompts user to select a version
func (v *VersionSelector) promptForSelection() (*SelectableVersion, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("Please select a version (1-%d): ", len(v.versions))

		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)

		// Parse selection
		selection, err := strconv.Atoi(input)
		if err != nil {
			terminal.PrintError("Invalid input. Please enter a number.")
			continue
		}

		// Validate selection
		if selection < 1 || selection > len(v.versions) {
			terminal.PrintError(fmt.Sprintf("Invalid selection. Please choose between 1 and %d.", len(v.versions)))
			continue
		}

		selectedVersion := &v.versions[selection-1]

		// Show confirmation
		v.displaySelectedVersion(selectedVersion)

		if v.confirmSelection() {
			return selectedVersion, nil
		}

		terminal.PrintInfo("Please select again.")
	}
}

// displaySelectedVersion shows the selected version details
func (v *VersionSelector) displaySelectedVersion(version *SelectableVersion) {
	terminal.PrintInfo(fmt.Sprintf("Selected MariaDB version: %s", version.Version))
	terminal.PrintInfo(fmt.Sprintf("Latest patch version: %s", version.LatestVersion))
	terminal.PrintInfo(fmt.Sprintf("Support type: %s", version.SupportType))
	terminal.PrintInfo(fmt.Sprintf("End of Life: %s", version.EOL))
}

// confirmSelection asks user to confirm their selection
func (v *VersionSelector) confirmSelection() bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Proceed with this version? (y/N): ")

		input, err := reader.ReadString('\n')
		if err != nil {
			terminal.PrintError("Failed to read input")
			return false
		}

		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "y", "yes":
			return true
		case "n", "no", "":
			return false
		default:
			terminal.PrintError("Please enter 'y' for yes or 'n' for no.")
		}
	}
}

// getLatestStableVersion returns the latest stable version from available versions
func (v *VersionSelector) getLatestStableVersion() *SelectableVersion {
	// For now, return the first version (they should be sorted)
	// In future, we might want to add logic to prefer LTS or stable versions
	if len(v.versions) > 0 {
		return &v.versions[0]
	}
	return nil
}

// GetVersionByIndex returns version by index (1-based)
func (v *VersionSelector) GetVersionByIndex(index int) (*SelectableVersion, error) {
	if index < 1 || index > len(v.versions) {
		return nil, fmt.Errorf("invalid version index: %d", index)
	}

	return &v.versions[index-1], nil
}

// GetVersionByString returns version by version string
func (v *VersionSelector) GetVersionByString(versionStr string) (*SelectableVersion, error) {
	for _, version := range v.versions {
		if version.Version == versionStr {
			return &version, nil
		}
	}

	return nil, fmt.Errorf("version %s not found in available versions", versionStr)
}

// GetAvailableVersions returns all available versions
func (v *VersionSelector) GetAvailableVersions() []SelectableVersion {
	return v.versions
}
