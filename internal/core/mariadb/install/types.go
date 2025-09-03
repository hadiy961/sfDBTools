package install

import (
	"time"

	"sfDBTools/utils/common"
)

// Config represents installation configuration
type Config struct {
	DryRun bool
	// SkipConfirm, when true, makes the installer non-interactive and
	// automatically answers confirmations (equivalent to --yes)
	SkipConfirm bool
}

// InstallResult represents the result of installation
type InstallResult struct {
	Success       bool          `json:"success"`
	Message       string        `json:"message"`
	Version       string        `json:"version,omitempty"`
	InstalledAt   time.Time     `json:"installed_at"`
	ServiceStatus string        `json:"service_status,omitempty"`
	PackagesCount int           `json:"packages_count,omitempty"`
	Duration      time.Duration `json:"duration,omitempty"`
}

// SystemInfo represents system information for installation
type SystemInfo struct {
	OSInfo            *common.OSInfo
	ExistingService   bool
	ExistingPackages  []string
	InternetAvailable bool
	RepoAvailable     bool
}

// DefaultConfig returns default installation configuration
func DefaultConfig() *Config {
	return &Config{
		DryRun: false,
	}
}

// NewSystemInfo creates a new system info instance
func NewSystemInfo() *SystemInfo {
	return &SystemInfo{
		ExistingPackages: make([]string, 0),
	}
}
