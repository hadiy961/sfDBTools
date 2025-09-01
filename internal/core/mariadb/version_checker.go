package mariadb

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"sfDBTools/internal/logger"
)

// VersionInfo represents MariaDB version information
type VersionInfo struct {
	Version     string `json:"version"`
	Type        string `json:"type"` // stable, rc, rolling
	ReleaseDate string `json:"release_date,omitempty"`
}

// VersionCheckResult contains the result of version checking
type VersionCheckResult struct {
	AvailableVersions []VersionInfo `json:"available_versions"`
	CurrentStable     string        `json:"current_stable"`
	LatestVersion     string        `json:"latest_version"`
	LatestMinor       string        `json:"latest_minor"`
	CheckTime         time.Time     `json:"check_time"`
}

// VersionChecker handles MariaDB version checking operations
type VersionChecker struct {
	timeout time.Duration
}

// NewVersionChecker creates a new version checker instance
func NewVersionChecker() *VersionChecker {
	return &VersionChecker{
		timeout: 30 * time.Second,
	}
}

// CheckAvailableVersions fetches available MariaDB versions from official sources
func (vc *VersionChecker) CheckAvailableVersions() (*VersionCheckResult, error) {
	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting MariaDB version check")

	// Get supported versions from the repository setup script
	versions, err := vc.fetchSupportedVersions()
	if err != nil {
		lg.Error("Failed to fetch supported versions", logger.Error(err))
		return nil, fmt.Errorf("failed to fetch supported versions: %w", err)
	}

	// Process and categorize versions
	result := &VersionCheckResult{
		AvailableVersions: versions,
		CheckTime:         time.Now(),
	}

	// Find current stable and latest versions
	result.CurrentStable = vc.findCurrentStable(versions)
	result.LatestVersion = vc.findLatestVersion(versions)
	result.LatestMinor = vc.findLatestMinor(versions)

	lg.Info("MariaDB version check completed",
		logger.Int("versions_found", len(versions)),
		logger.String("current_stable", result.CurrentStable),
		logger.String("latest_version", result.LatestVersion),
		logger.String("latest_minor", result.LatestMinor))

	return result, nil
}

// fetchSupportedVersions gets supported versions dynamically from MariaDB sources
func (vc *VersionChecker) fetchSupportedVersions() ([]VersionInfo, error) {
	lg, _ := logger.Get()

	lg.Info("Fetching MariaDB versions dynamically from official sources")

	// Try to fetch from multiple sources
	versions, err := vc.fetchVersionsFromOfficialSources()
	if err != nil {
		lg.Warn("Failed to fetch versions dynamically, falling back to repository script", logger.Error(err))
		// Fallback to script parsing
		versions, err = vc.fetchVersionsFromScript()
		if err != nil {
			lg.Error("All dynamic fetching methods failed", logger.Error(err))
			return nil, fmt.Errorf("failed to fetch versions dynamically: %w", err)
		}
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found from any source")
	}

	lg.Info("Successfully fetched MariaDB versions dynamically",
		logger.Int("version_count", len(versions)))

	// Sort versions
	sort.Slice(versions, func(i, j int) bool {
		return vc.compareVersions(versions[i].Version, versions[j].Version)
	})

	return versions, nil
}

// fetchVersionsFromOfficialSources attempts to fetch versions from multiple official MariaDB sources
func (vc *VersionChecker) fetchVersionsFromOfficialSources() ([]VersionInfo, error) {
	lg, _ := logger.Get()

	// Try different sources in order of preference
	sources := []func() ([]VersionInfo, error){
		vc.fetchFromMariaDBDownloadsPage,
		vc.fetchFromMariaDBAPI,
		vc.fetchVersionsFromScript,
	}

	for i, fetchFunc := range sources {
		lg.Debug("Trying to fetch versions from source", logger.Int("source_index", i))
		versions, err := fetchFunc()
		if err != nil {
			lg.Debug("Source failed, trying next", logger.Int("source_index", i), logger.Error(err))
			continue
		}
		if len(versions) > 0 {
			lg.Info("Successfully fetched versions from source", logger.Int("source_index", i), logger.Int("version_count", len(versions)))
			return versions, nil
		}
	}

	return nil, fmt.Errorf("all sources failed to provide versions")
}

// fetchFromMariaDBDownloadsPage scrapes versions from the official downloads page
func (vc *VersionChecker) fetchFromMariaDBDownloadsPage() ([]VersionInfo, error) {
	client := &http.Client{Timeout: vc.timeout}

	resp, err := client.Get("https://mariadb.org/download/")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch downloads page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return vc.parseVersionsFromDownloadsPage(string(body))
}

