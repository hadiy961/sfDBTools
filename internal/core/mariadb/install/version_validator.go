package install

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
)

// VersionValidator handles validation of MariaDB versions against repository availability
type VersionValidator struct {
	osInfo *system.OSInfo
}

// NewVersionValidator creates a new version validator instance
func NewVersionValidator(osInfo *system.OSInfo) *VersionValidator {
	return &VersionValidator{
		osInfo: osInfo,
	}
}

// SupportedVersions returns the list of versions supported by the official repository script
func (vv *VersionValidator) GetSupportedVersions() ([]string, error) {
	lg, _ := logger.Get()

	lg.Info("Checking supported versions from MariaDB repository script")

	// Try to get supported versions by running the repository script with an invalid version
	// This will return an error message containing the supported versions
	scriptURL := "https://r.mariadb.com/downloads/mariadb_repo_setup"
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf("curl -LsSf %s | sudo bash -s -- --mariadb-server-version=invalid_version_to_get_list 2>&1", scriptURL))

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		// Parse the error output to extract supported versions
		versions := vv.parseSupportedVersionsFromError(outputStr)
		if len(versions) > 0 {
			lg.Info("Successfully extracted supported versions from repository script",
				logger.Int("count", len(versions)),
				logger.Strings("versions", versions))
			return versions, nil
		}

		// Fallback to hardcoded known versions if parsing fails
		lg.Warn("Failed to parse supported versions from script output, using fallback list",
			logger.String("output", outputStr))
		return vv.getFallbackVersions(), nil
	}

	// If no error, try to parse supported versions anyway
	versions := vv.parseSupportedVersionsFromError(outputStr)
	if len(versions) > 0 {
		return versions, nil
	}

	// Final fallback
	return vv.getFallbackVersions(), nil
}

// parseSupportedVersionsFromError parses supported versions from repository script error output
func (vv *VersionValidator) parseSupportedVersionsFromError(output string) []string {
	// Look for the line containing version numbers (format like "10.6.23 10.11.14 11.4.8 11.8.3")
	re := regexp.MustCompile(`(?m)^\s*(\d+\.\d+\.\d+(?:\s+\d+\.\d+\.\d+)*)\s*$`)
	matches := re.FindStringSubmatch(output)

	if len(matches) > 1 {
		// Split the version string into individual versions
		versionLine := strings.TrimSpace(matches[1])
		versions := strings.Fields(versionLine)

		// Filter out any non-version strings
		var validVersions []string
		versionRegex := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
		for _, version := range versions {
			if versionRegex.MatchString(version) {
				validVersions = append(validVersions, version)
			}
		}

		return validVersions
	}

	return nil
}

// getFallbackVersions returns hardcoded fallback versions based on the error message we saw
func (vv *VersionValidator) getFallbackVersions() []string {
	// Based on the error message: "10.6.23 10.11.14 11.4.8 11.8.3"
	return []string{"10.6.23", "10.11.14", "11.4.8", "11.8.3"}
}

// IsVersionSupported checks if a specific version is supported by the repository script
func (vv *VersionValidator) IsVersionSupported(version string) (bool, error) {
	supportedVersions, err := vv.GetSupportedVersions()
	if err != nil {
		return false, err
	}

	for _, supportedVersion := range supportedVersions {
		if supportedVersion == version {
			return true, nil
		}
	}

	return false, nil
}

// FindBestMatch finds the best matching supported version for a requested version
func (vv *VersionValidator) FindBestMatch(requestedVersion string) (string, error) {
	supportedVersions, err := vv.GetSupportedVersions()
	if err != nil {
		return "", err
	}

	// First check for exact match
	for _, supportedVersion := range supportedVersions {
		if supportedVersion == requestedVersion {
			return supportedVersion, nil
		}
	}

	// Parse requested version to find closest match
	requestedMajor, requestedMinor, err := vv.parseVersion(requestedVersion)
	if err != nil {
		// If we can't parse, return the latest stable version
		return vv.getLatestVersion(supportedVersions), nil
	}

	// Find closest version by major.minor
	var bestMatch string
	var bestScore int = -1

	for _, supportedVersion := range supportedVersions {
		supportedMajor, supportedMinor, err := vv.parseVersion(supportedVersion)
		if err != nil {
			continue
		}

		// Calculate match score
		score := vv.calculateVersionScore(requestedMajor, requestedMinor, supportedMajor, supportedMinor)
		if score > bestScore {
			bestScore = score
			bestMatch = supportedVersion
		}
	}

	if bestMatch == "" {
		// Fallback to latest version
		return vv.getLatestVersion(supportedVersions), nil
	}

	return bestMatch, nil
}

// parseVersion extracts major and minor version numbers
func (vv *VersionValidator) parseVersion(version string) (int, int, error) {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("invalid version format: %s", version)
	}

	var major, minor int
	if _, err := fmt.Sscanf(parts[0], "%d", &major); err != nil {
		return 0, 0, fmt.Errorf("invalid major version: %s", parts[0])
	}

	if _, err := fmt.Sscanf(parts[1], "%d", &minor); err != nil {
		return 0, 0, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	return major, minor, nil
}

// calculateVersionScore calculates how close two versions are
func (vv *VersionValidator) calculateVersionScore(reqMajor, reqMinor, supMajor, supMinor int) int {
	// Exact major.minor match gets highest score
	if reqMajor == supMajor && reqMinor == supMinor {
		return 1000
	}

	// Same major version gets medium score
	if reqMajor == supMajor {
		minorDiff := reqMinor - supMinor
		if minorDiff < 0 {
			minorDiff = -minorDiff
		}
		return 500 - minorDiff
	}

	// Different major version gets low score
	majorDiff := reqMajor - supMajor
	if majorDiff < 0 {
		majorDiff = -majorDiff
	}
	return 100 - majorDiff
}

// getLatestVersion returns the latest version from supported versions
func (vv *VersionValidator) getLatestVersion(versions []string) string {
	if len(versions) == 0 {
		return ""
	}

	// Simple approach: return the last version (assuming they're sorted)
	// For more sophisticated sorting, we could implement proper version comparison
	latest := versions[0]
	for _, version := range versions {
		if vv.isVersionGreater(version, latest) {
			latest = version
		}
	}

	return latest
}

// isVersionGreater compares two versions to see if first is greater than second
func (vv *VersionValidator) isVersionGreater(v1, v2 string) bool {
	major1, minor1, patch1 := vv.parseVersionComponents(v1)
	major2, minor2, patch2 := vv.parseVersionComponents(v2)

	if major1 != major2 {
		return major1 > major2
	}
	if minor1 != minor2 {
		return minor1 > minor2
	}
	return patch1 > patch2
}

// parseVersionComponents parses version into major, minor, patch integers
func (vv *VersionValidator) parseVersionComponents(version string) (int, int, int) {
	parts := strings.Split(version, ".")
	major, minor, patch := 0, 0, 0

	if len(parts) >= 1 {
		fmt.Sscanf(parts[0], "%d", &major)
	}
	if len(parts) >= 2 {
		fmt.Sscanf(parts[1], "%d", &minor)
	}
	if len(parts) >= 3 {
		fmt.Sscanf(parts[2], "%d", &patch)
	}

	return major, minor, patch
}
