package check_version

import (
	"time"
)

// Release represents a MariaDB major release from the API
type Release struct {
	ID          string `json:"release_id"`
	Name        string `json:"release_name"`
	Status      string `json:"release_status"`
	SupportType string `json:"release_support_type"`
	EOLDate     string `json:"release_eol_date"`
}

// APIResponse represents the API response structure
type APIResponse struct {
	MajorReleases []Release `json:"major_releases"`
}

// VersionInfo represents version information for display
type VersionInfo struct {
	Version       string
	EOL           string
	LatestVersion string
	SupportType   string
}

// VersionDetails represents detailed version information from version-specific API
type VersionDetails struct {
	Releases map[string]interface{} `json:"releases"`
}

// CheckVersionConfig holds configuration for version checking
type CheckVersionConfig struct {
	MinimumVersion string
	APITimeout     time.Duration
	APIBaseURL     string
}

// DefaultCheckVersionConfig returns default configuration for version checking
func DefaultCheckVersionConfig() *CheckVersionConfig {
	return &CheckVersionConfig{
		MinimumVersion: "10.6",
		APITimeout:     30 * time.Second,
		APIBaseURL:     "https://downloads.mariadb.org/rest-api/mariadb/",
	}
}
