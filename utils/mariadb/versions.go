package mariadb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
)

// MariaDBVersions holds all supported MariaDB versions organized by series
type MariaDBVersions struct {
	StableVersions map[string][]string `json:"stable_versions"`
	OtherVersions  map[string][]string `json:"other_versions"`
	AllVersions    []string            `json:"all_versions"`
	LastUpdated    time.Time           `json:"last_updated"`
	Source         string              `json:"source"` // "api", "cache", "fallback"
}

// Version cache
var (
	versionCache    *MariaDBVersions
	cacheExpiration = 6 * time.Hour // Cache valid for 6 hours
	lastCacheUpdate time.Time
)

// MariaDBRelease represents a MariaDB release from API
type MariaDBRelease struct {
	Version  string `json:"release_id"`
	Date     string `json:"release_date"`
	Status   string `json:"release_status"`
	Maturity string `json:"maturity"`
}

// MariaDBAPIResponse represents the API response structure for major releases
type MariaDBAPIResponse struct {
	MajorReleases []MariaDBMajorRelease `json:"major_releases"`
}

// MariaDBMajorRelease represents a major release from API
type MariaDBMajorRelease struct {
	ReleaseID     string `json:"release_id"`
	ReleaseName   string `json:"release_name"`
	ReleaseStatus string `json:"release_status"`
}

// GetSupportedVersions returns all supported MariaDB versions
func GetSupportedVersions() *MariaDBVersions {
	return GetSupportedVersionsWithConnectivityCheck(true)
}

// GetSupportedVersionsWithConnectivityCheck returns all supported MariaDB versions with optional connectivity check
func GetSupportedVersionsWithConnectivityCheck(requireConnectivity bool) *MariaDBVersions {
	lg, _ := logger.Get()

	// Check if cache is still valid
	if versionCache != nil && time.Since(lastCacheUpdate) < cacheExpiration {
		lg.Debug("Using cached MariaDB versions",
			logger.String("source", versionCache.Source),
			logger.String("last_updated", versionCache.LastUpdated.Format("2006-01-02 15:04:05")))
		return versionCache
	}

	// Check internet connectivity only if required
	if requireConnectivity {
		lg.Info("Checking internet connectivity for MariaDB operations")
		if err := common.RequireInternetForOperation("MariaDB operations"); err != nil {
			lg.Debug("Internet connectivity not available, using fallback options", logger.Error(err))
			// Return cached data if available, even if expired
			if versionCache != nil {
				lg.Debug("Using expired cache due to connectivity issues")
				return versionCache
			}
			// Return empty structure as last resort
			lg.Error("No cached data available and no internet connectivity")
			return &MariaDBVersions{
				StableVersions: make(map[string][]string),
				OtherVersions:  make(map[string][]string),
				AllVersions:    []string{},
				LastUpdated:    time.Now(),
				Source:         "failed",
			}
		}
	}

	// Try to fetch from API (connectivity is assumed to be already verified)
	if versions := fetchVersionsFromAPI(); versions != nil {
		versionCache = versions
		lastCacheUpdate = time.Now()
		lg.Info("Fetched MariaDB versions from API",
			logger.Int("total_versions", len(versions.AllVersions)))
		return versions
	}

	// If API fetch failed, return cached data if available
	if versionCache != nil {
		lg.Debug("API fetch failed, using cached data")
		return versionCache
	}

	// If all methods fail, return empty structure
	lg.Error("Failed to fetch MariaDB versions from all sources")
	return &MariaDBVersions{
		StableVersions: make(map[string][]string),
		OtherVersions:  make(map[string][]string),
		AllVersions:    []string{},
		LastUpdated:    time.Now(),
		Source:         "failed",
	}
}

