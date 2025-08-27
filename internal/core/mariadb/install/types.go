package install

import (
	"time"

	"sfDBTools/internal/core/mariadb/check_version"
	"sfDBTools/utils/common"
)

// InstallConfig holds configuration for MariaDB installation
type InstallConfig struct {
	Version        string
	AutoConfirm    bool
	DataDir        string
	ConfigFile     string
	RootPassword   string
	RemoveExisting bool
	Timeout        time.Duration
	EnableSecurity bool
	StartService   bool
}

// DefaultInstallConfig returns default installation configuration
func DefaultInstallConfig() *InstallConfig {
	return &InstallConfig{
		AutoConfirm:    false,
		DataDir:        "/var/lib/mysql",
		ConfigFile:     "/etc/mysql/mariadb.conf.d/50-server.cnf",
		RemoveExisting: false,
		Timeout:        30 * time.Minute,
		EnableSecurity: true,
		StartService:   true,
	}
}

// InstallStatus represents the current installation status
type InstallStatus struct {
	Step       string
	Progress   int
	Message    string
	IsError    bool
	IsComplete bool
}

// InstallResult represents the result of installation
type InstallResult struct {
	Success       bool
	Version       string
	InstalledPath string
	DataDir       string
	ConfigFile    string
	ServiceName   string
	Message       string
	Error         error
}

// PackageManager interface for different OS package managers
type PackageManager interface {
	Install(packageName string, version string) error
	Remove(packageName string) error
	IsInstalled(packageName string) (bool, string, error)
	Update() error
	AddRepository(repoConfig RepositoryConfig) error
	GetPackageName(version string) string
}

// RepositoryConfig holds repository configuration for different OS
type RepositoryConfig struct {
	Name     string
	BaseURL  string
	GPGKey   string
	Priority int
}

// OSInfo represents operating system information
// Deprecated: Use common.OSInfo instead
type OSInfo = common.OSInfo

// InstallationStep represents a step in the installation process
type InstallationStep struct {
	Name        string
	Description string
	Required    bool
	Function    func() error
}

// SelectableVersion represents a version that can be selected for installation
type SelectableVersion struct {
	Version       string
	LatestVersion string
	EOL           string
	SupportType   string
	Index         int
}

// ConvertVersionInfo converts check_version.VersionInfo to SelectableVersion
func ConvertVersionInfo(versions []check_version.VersionInfo) []SelectableVersion {
	selectableVersions := make([]SelectableVersion, len(versions))
	for i, v := range versions {
		selectableVersions[i] = SelectableVersion{
			Version:       v.Version,
			LatestVersion: v.LatestVersion,
			EOL:           v.EOL,
			SupportType:   v.SupportType,
			Index:         i + 1,
		}
	}
	return selectableVersions
}
