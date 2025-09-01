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
