package check_version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"sfDBTools/internal/logger"
)

// VersionService handles MariaDB version operations
type VersionService struct {
	config *CheckVersionConfig
	client *http.Client
}

// NewVersionService creates a new version service with the given configuration
func NewVersionService(config *CheckVersionConfig) *VersionService {
	if config == nil {
		config = DefaultCheckVersionConfig()
	}

	return &VersionService{
		config: config,
		client: &http.Client{
			Timeout: config.APITimeout,
		},
	}
}

// FetchAvailableVersions retrieves MariaDB version information from the API
func (s *VersionService) FetchAvailableVersions() ([]VersionInfo, error) {
	lg, _ := logger.Get()

	// Get major releases first
	resp, err := s.client.Get(s.config.APIBaseURL)
	if err != nil {
		lg.Error("Failed to fetch MariaDB API", logger.Error(err))
		return nil, fmt.Errorf("failed to fetch MariaDB versions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		lg.Error("MariaDB API returned non-200 status", logger.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		lg.Error("Failed to read API response", logger.Error(err))
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	var apiResponse APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		lg.Error("Failed to parse API response", logger.Error(err))
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Filter and process versions
	versionInfos := s.processReleases(apiResponse.MajorReleases)

	lg.Info("Successfully processed MariaDB versions", logger.Int("count", len(versionInfos)))
	return versionInfos, nil
}

// processReleases filters and processes the releases from API response
func (s *VersionService) processReleases(releases []Release) []VersionInfo {
	lg, _ := logger.Get()
	var versionInfos []VersionInfo

	for _, release := range releases {
		// Only include versions meeting our criteria
		if !s.isVersionEligible(release.ID, release.Status, release.SupportType) {
			continue
		}

		// Get the latest minor version for this major version
		latestVersion, err := s.getLatestMinorVersion(release.ID)
		if err != nil {
			lg.Warn("Failed to get latest minor version",
				logger.String("major_version", release.ID),
				logger.Error(err))
			latestVersion = release.ID + ".0" // fallback
		}

		versionInfo := VersionInfo{
			Version:       release.ID,
			EOL:           s.formatEOLDate(release.EOLDate),
			LatestVersion: latestVersion,
			SupportType:   release.SupportType,
		}

		versionInfos = append(versionInfos, versionInfo)
	}

	// Sort versions in descending order (newest first)
	sort.Slice(versionInfos, func(i, j int) bool {
		return s.compareVersions(versionInfos[i].Version, versionInfos[j].Version)
	})

	return versionInfos
}

// isVersionEligible checks if a version should be included in the results
func (s *VersionService) isVersionEligible(versionID, status, supportType string) bool {
	// Only include stable releases
	if status != "Stable" {
		return false
	}

	// Only include Long Term Support versions
	if supportType != "Long Term Support" {
		return false
	}

	// Parse version to check if it meets minimum version requirement
	return s.isVersionAtLeastMinimum(versionID)
}

// isVersionAtLeastMinimum checks if version meets minimum requirement
func (s *VersionService) isVersionAtLeastMinimum(versionID string) bool {
	minParts := strings.Split(s.config.MinimumVersion, ".")
	versionParts := strings.Split(versionID, ".")

	if len(versionParts) < 2 || len(minParts) < 2 {
		return false
	}

	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return false
	}

	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return false
	}

	minMajor, err := strconv.Atoi(minParts[0])
	if err != nil {
		return false
	}

	minMinor, err := strconv.Atoi(minParts[1])
	if err != nil {
		return false
	}

	return major > minMajor || (major == minMajor && minor >= minMinor)
}

// getLatestMinorVersion fetches the latest minor version for a major version
func (s *VersionService) getLatestMinorVersion(majorVersion string) (string, error) {
	lg, _ := logger.Get()

	apiURL := fmt.Sprintf("%s%s/", s.config.APIBaseURL, majorVersion)

	resp, err := s.client.Get(apiURL)
	if err != nil {
		lg.Error("Failed to fetch version details",
			logger.String("version", majorVersion),
			logger.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var versionDetails VersionDetails
	if err := json.Unmarshal(body, &versionDetails); err != nil {
		return "", err
	}

	// Find the latest version
	var versions []string
	for version := range versionDetails.Releases {
		versions = append(versions, version)
	}

	if len(versions) == 0 {
		return majorVersion + ".0", nil
	}

	// Sort versions and return the latest
	sort.Slice(versions, func(i, j int) bool {
		return s.compareVersions(versions[i], versions[j])
	})

	return versions[0], nil
}

// compareVersions compares two version strings, returns true if v1 > v2
func (s *VersionService) compareVersions(v1, v2 string) bool {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Pad with zeros if needed
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int

		if i < len(parts1) {
			p1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			p2, _ = strconv.Atoi(parts2[i])
		}

		if p1 > p2 {
			return true
		} else if p1 < p2 {
			return false
		}
	}

	return false // versions are equal
}

// formatEOLDate formats the EOL date for display
func (s *VersionService) formatEOLDate(eolDate string) string {
	if eolDate == "" || eolDate == "null" {
		return "No EOL Date"
	}

	// Parse the date and format it nicely
	if t, err := time.Parse("2006-01-02", eolDate); err == nil {
		return t.Format("2 January 2006")
	}

	return eolDate // Return as-is if parsing fails
}