// fetchVersionsFromAPI fetches version list from MariaDB API using two-step approach
func fetchVersionsFromAPI() *MariaDBVersions {
	lg, _ := logger.Get()

	// Step 1: Get major releases
	majorReleases := fetchMajorReleases()
	if len(majorReleases) == 0 {
		lg.Debug("No major releases found, trying fallback methods")

		// Try scraping from downloads page as backup
		if versions := scrapeVersionsFromDownloadsPage(); versions != nil {
			lg.Debug("Successfully scraped versions from downloads page")
			return versions
		}

		// Last resort: GitHub releases API
		if versions := fetchFromGitHubAPI(); versions != nil {
			lg.Debug("Successfully fetched from GitHub API")
			return versions
		}

		return nil
	}

	// Step 2: Get point releases for each major release
	var allVersions []string
	versionSet := make(map[string]bool)

	for _, majorRelease := range majorReleases {
		pointVersions := fetchPointReleases(majorRelease.ReleaseID)
		for _, version := range pointVersions {
			if !versionSet[version] && isValidVersionFormat(version) {
				allVersions = append(allVersions, version)
				versionSet[version] = true
			}
		}
	}

	if len(allVersions) == 0 {
		lg.Debug("No point releases found from API, trying fallback methods")

		// Try scraping from downloads page as backup
		if versions := scrapeVersionsFromDownloadsPage(); versions != nil {
			lg.Debug("Successfully scraped versions from downloads page")
			return versions
		}

		// Last resort: GitHub releases API
		if versions := fetchFromGitHubAPI(); versions != nil {
			lg.Debug("Successfully fetched from GitHub API")
			return versions
		}

		return nil
	}

	lg.Info("Successfully fetched versions from API", logger.Int("count", len(allVersions)))
	return organizeVersions(allVersions, "api")
}

// fetchMajorReleases gets the list of major releases
func fetchMajorReleases() []MariaDBMajorRelease {
	lg, _ := logger.Get()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	apiURL := "https://downloads.mariadb.org/rest-api/mariadb/"
	resp, err := client.Get(apiURL)
	if err != nil {
		lg.Debug("Failed to fetch major releases", logger.Error(err))
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		lg.Debug("Major releases API returned non-200 status", logger.Int("status", resp.StatusCode))
		return nil
	}

	var apiResponse MariaDBAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		lg.Debug("Failed to parse major releases response", logger.Error(err))
		return nil
	}

	return apiResponse.MajorReleases
}