// parseVersionsFromDownloadsPage extracts version information from the downloads page
func (vc *VersionChecker) parseVersionsFromDownloadsPage(content string) ([]VersionInfo, error) {
	var versions []VersionInfo
	seenVersions := make(map[string]bool)

	// Look for version patterns in the downloads page
	// Pattern: "MariaDB X.Y" or "version X.Y.Z"
	versionRegex := regexp.MustCompile(`(?i)(?:mariadb|version)\s*(\d+\.\d+(?:\.\d+)?(?:-rc\d*|\.rolling)?)`)
	matches := versionRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			version := strings.TrimSpace(match[1])
			if !seenVersions[version] && vc.isValidVersion(version) {
				versionType := vc.determineVersionType(version)
				versions = append(versions, VersionInfo{
					Version: version,
					Type:    versionType,
				})
				seenVersions[version] = true
			}
		}
	}

	// Also look for more specific patterns
	stableVersionRegex := regexp.MustCompile(`(?i)stable.*?(\d+\.\d+(?:\.\d+)?)`)
	stableMatches := stableVersionRegex.FindAllStringSubmatch(content, -1)

	for _, match := range stableMatches {
		if len(match) > 1 {
			version := strings.TrimSpace(match[1])
			if !seenVersions[version] && vc.isValidVersion(version) {
				versions = append(versions, VersionInfo{
					Version: version,
					Type:    "stable",
				})
				seenVersions[version] = true
			}
		}
	}

	return versions, nil
}

// fetchFromMariaDBAPI attempts to fetch versions from MariaDB API endpoints
func (vc *VersionChecker) fetchFromMariaDBAPI() ([]VersionInfo, error) {
	// Try the GitHub releases API for MariaDB server
	client := &http.Client{Timeout: vc.timeout}

	resp, err := client.Get("https://api.github.com/repos/MariaDB/server/releases")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return vc.parseVersionsFromGitHubAPI(string(body))
}

// parseVersionsFromGitHubAPI extracts version information from GitHub API response
func (vc *VersionChecker) parseVersionsFromGitHubAPI(content string) ([]VersionInfo, error) {
	var versions []VersionInfo
	seenVersions := make(map[string]bool)

	// Look for tag names in the GitHub API response
	tagRegex := regexp.MustCompile(`"tag_name":\s*"([^"]*)"`)
	matches := tagRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			tag := strings.TrimSpace(match[1])
			// Extract version from tag (e.g., "mariadb-10.6.15" -> "10.6.15")
			versionMatch := regexp.MustCompile(`mariadb-(\d+\.\d+(?:\.\d+)?(?:-rc\d*)?)`).FindStringSubmatch(tag)
			if len(versionMatch) > 1 {
				version := versionMatch[1]
				if !seenVersions[version] && vc.isValidVersion(version) {
					versionType := vc.determineVersionType(version)
					versions = append(versions, VersionInfo{
						Version: version,
						Type:    versionType,
					})
					seenVersions[version] = true
				}
			}
		}
	}

	return versions, nil
}

// isValidVersion checks if a version string is valid for MariaDB
func (vc *VersionChecker) isValidVersion(version string) bool {
	// Basic validation for MariaDB versions
	matched, _ := regexp.MatchString(`^\d+\.\d+(?:\.\d+)?(?:-rc\d*|\.rolling)?$`, version)
	return matched
}

// determineVersionType determines the type of a version (stable, rc, rolling)
func (vc *VersionChecker) determineVersionType(version string) string {
	if strings.Contains(version, "rolling") {
		return "rolling"
	}
	if strings.Contains(version, "rc") {
		return "rc"
	}
	return "stable"
}

// fetchVersionsFromScript attempts to extract version info from the repository setup script
func (vc *VersionChecker) fetchVersionsFromScript() ([]VersionInfo, error) {
	client := &http.Client{Timeout: vc.timeout}

	resp, err := client.Get("https://r.mariadb.com/downloads/mariadb_repo_setup")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository script: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Extract version information from the script
	return vc.parseVersionsFromScript(string(body))
}

// parseVersionsFromScript extracts version information from the setup script content
func (vc *VersionChecker) parseVersionsFromScript(content string) ([]VersionInfo, error) {
	var versions []VersionInfo
	seenVersions := make(map[string]bool)

	// Look for mariadb version patterns in the script
	patterns := []string{
		`mariadb-(\d+\.\d+(?:\.\d+)?(?:-rc\d*)?(?:\.rolling)?)`,                                   // mariadb-10.6, mariadb-11.4
		`"(\d+\.\d+(?:\.\d+)?(?:-rc\d*)?(?:\.rolling)?)"`,                                         // quoted versions
		`--mariadb-server-version[=\s]+["\']?(\d+\.\d+(?:\.\d+)?(?:-rc\d*)?(?:\.rolling)?)["\']?`, // script parameters
		`version[=\s]*["\']?(\d+\.\d+(?:\.\d+)?(?:-rc\d*)?(?:\.rolling)?)["\']?`,                  // version parameters
	}

	for _, pattern := range patterns {
		versionRegex := regexp.MustCompile(pattern)
		matches := versionRegex.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) > 1 {
				version := strings.TrimSpace(match[1])
				if !seenVersions[version] && vc.isValidVersion(version) {
					versionType := vc.determineVersionType(version)
					versions = append(versions, VersionInfo{
						Version: version,
						Type:    versionType,
					})
					seenVersions[version] = true
				}
			}
		}
	}

	// Look for supported versions list in comments or documentation within the script
	supportedRegex := regexp.MustCompile(`(?i)(?:supported|available).*?versions?.*?:?\s*((?:\d+\.\d+(?:\.\d+)?(?:-rc\d*)?(?:\.rolling)?\s*,?\s*)+)`)
	supportedMatches := supportedRegex.FindAllStringSubmatch(content, -1)

	for _, match := range supportedMatches {
		if len(match) > 1 {
			versionList := match[1]
			// Extract individual versions from the list
			individualVersions := regexp.MustCompile(`(\d+\.\d+(?:\.\d+)?(?:-rc\d*)?(?:\.rolling)?)`).FindAllString(versionList, -1)
			for _, version := range individualVersions {
				version = strings.TrimSpace(version)
				if !seenVersions[version] && vc.isValidVersion(version) {
					versionType := vc.determineVersionType(version)
					versions = append(versions, VersionInfo{
						Version: version,
						Type:    versionType,
					})
					seenVersions[version] = true
				}
			}
		}
	}

	return versions, nil
}

