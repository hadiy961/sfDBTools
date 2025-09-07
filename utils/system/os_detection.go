package system

import (
	"fmt"
	"strings"

	"sfDBTools/internal/logger"

	"github.com/shirou/gopsutil/v3/host"
)

// OSInfo represents basic operating system information
type OSInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	PackageType string `json:"package_type"`
}

// DetectOS detects the current operating system and returns basic info
func DetectOS() (*OSInfo, error) {
	lg, _ := logger.Get()

	info, err := host.Info()
	if err != nil {
		lg.Error("Failed to get host information", logger.Error(err))
		return nil, fmt.Errorf("unable to detect operating system: %w", err)
	}

	osID := normalizeOSID(info.Platform)
	osInfo := &OSInfo{
		ID:          osID,
		Name:        info.Platform,
		Version:     info.PlatformVersion,
		PackageType: getPackageType(osID),
	}

	lg.Info("OS detected",
		logger.String("id", osInfo.ID),
		logger.String("name", osInfo.Name),
		logger.String("version", osInfo.Version),
		logger.String("package_type", osInfo.PackageType))

	return osInfo, nil
}

// normalizeOSID normalizes OS ID to standard values
func normalizeOSID(platform string) string {
	osID := strings.ToLower(platform)

	switch osID {
	case "red hat enterprise linux", "red hat", "rhel":
		return "rhel"
	case "rocky linux":
		return "rocky"
	case "alma linux", "almalinux":
		return "almalinux"
	default:
		return osID
	}
}

// getPackageType returns package manager type based on OS
func getPackageType(osID string) string {
	switch osID {
	case "ubuntu", "debian":
		return "deb"
	case "centos", "rhel", "rocky", "almalinux", "fedora":
		return "rpm"
	case "arch", "manjaro":
		return "pacman"
	case "alpine":
		return "apk"
	default:
		return "unknown"
	}
}

// ValidateOperatingSystem checks if OS is supported for MariaDB installation
func ValidateOperatingSystem() error {
	lg, _ := logger.Get()

	osInfo, err := DetectOS()
	if err != nil {
		return fmt.Errorf("unable to detect operating system: %w", err)
	}

	supportedOS := map[string]bool{
		"centos":    true,
		"ubuntu":    true,
		"rhel":      true,
		"rocky":     true,
		"almalinux": true,
		"debian":    true,
	}

	if !supportedOS[osInfo.ID] {
		lg.Error("Unsupported operating system", logger.String("detected_os", osInfo.ID))
		return fmt.Errorf("unsupported operating system: %s", osInfo.ID)
	}

	lg.Info("Operating system is supported", logger.String("os", osInfo.ID))
	return nil
}
