package common

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"sfDBTools/internal/logger"
)

// OSInfo represents operating system information
type OSInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	VersionID    string `json:"version_id"`
	Architecture string `json:"architecture"`
	PackageType  string `json:"package_type"`
	Codename     string `json:"codename,omitempty"`
}

// OSDetector handles OS detection and information gathering
type OSDetector struct{}

// NewOSDetector creates a new OS detector
func NewOSDetector() *OSDetector {
	return &OSDetector{}
}

// DetectOS detects the current operating system and returns OSInfo
func (d *OSDetector) DetectOS() (*OSInfo, error) {
	lg, _ := logger.Get()

	// Read /etc/os-release
	osReleaseContent, err := os.ReadFile("/etc/os-release")
	if err != nil {
		lg.Error("Failed to read /etc/os-release", logger.Error(err))
		return nil, fmt.Errorf("unable to detect operating system: %w", err)
	}

	osInfo := &OSInfo{}
	content := string(osReleaseContent)

	// Parse os-release content
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ID=") {
			osInfo.ID = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		} else if strings.HasPrefix(line, "NAME=") {
			osInfo.Name = strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
		} else if strings.HasPrefix(line, "VERSION=") {
			osInfo.Version = strings.Trim(strings.TrimPrefix(line, "VERSION="), "\"")
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			osInfo.VersionID = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
		} else if strings.HasPrefix(line, "VERSION_CODENAME=") {
			osInfo.Codename = strings.Trim(strings.TrimPrefix(line, "VERSION_CODENAME="), "\"")
		}
	}

	// Detect architecture
	arch, err := d.detectArchitecture()
	if err != nil {
		lg.Warn("Failed to detect architecture, using default", logger.Error(err))
		arch = "x86_64"
	}
	osInfo.Architecture = arch

	// Determine package type
	osInfo.PackageType = d.getPackageType(osInfo.ID)

	lg.Info("OS detected",
		logger.String("id", osInfo.ID),
		logger.String("name", osInfo.Name),
		logger.String("version", osInfo.Version),
		logger.String("version_id", osInfo.VersionID),
		logger.String("architecture", osInfo.Architecture),
		logger.String("package_type", osInfo.PackageType))

	return osInfo, nil
}

// detectArchitecture detects system architecture
func (d *OSDetector) detectArchitecture() (string, error) {
	// Try uname first (more reliable)
	if arch := d.getUnameArchitecture(); arch != "" {
		return arch, nil
	}

	// Fallback to /proc/cpuinfo
	content, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return "", fmt.Errorf("failed to read /proc/cpuinfo: %w", err)
	}

	cpuInfo := string(content)
	if strings.Contains(cpuInfo, "x86_64") || strings.Contains(cpuInfo, "amd64") {
		return "x86_64", nil
	}
	if strings.Contains(cpuInfo, "aarch64") || strings.Contains(cpuInfo, "arm64") {
		return "aarch64", nil
	}

	return "x86_64", nil // fallback
}

// getUnameArchitecture gets architecture using uname -m equivalent
func (d *OSDetector) getUnameArchitecture() string {
	// This would be better implemented with syscall, but for simplicity using file approach
	return ""
}

// getPackageType determines package manager type based on OS
func (d *OSDetector) getPackageType(osID string) string {
	switch strings.ToLower(osID) {
	case "ubuntu", "debian":
		return "deb"
	case "centos", "rhel", "rocky", "almalinux", "fedora", "opensuse", "sles":
		return "rpm"
	case "arch", "manjaro":
		return "pacman"
	case "alpine":
		return "apk"
	default:
		return "unknown"
	}
}

// OSCompatibilityChecker provides OS compatibility checking functionality
type OSCompatibilityChecker struct {
	supportedOS map[string][]string // OS ID -> supported versions
}

// NewOSCompatibilityChecker creates a new OS compatibility checker
func NewOSCompatibilityChecker() *OSCompatibilityChecker {
	return &OSCompatibilityChecker{
		supportedOS: make(map[string][]string),
	}
}

// AddSupportedOS adds a supported operating system with versions
func (c *OSCompatibilityChecker) AddSupportedOS(osID string, versions []string) {
	c.supportedOS[strings.ToLower(osID)] = versions
}

// AddSupportedOSList adds multiple supported operating systems
func (c *OSCompatibilityChecker) AddSupportedOSList(supportedList map[string][]string) {
	for osID, versions := range supportedList {
		c.AddSupportedOS(osID, versions)
	}
}

// IsSupported checks if the detected OS is supported
func (c *OSCompatibilityChecker) IsSupported(osInfo *OSInfo) bool {
	_, exists := c.supportedOS[strings.ToLower(osInfo.ID)]
	return exists
}

