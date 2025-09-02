package mariadb

import (
	"sfDBTools/utils/common"
	"sfDBTools/utils/mariadb/check_version"
)

// Compatibility aliases - re-export types from check_version package
type VersionInfo = check_version.VersionInfo
type VersionFetcher = check_version.VersionFetcher
type VersionParser = check_version.VersionParser
type HTTPVersionFetcher = check_version.HTTPVersionFetcher

// Compatibility functions - re-export functions from check_version package

// NewHTTPVersionFetcher creates a new HTTP-based version fetcher
func NewHTTPVersionFetcher(url string, parser VersionParser) *HTTPVersionFetcher {
	return check_version.NewHTTPVersionFetcher(url, parser)
}

// IsValidVersion checks if a version string is valid for MariaDB
func IsValidVersion(version string) bool {
	return check_version.IsValidVersion(version)
}

// DetermineVersionType determines the type of a version (stable, rc, rolling)
func DetermineVersionType(version string) string {
	return check_version.DetermineVersionType(version)
}

// GetVersionsForOS returns MariaDB versions available for the specified OS
func GetVersionsForOS(osInfo *common.OSInfo, fetchers []VersionFetcher) ([]VersionInfo, error) {
	return check_version.GetVersionsForOS(osInfo, fetchers)
}

// CompareVersions compares two version strings, returns true if v1 < v2
func CompareVersions(v1, v2 string) bool {
	return check_version.CompareVersions(v1, v2)
}

// GetMariaDBEOLDate returns EOL date for MariaDB version using dynamic approach
func GetMariaDBEOLDate(version string) string {
	return check_version.GetMariaDBEOLDate(version)
}
