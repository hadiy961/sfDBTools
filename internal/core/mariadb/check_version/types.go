package check_version

import (
	"time"

	"sfDBTools/utils/common"
)

// VersionInfo represents MariaDB version information
type VersionInfo struct {
	Version     string `json:"version"`
	Type        string `json:"type"` // stable, rc, rolling
	ReleaseDate string `json:"release_date,omitempty"`
	EOLDate     string `json:"eol_date,omitempty"`
}

// VersionCheckResult contains the result of version checking
type VersionCheckResult struct {
	AvailableVersions []VersionInfo  `json:"available_versions"`
	CurrentStable     string         `json:"current_stable"`
	LatestVersion     string         `json:"latest_version"`
	LatestMinor       string         `json:"latest_minor"`
	CheckTime         time.Time      `json:"check_time"`
	OSInfo            *common.OSInfo `json:"os_info,omitempty"`
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