// ValidateOSVersion validates if the OS version is supported
func (c *OSCompatibilityChecker) ValidateOSVersion(osInfo *OSInfo) error {
	lg, _ := logger.Get()

	osID := strings.ToLower(osInfo.ID)
	versions, exists := c.supportedOS[osID]
	if !exists {
		return fmt.Errorf("unsupported operating system: %s", osInfo.ID)
	}

	// If no specific versions defined, accept any version
	if len(versions) == 0 {
		return nil
	}

	// Check version compatibility
	switch osID {
	case "ubuntu":
		return c.validateUbuntuVersion(osInfo.VersionID, versions)
	case "centos", "rhel":
		return c.validateRHELVersion(osInfo.VersionID, versions)
	case "rocky", "almalinux":
		return c.validateELVersion(osInfo.VersionID, versions)
	case "debian":
		return c.validateDebianVersion(osInfo.VersionID, versions)
	default:
		lg.Warn("OS version validation not implemented for this OS", logger.String("os", osInfo.ID))
		return nil
	}
}

// validateUbuntuVersion validates Ubuntu version
func (c *OSCompatibilityChecker) validateUbuntuVersion(versionID string, supportedVersions []string) error {
	for _, supported := range supportedVersions {
		if strings.HasPrefix(versionID, supported) {
			return nil
		}
	}
	return fmt.Errorf("unsupported Ubuntu version: %s. Supported versions: %v", versionID, supportedVersions)
}

// validateRHELVersion validates RHEL/CentOS version
func (c *OSCompatibilityChecker) validateRHELVersion(versionID string, supportedVersions []string) error {
	// Create regex pattern from supported versions
	pattern := "^(" + strings.Join(supportedVersions, "|") + ")"
	re := regexp.MustCompile(pattern)

	if re.MatchString(versionID) {
		return nil
	}

	return fmt.Errorf("unsupported RHEL/CentOS version: %s. Supported versions: %v", versionID, supportedVersions)
}

// validateELVersion validates Enterprise Linux (Rocky/AlmaLinux) version
func (c *OSCompatibilityChecker) validateELVersion(versionID string, supportedVersions []string) error {
	// Create regex pattern from supported versions
	pattern := "^(" + strings.Join(supportedVersions, "|") + ")"
	re := regexp.MustCompile(pattern)

	if re.MatchString(versionID) {
		return nil
	}

	return fmt.Errorf("unsupported Enterprise Linux version: %s. Supported versions: %v", versionID, supportedVersions)
}

// validateDebianVersion validates Debian version
func (c *OSCompatibilityChecker) validateDebianVersion(versionID string, supportedVersions []string) error {
	for _, supported := range supportedVersions {
		if strings.HasPrefix(versionID, supported) {
			return nil
		}
	}
	return fmt.Errorf("unsupported Debian version: %s. Supported versions: %v", versionID, supportedVersions)
}

// GetSupportedOSList returns list of supported operating systems
func (c *OSCompatibilityChecker) GetSupportedOSList() []string {
	var osList []string
	for osID := range c.supportedOS {
		osList = append(osList, osID)
	}
	return osList
}

// ValidateOperatingSystem is a general function that checks OS compatibility
// This replaces the specific mariadb os_validator for reusability
func ValidateOperatingSystem(supportedOS map[string][]string) error {
	lg, _ := logger.Get()

	detector := NewOSDetector()
	osInfo, err := detector.DetectOS()
	if err != nil {
		return fmt.Errorf("unable to detect operating system: %w", err)
	}

	checker := NewOSCompatibilityChecker()
	checker.AddSupportedOSList(supportedOS)

	if !checker.IsSupported(osInfo) {
		lg.Error("Unsupported operating system", logger.String("detected_os", osInfo.ID))
		return fmt.Errorf("unsupported operating system: %s. Supported OS: %v",
			osInfo.ID, checker.GetSupportedOSList())
	}

	if err := checker.ValidateOSVersion(osInfo); err != nil {
		lg.Error("OS version validation failed", logger.Error(err))
		return fmt.Errorf("OS version validation failed: %w", err)
	}

	lg.Info("Operating system detected and supported",
		logger.String("os", osInfo.ID),
		logger.String("version", osInfo.Version))

	return nil
}

// MariaDBSupportedOS returns the supported OS configuration for MariaDB
func MariaDBSupportedOS() map[string][]string {
	return map[string][]string{
		"centos":    {"7", "8", "9"},
		"ubuntu":    {"18.04", "20.04", "22.04", "24.04"},
		"rhel":      {"7", "8", "9"},
		"rocky":     {"8", "9"},
		"almalinux": {"8", "9"},
		"debian":    {"10", "11", "12"},
	}
}