// findCurrentStable finds the current stable version (typically the latest non-rolling, non-rc)
func (vc *VersionChecker) findCurrentStable(versions []VersionInfo) string {
	var stableVersions []string

	for _, v := range versions {
		if v.Type == "stable" && !strings.Contains(v.Version, "rolling") && !strings.Contains(v.Version, "rc") {
			stableVersions = append(stableVersions, v.Version)
		}
	}

	if len(stableVersions) == 0 {
		return ""
	}

	// Sort and return the latest stable
	sort.Slice(stableVersions, func(i, j int) bool {
		return vc.compareVersions(stableVersions[i], stableVersions[j])
	})

	return stableVersions[len(stableVersions)-1]
}

// findLatestVersion finds the absolute latest version (including rolling/rc)
func (vc *VersionChecker) findLatestVersion(versions []VersionInfo) string {
	if len(versions) == 0 {
		return ""
	}

	// Check for rolling first (it's usually the latest)
	for _, v := range versions {
		if strings.Contains(v.Version, "rolling") {
			return v.Version
		}
	}

	// Otherwise find the highest numbered version
	var allVersions []string
	for _, v := range versions {
		if !strings.Contains(v.Version, "rc") { // Exclude RC versions from "latest"
			allVersions = append(allVersions, v.Version)
		}
	}

	if len(allVersions) == 0 {
		return versions[len(versions)-1].Version
	}

	sort.Slice(allVersions, func(i, j int) bool {
		return vc.compareVersions(allVersions[i], allVersions[j])
	})

	return allVersions[len(allVersions)-1]
}

// findLatestMinor finds the latest minor version across all major versions
func (vc *VersionChecker) findLatestMinor(versions []VersionInfo) string {
	if len(versions) == 0 {
		return ""
	}

	// Group versions by major version
	majorVersions := make(map[string][]string)

	for _, v := range versions {
		if v.Type != "stable" || strings.Contains(v.Version, "rolling") || strings.Contains(v.Version, "rc") {
			continue // Only consider stable versions
		}

		// Extract major version (e.g., "10" from "10.5", "11" from "11.4")
		parts := strings.Split(v.Version, ".")
		if len(parts) >= 2 {
			major := parts[0]
			majorVersions[major] = append(majorVersions[major], v.Version)
		}
	}

	// Find the latest minor version for each major version
	var latestMinors []string
	for _, minorVersions := range majorVersions {
		if len(minorVersions) == 0 {
			continue
		}

		// Sort minor versions within this major version
		sort.Slice(minorVersions, func(i, j int) bool {
			return vc.compareVersions(minorVersions[i], minorVersions[j])
		})

		// Get the latest (last) minor version for this major
		latestMinor := minorVersions[len(minorVersions)-1]
		latestMinors = append(latestMinors, latestMinor)
	}

	if len(latestMinors) == 0 {
		return ""
	}

	// Sort all latest minor versions and return the absolute latest
	sort.Slice(latestMinors, func(i, j int) bool {
		return vc.compareVersions(latestMinors[i], latestMinors[j])
	})

	return latestMinors[len(latestMinors)-1]
}

// compareVersions compares two version strings, returns true if v1 < v2
func (vc *VersionChecker) compareVersions(v1, v2 string) bool {
	// Handle special cases
	if strings.Contains(v1, "rolling") {
		return false // rolling is always "latest"
	}
	if strings.Contains(v2, "rolling") {
		return true
	}
	if strings.Contains(v1, "rc") && !strings.Contains(v2, "rc") {
		return true // rc versions come before stable
	}
	if !strings.Contains(v1, "rc") && strings.Contains(v2, "rc") {
		return false
	}

	// Parse version numbers
	parts1 := strings.Split(strings.Replace(v1, "rc", "", -1), ".")
	parts2 := strings.Split(strings.Replace(v2, "rc", "", -1), ".")

	// Compare each part
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		if i < len(parts1) && parts1[i] != "" {
			fmt.Sscanf(parts1[i], "%d", &p1)
		}
		if i < len(parts2) && parts2[i] != "" {
			fmt.Sscanf(parts2[i], "%d", &p2)
		}

		if p1 < p2 {
			return true
		}
		if p1 > p2 {
			return false
		}
	}

	return false // versions are equal
}