// fetchPointReleases gets point releases for a specific major release
func fetchPointReleases(majorRelease string) []string {
	lg, _ := logger.Get()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	apiURL := fmt.Sprintf("https://downloads.mariadb.org/rest-api/mariadb/%s/", majorRelease)
	resp, err := client.Get(apiURL)
	if err != nil {
		lg.Debug("Failed to fetch point releases", logger.String("major", majorRelease), logger.Error(err))
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		lg.Debug("Point releases API returned non-200 status",
			logger.String("major", majorRelease),
			logger.Int("status", resp.StatusCode))
		return nil
	}

	var response struct {
		Releases map[string]interface{} `json:"releases"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		lg.Debug("Failed to parse point releases response", logger.String("major", majorRelease), logger.Error(err))
		return nil
	}

	var versions []string
	for versionKey := range response.Releases {
		versions = append(versions, versionKey)
	}

	return versions
}

// scrapeVersionsFromDownloadsPage scrapes version info from MariaDB downloads page
func scrapeVersionsFromDownloadsPage() *MariaDBVersions {
	lg, _ := logger.Get()

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Get("https://downloads.mariadb.org/mariadb/")
	if err != nil {
		lg.Debug("Failed to scrape downloads page", logger.Error(err))
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	return parseVersionsFromHTML(string(body))
}

// parseVersionsFromHTML parses version numbers from HTML content
func parseVersionsFromHTML(html string) *MariaDBVersions {
	lg, _ := logger.Get()

	// Look specifically for MariaDB version patterns
	// This regex matches versions like 10.6.23, 11.4.8, 12.0.2
	versionRegex := regexp.MustCompile(`(?i)mariadb[^\d]*(\d+\.\d+\.\d+)`)
	matches := versionRegex.FindAllStringSubmatch(html, -1)

	// Fallback: look for standalone version numbers in typical MariaDB format
	if len(matches) == 0 {
		lg.Debug("No MariaDB-specific versions found, trying general pattern")
		versionRegex = regexp.MustCompile(`\b(1[0-9]\.\d+\.\d+)\b`)
		matches = versionRegex.FindAllStringSubmatch(html, -1)
	}

	if len(matches) == 0 {
		lg.Debug("No versions found in HTML content")
		return nil
	}

	var versions []string
	versionSet := make(map[string]bool)

	for _, match := range matches {
		version := match[1]
		if !versionSet[version] && isValidVersionFormat(version) {
			versions = append(versions, version)
			versionSet[version] = true
		}
	}

	if len(versions) == 0 {
		lg.Debug("No valid MariaDB versions found after filtering")
		return nil
	}

	lg.Info("Parsed versions from HTML", logger.Int("count", len(versions)))

	// Sort versions
	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i], versions[j]) < 0
	})

	return organizeVersions(versions, "html")
}

// fetchFromGitHubAPI fetches MariaDB versions from GitHub releases API
func fetchFromGitHubAPI() *MariaDBVersions {
	lg, _ := logger.Get()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	apiURL := "https://api.github.com/repos/MariaDB/server/releases"
	resp, err := client.Get(apiURL)
	if err != nil {
		lg.Debug("GitHub API request failed", logger.Error(err))
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		lg.Debug("GitHub API returned non-200 status", logger.Int("status", resp.StatusCode))
		return nil
	}

	var releases []struct {
		TagName string `json:"tag_name"`
		Name    string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		lg.Debug("Failed to parse GitHub API response", logger.Error(err))
		return nil
	}

	var versions []string
	versionSet := make(map[string]bool)

	for _, release := range releases {
		// Extract version from tag name (e.g., "mariadb-10.6.23" -> "10.6.23")
		version := strings.TrimPrefix(release.TagName, "mariadb-")
		if !versionSet[version] && isValidVersionFormat(version) {
			versions = append(versions, version)
			versionSet[version] = true
		}
	}

	if len(versions) == 0 {
		lg.Debug("No valid versions found in GitHub releases")
		return nil
	}

	lg.Info("Fetched versions from GitHub API", logger.Int("count", len(versions)))
	return organizeVersions(versions, "github")
}

// organizeVersions organizes versions into stable and other categories
func organizeVersions(versions []string, source string) *MariaDBVersions {
	stableVersions := make(map[string][]string)
	otherVersions := make(map[string][]string)

	// Dynamically determine stable series based on version patterns
	seriesCount := make(map[string]int)
	seriesVersions := make(map[string][]string)

	// First pass: count versions per series and collect them
	for _, version := range versions {
		series := GetVersionSeries(version)
		if series == "" {
			continue
		}
		seriesCount[series]++
		seriesVersions[series] = append(seriesVersions[series], version)
	}

	// Determine stable series based on multiple criteria
	stableSeries := make(map[string]bool)

	// Calculate average versions per series to determine threshold
	totalSeries := len(seriesCount)
	if totalSeries > 0 {
		totalVersions := 0
		for _, count := range seriesCount {
			totalVersions += count
		}
		avgVersionsPerSeries := float64(totalVersions) / float64(totalSeries)

		// Dynamic threshold: series with version count above average are likely stable/LTS
		dynamicThreshold := int(avgVersionsPerSeries * 1.2) // 20% above average
		if dynamicThreshold < 3 {
			dynamicThreshold = 3 // Minimum threshold
		}

		for series, count := range seriesCount {
			parts := strings.Split(series, ".")
			if len(parts) == 2 {
				_, majorErr := strconv.Atoi(parts[0])
				minor, minorErr := strconv.Atoi(parts[1])

				if majorErr == nil && minorErr == nil {
					// Consider as stable if:
					// 1. Has many versions (above dynamic threshold)
					// 2. Even minor versions (traditional LTS pattern for many projects)
					// 3. Recent major versions with reasonable version count
					isStable := count >= dynamicThreshold ||
						(minor%2 == 0 && count >= 3) ||
						(count >= 3) // Any series with reasonable version count

					if isStable {
						stableSeries[series] = true
					}
				}
			}
		}
	}

	// Second pass: categorize versions
	for _, version := range versions {
		series := GetVersionSeries(version)
		if series == "" {
			continue
		}

		if stableSeries[series] {
			stableVersions[series] = append(stableVersions[series], version)
		} else {
			otherVersions[series] = append(otherVersions[series], version)
		}
	}

	return &MariaDBVersions{
		StableVersions: stableVersions,
		OtherVersions:  otherVersions,
		AllVersions:    versions,
		LastUpdated:    time.Now(),
		Source:         source,
	}
}

// isValidVersionFormat checks if version string is in valid format
func isValidVersionFormat(version string) bool {
	// Match pattern like 10.6.22, 11.4.3, etc.
	matched, _ := regexp.MatchString(`^\d+\.\d+\.\d+$`, version)
	if !matched {
		return false
	}

	// Additional validation: ensure all parts are valid numbers
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return false
	}

	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}

	return true
}

// compareVersions compares two version strings
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		num1, _ := strconv.Atoi(parts1[i])
		num2, _ := strconv.Atoi(parts2[i])

		if num1 < num2 {
			return -1
		}
		if num1 > num2 {
			return 1
		}
	}

	return len(parts1) - len(parts2)
}

// RefreshVersionCache forces a refresh of the version cache
func RefreshVersionCache() *MariaDBVersions {
	lg, _ := logger.Get()

	lg.Info("Forcing refresh of MariaDB version cache")

	// Clear existing cache
	versionCache = nil
	lastCacheUpdate = time.Time{}

	// Fetch fresh data
	return GetSupportedVersions()
}

// GetCacheInfo returns information about the current cache
func GetCacheInfo() (source string, lastUpdated time.Time, isExpired bool) {
	if versionCache == nil {
		return "none", time.Time{}, true
	}

	return versionCache.Source, versionCache.LastUpdated, time.Since(lastCacheUpdate) >= cacheExpiration
}

// IsValidVersion checks if the provided version is supported
func IsValidVersion(version string) bool {
	return IsValidVersionWithConnectivityCheck(version, true)
}

// IsValidVersionWithConnectivityCheck checks if the provided version is supported with optional connectivity check
func IsValidVersionWithConnectivityCheck(version string, requireConnectivity bool) bool {
	versions := GetSupportedVersionsWithConnectivityCheck(requireConnectivity)
	for _, v := range versions.AllVersions {
		if v == version {
			return true
		}
	}
	return false
}

// GetLatestVersion returns the latest stable version
func GetLatestVersion() string {
	versions := GetSupportedVersions()

	// Get the latest version from the most stable series
	var latestVersion string

	// First try to get from stable versions
	for _, versionList := range versions.StableVersions {
		if len(versionList) > 0 {
			lastVersion := versionList[len(versionList)-1]
			if latestVersion == "" || compareVersions(lastVersion, latestVersion) > 0 {
				latestVersion = lastVersion
			}
		}
	}

	// If no stable versions found, get from other versions
	if latestVersion == "" {
		for _, versionList := range versions.OtherVersions {
			if len(versionList) > 0 {
				lastVersion := versionList[len(versionList)-1]
				if latestVersion == "" || compareVersions(lastVersion, latestVersion) > 0 {
					latestVersion = lastVersion
				}
			}
		}
	}

	// If still no version found, get from all versions
	if latestVersion == "" && len(versions.AllVersions) > 0 {
		// AllVersions should be sorted, so get the last one
		latestVersion = versions.AllVersions[len(versions.AllVersions)-1]
	}

	return latestVersion
}

// GetLatestVersionWithCache returns the latest stable version using provided cache
func GetLatestVersionWithCache(versions *MariaDBVersions) string {
	// Get the latest version from the most stable series
	var latestVersion string

	// First try to get from stable versions
	for _, versionList := range versions.StableVersions {
		if len(versionList) > 0 {
			lastVersion := versionList[len(versionList)-1]
			if latestVersion == "" || compareVersions(lastVersion, latestVersion) > 0 {
				latestVersion = lastVersion
			}
		}
	}

	// If no stable versions found, get from other versions
	if latestVersion == "" {
		for _, versionList := range versions.OtherVersions {
			if len(versionList) > 0 {
				lastVersion := versionList[len(versionList)-1]
				if latestVersion == "" || compareVersions(lastVersion, latestVersion) > 0 {
					latestVersion = lastVersion
				}
			}
		}
	}

	// If still no version found, get from all versions
	if latestVersion == "" && len(versions.AllVersions) > 0 {
		// AllVersions should be sorted, so get the last one
		latestVersion = versions.AllVersions[len(versions.AllVersions)-1]
	}

	return latestVersion
}

// GetRecommendedVersion returns recommended version based on OS
func GetRecommendedVersion(osInfo *OSInfo) string {
	lg, _ := logger.Get()
	versions := GetSupportedVersions()

	lg.Debug("Getting recommended version for OS",
		logger.String("os", osInfo.ID),
		logger.String("version", osInfo.Version))

	// Get the latest LTS/stable version as the primary recommendation
	latestStable := getLatestStableVersion(versions)
	if latestStable != "" {
		lg.Debug("Selected latest stable version",
			logger.String("version", latestStable))
		return latestStable
	}

	// Fallback to general latest version
	latest := GetLatestVersion()
	lg.Debug("Fallback to latest available version",
		logger.String("version", latest))
	return latest
}

// getLatestStableVersion returns the latest version from stable series
func getLatestStableVersion(versions *MariaDBVersions) string {
	var latestVersion string
	var latestSeries string

	// Find the highest stable series and get its latest version
	for series, versionList := range versions.StableVersions {
		if len(versionList) > 0 {
			lastVersion := versionList[len(versionList)-1]
			if latestSeries == "" || compareVersions(series+".0", latestSeries+".0") > 0 {
				latestSeries = series
				latestVersion = lastVersion
			}
		}
	}

	return latestVersion
}

// GetVersionSeries returns the series (e.g., "10.6") for a given version
func GetVersionSeries(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		return fmt.Sprintf("%s.%s", parts[0], parts[1])
	}
	return ""
}

// IsStableVersion checks if the version is from a stable series
func IsStableVersion(version string) bool {
	versions := GetSupportedVersions()
	series := GetVersionSeries(version)

	if stableVersions, exists := versions.StableVersions[series]; exists {
		for _, v := range stableVersions {
			if v == version {
				return true
			}
		}
	}
	return false
}

// ValidateVersionForOS validates if version is compatible with OS
func ValidateVersionForOS(version string, osInfo *OSInfo) error {
	if !IsValidVersion(version) {
		return fmt.Errorf("unsupported version: %s", version)
	}

	lg, _ := logger.Get()
	lg.Debug("Validating version compatibility",
		logger.String("version", version),
		logger.String("os", osInfo.ID),
		logger.String("os_version", osInfo.Version))

	// Check OS compatibility for specific versions
	majorVersion := GetVersionSeries(version)

	// CentOS 10 has limited support for MariaDB versions due to dependency conflicts
	// Only show warning for official MariaDB repository, native repo might work
	if osInfo.ID == "centos" && osInfo.Version == "10" {
		switch majorVersion {
		case "10.5", "10.6", "10.7", "10.8", "10.9", "10.10", "10.11":
			lg.Warn("MariaDB official repository may have dependency conflicts on CentOS Stream 10. Will try native repository instead.")
		}
	}

	// All validation passed - version is compatible
	return nil
}

// GetAvailableVersionsForSeries returns all versions for a specific series
func GetAvailableVersionsForSeries(series string) []string {
	versions := GetSupportedVersions()

	if stableVersions, exists := versions.StableVersions[series]; exists {
		return stableVersions
	}

	if otherVersions, exists := versions.OtherVersions[series]; exists {
		return otherVersions
	}

	return []string{}
}

// ListAllSeries returns all available series
func ListAllSeries() []string {
	versions := GetSupportedVersions()
	var series []string

	for s := range versions.StableVersions {
		series = append(series, s)
	}
	for s := range versions.OtherVersions {
		series = append(series, s)
	}

	return series
}

// GetVersionSeriesInfo returns dynamic information about version series
func GetVersionSeriesInfo() map[string]string {
	versions := GetSupportedVersions()
	return GetVersionSeriesInfoWithCache(versions)
}

// GetVersionSeriesInfoWithCache returns dynamic information about version series using provided cache
func GetVersionSeriesInfoWithCache(versions *MariaDBVersions) map[string]string {
	info := make(map[string]string)

	// Analyze stable series to provide dynamic descriptions
	for series := range versions.StableVersions {
		switch {
		case strings.HasPrefix(series, "10.6"):
			info[series] = "Current stable LTS series"
		case strings.HasPrefix(series, "10.11"):
			info[series] = "Long-term support series"
		case strings.HasPrefix(series, "11."):
			info[series] = "Feature series with latest capabilities"
		case strings.HasPrefix(series, "12."):
			info[series] = "Latest stable series"
		default:
			info[series] = "Stable series"
		}
	}

	return info
}

// GetLTSVersions returns versions that are considered LTS
func GetLTSVersions() []string {
	versions := GetSupportedVersions()
	var ltsVersions []string

	// LTS series are typically 10.6.x and 10.11.x
	for series, versionList := range versions.StableVersions {
		if strings.HasPrefix(series, "10.6") || strings.HasPrefix(series, "10.11") {
			if len(versionList) > 0 {
				// Get the latest from each LTS series
				ltsVersions = append(ltsVersions, versionList[len(versionList)-1])
			}
		}
	}

	return ltsVersions
}

// GetMostStableSeries returns the most stable series based on version count
func GetMostStableSeries() string {
	versions := GetSupportedVersions()
	maxCount := 0
	mostStableSeries := ""

	for series, versionList := range versions.StableVersions {
		if len(versionList) > maxCount {
			maxCount = len(versionList)
			mostStableSeries = series
		}
	}

	return mostStableSeries
}

// GetOSSpecificRecommendation returns OS-specific recommendations dynamically
func GetOSSpecificRecommendation(osInfo *OSInfo) map[string]string {
	recommendations := make(map[string]string)

	switch osInfo.ID {
	case "centos", "rocky", "alma":
		if osInfo.Version >= "10" {
			recommendations["primary"] = "Native repository (10.11.x series)"
			recommendations["reason"] = "Better dependency compatibility"
			recommendations["alternative"] = "External MariaDB repository for latest features"
		} else {
			recommendations["primary"] = "MariaDB official repository"
			recommendations["reason"] = "More recent versions available"
		}
	case "ubuntu":
		recommendations["primary"] = "MariaDB official repository"
		recommendations["reason"] = "Latest features and faster updates"
	case "debian":
		recommendations["primary"] = "Native repository or MariaDB official"
		recommendations["reason"] = "Both provide good stability"
	default:
		recommendations["primary"] = "MariaDB official repository"
		recommendations["reason"] = "Universal compatibility"
	}

	return recommendations
}
