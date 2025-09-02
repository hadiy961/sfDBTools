package mariadb

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
)

// VersionInfo represents MariaDB version information
type VersionInfo struct {
	Version     string `json:"version"`
	Type        string `json:"type"` // stable, rc, rolling
	ReleaseDate string `json:"release_date,omitempty"`
	EOLDate     string `json:"eol_date,omitempty"`
}

// VersionFetcher provides interface for fetching MariaDB versions
type VersionFetcher interface {
	FetchVersions() ([]VersionInfo, error)
	GetName() string
}

// HTTPVersionFetcher fetches versions via HTTP
type HTTPVersionFetcher struct {
	URL       string
	Timeout   time.Duration
	UserAgent string
	Parser    VersionParser
}

// VersionParser provides interface for parsing version information
type VersionParser interface {
	ParseVersions(content string) ([]VersionInfo, error)
}

// NewHTTPVersionFetcher creates a new HTTP-based version fetcher
func NewHTTPVersionFetcher(url string, parser VersionParser) *HTTPVersionFetcher {
	return &HTTPVersionFetcher{
		URL:       url,
		Timeout:   30 * time.Second,
		UserAgent: "sfDBTools/1.0 MariaDB-Version-Checker",
		Parser:    parser,
	}
}

// FetchVersions implements VersionFetcher interface
func (f *HTTPVersionFetcher) FetchVersions() ([]VersionInfo, error) {
	client := &http.Client{Timeout: f.Timeout}

	req, err := http.NewRequest("GET", f.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", f.UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from %s: %w", f.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, f.URL)
	}

	// Read response body safely
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return f.Parser.ParseVersions(string(body))
}

// GetName implements VersionFetcher interface
func (f *HTTPVersionFetcher) GetName() string {
	return fmt.Sprintf("HTTP Fetcher (%s)", f.URL)
}

// IsValidVersion checks if a version string is valid for MariaDB
func IsValidVersion(version string) bool {
	matched, _ := regexp.MatchString(`^\d+\.\d+(?:\.\d+)?(?:-rc\d*|\.rolling)?$`, version)
	return matched
}

// DetermineVersionType determines the type of a version (stable, rc, rolling)
func DetermineVersionType(version string) string {
	if strings.Contains(version, "rolling") {
		return "rolling"
	}
	if strings.Contains(version, "rc") {
		return "rc"
	}
	return "stable"
}

// GetVersionsForOS returns MariaDB versions available for the specified OS
func GetVersionsForOS(osInfo *common.OSInfo, fetchers []VersionFetcher) ([]VersionInfo, error) {
	lg, _ := logger.Get()

	lg.Info("Fetching MariaDB versions for OS",
		logger.String("os_id", osInfo.ID),
		logger.String("os_version", osInfo.Version),
		logger.String("package_type", osInfo.PackageType))

	var allVersions []VersionInfo
	seenVersions := make(map[string]bool)

	for _, fetcher := range fetchers {
		lg.Debug("Trying fetcher", logger.String("fetcher", fetcher.GetName()))

		versions, err := fetcher.FetchVersions()
		if err != nil {
			lg.Warn("Fetcher failed",
				logger.String("fetcher", fetcher.GetName()),
				logger.Error(err))
			continue
		}

		// Filter versions based on OS compatibility
		for _, version := range versions {
			if !seenVersions[version.Version] && isVersionCompatibleWithOS(version, osInfo) {
				allVersions = append(allVersions, version)
				seenVersions[version.Version] = true
			}
		}

		if len(versions) > 0 {
			lg.Info("Successfully fetched versions",
				logger.String("fetcher", fetcher.GetName()),
				logger.Int("version_count", len(versions)))
		}
	}

	if len(allVersions) == 0 {
		return nil, fmt.Errorf("no compatible versions found for OS %s %s", osInfo.ID, osInfo.Version)
	}

	return allVersions, nil
}

// isVersionCompatibleWithOS checks if a version is compatible with the given OS
func isVersionCompatibleWithOS(version VersionInfo, osInfo *common.OSInfo) bool {
	// All stable versions are generally compatible
	// This can be enhanced with more specific OS/version compatibility rules

	// Basic compatibility rules
	switch osInfo.ID {
	case "ubuntu", "debian":
		return osInfo.PackageType == "deb"
	case "centos", "rhel", "rocky", "almalinux":
		return osInfo.PackageType == "rpm"
	case "sles":
		return osInfo.PackageType == "rpm"
	default:
		// For unknown OS, assume compatibility
		return true
	}
}

