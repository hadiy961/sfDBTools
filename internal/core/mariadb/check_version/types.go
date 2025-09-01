package check_version

import "time"

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
	OSInfo            *OSInfo       `json:"os_info,omitempty"`
}

// OSInfo represents operating system information
type OSInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	Architecture string `json:"architecture"`
	PackageType  string `json:"package_type"` // rpm, deb
}

// VersionType constants
const (
	VersionTypeStable  = "stable"
	VersionTypeRC      = "rc"
	VersionTypeRolling = "rolling"
)

// Config holds configuration for version checking
type Config struct {
	Timeout        time.Duration `json:"timeout"`
	EnableFallback bool          `json:"enable_fallback"`
	UserAgent      string        `json:"user_agent"`
	OSSpecific     bool          `json:"os_specific"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Timeout:        30 * time.Second,
		EnableFallback: true,
		UserAgent:      "sfDBTools/1.0 MariaDB-Version-Checker",
		OSSpecific:     true,
	}
}