// CompareVersions compares two version strings, returns true if v1 < v2
func CompareVersions(v1, v2 string) bool {
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

// GetMariaDBEOLDate returns EOL date for MariaDB version using dynamic approach
func GetMariaDBEOLDate(version string) string {
	// For rolling/rc versions, return "No LTS"
	if strings.Contains(version, "rolling") || strings.Contains(version, "rc") {
		return "No LTS"
	}

	// For testing, skip external API and use local calculation only
	// TODO: Re-enable external API later
	if eolDate := tryFetchEOLFromExternal(version); eolDate != "" {
		return eolDate
	}

	// Fall back to calculation based on lifecycle
	return calculateEOLFromLifecycle(version)
}

// tryFetchEOLFromExternal attempts to fetch EOL from external sources
func tryFetchEOLFromExternal(version string) string {
	// Extract major.minor version
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return ""
	}
	majorMinor := parts[0] + "." + parts[1]

	// Try endoflife.date API with timeout
	client := &http.Client{Timeout: 3 * time.Second}
	url := fmt.Sprintf("https://endoflife.date/api/mariadb/%s.json", majorMinor)

	resp, err := client.Get(url)
	if err != nil {
		return "" // Fail silently for external sources
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	// Simple regex parsing to extract EOL date
	eolPattern := regexp.MustCompile(`"eol"\s*:\s*"([0-9]{4}-[0-9]{2}-[0-9]{2})"`)
	matches := eolPattern.FindStringSubmatch(string(body))

	if len(matches) > 1 {
		return matches[1]
	}

	// Check for boolean EOL (means already EOL)
	eolBoolPattern := regexp.MustCompile(`"eol"\s*:\s*true`)
	if eolBoolPattern.MatchString(string(body)) {
		return time.Now().AddDate(0, 0, -1).Format("2006-01-02") // Yesterday = already EOL
	}

	return ""
}

// calculateEOLFromLifecycle calculates EOL based on MariaDB lifecycle policy
func calculateEOLFromLifecycle(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return "TBD"
	}

	major := parts[0]
	minor := parts[1]

	// Determine if LTS based on known pattern
	isLTS := isLTSVersion(major, minor)

	if isLTS {
		return estimateLTSEOL(major, minor)
	}

	return estimateStableEOL(major, minor)
}

// isLTSVersion determines if a version is LTS based on MariaDB's pattern
func isLTSVersion(major, minor string) bool {
	// Known LTS pattern: typically every 1.5-2 years
	ltsVersions := map[string][]string{
		"10": {"5", "6", "11"}, // Known LTS
		"11": {"4"},            // Known LTS
		// Future pattern: likely 12.4, 13.4, etc.
	}

	if minors, exists := ltsVersions[major]; exists {
		for _, ltsMinor := range minors {
			if minor == ltsMinor {
				return true
			}
		}
	}

	// Pattern-based detection for future versions
	if majorNum, err := parseVersionNumber(major); err == nil && majorNum >= 12 {
		if minorNum, err := parseVersionNumber(minor); err == nil && minorNum == 4 {
			return true // Assume X.4 pattern continues
		}
	}

	return false
}

// estimateLTSEOL estimates EOL for LTS versions
func estimateLTSEOL(major, minor string) string {
	// Try to get release date and add 5 years
	if releaseDate := estimateReleaseDate(major, minor); releaseDate != "" {
		if releaseTime, err := time.Parse("2006-01-02", releaseDate); err == nil {
			eolTime := releaseTime.AddDate(5, 0, 0) // LTS = 5 years support
			return eolTime.Format("2006-01-02")
		}
	}

	// Fallback: conservative estimate
	return "TBD"
}

// estimateStableEOL estimates EOL for stable versions
func estimateStableEOL(major, minor string) string {
	// Stable versions typically get 18 months support
	if releaseDate := estimateReleaseDate(major, minor); releaseDate != "" {
		if releaseTime, err := time.Parse("2006-01-02", releaseDate); err == nil {
			eolTime := releaseTime.AddDate(1, 6, 0) // 18 months support
			return eolTime.Format("2006-01-02")
		}
	}

	return "TBD"
}

// estimateReleaseDate estimates release date based on version pattern
func estimateReleaseDate(major, minor string) string {
	// Known release dates for reference
	knownReleases := map[string]string{
		"10.5":  "2020-06-24",
		"10.6":  "2021-07-06",
		"10.11": "2023-02-16",
		"11.4":  "2024-05-29",
	}

	versionKey := major + "." + minor
	if date, exists := knownReleases[versionKey]; exists {
		return date
	}

	// Pattern-based estimation for unknown versions
	majorNum, err1 := parseVersionNumber(major)
	minorNum, err2 := parseVersionNumber(minor)

	if err1 != nil || err2 != nil {
		return ""
	}

	// Estimate based on release pattern
	if majorNum >= 11 {
		// MariaDB 11+ typically releases annually with minor releases quarterly
		baseYear := 2024 + (majorNum - 11)
		estimatedMonth := 2 + (minorNum * 3) // Quarterly releases starting Feb
		if estimatedMonth > 12 {
			baseYear++
			estimatedMonth = estimatedMonth - 12
		}

		estimated := time.Date(baseYear, time.Month(estimatedMonth), 15, 0, 0, 0, 0, time.UTC)
		if estimated.Before(time.Now().AddDate(5, 0, 0)) { // Reasonable future limit
			return estimated.Format("2006-01-02")
		}
	}

	return ""
}

// parseVersionNumber safely parses version number string to int
func parseVersionNumber(versionStr string) (int, error) {
	var num int
	_, err := fmt.Sscanf(versionStr, "%d", &num)
	return num, err
}
