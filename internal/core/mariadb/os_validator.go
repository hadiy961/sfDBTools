package mariadb

import (
	"fmt"
	"os"
	"strings"

	"sfDBTools/internal/logger"
)

// ValidateOperatingSystem checks if the current OS is supported for MariaDB operations
func ValidateOperatingSystem() error {
	lg, _ := logger.Get()

	// Read /etc/os-release to detect the OS
	osReleaseContent, err := os.ReadFile("/etc/os-release")
	if err != nil {
		lg.Error("Failed to read /etc/os-release", logger.Error(err))
		return fmt.Errorf("unable to detect operating system: %w", err)
	}

	osInfo := string(osReleaseContent)
	lg.Debug("OS release content", logger.String("content", osInfo))

	// Extract ID from os-release
	osID := extractOSID(osInfo)
	if osID == "" {
		lg.Error("Failed to extract OS ID from /etc/os-release")
		return fmt.Errorf("unable to determine operating system ID")
	}

	// Check if OS is supported
	if !isSupportedOS(osID) {
		lg.Error("Unsupported operating system", logger.String("detected_os", osID))
		return fmt.Errorf("unsupported operating system: %s. Supported OS: %s",
			osID, strings.Join(getSupportedOSList(), ", "))
	}

	lg.Info("Operating system detected and supported", logger.String("os", osID))
	return nil
}

// extractOSID extracts the OS ID from /etc/os-release content
func extractOSID(osInfo string) string {
	for _, line := range strings.Split(osInfo, "\n") {
		if strings.HasPrefix(line, "ID=") {
			return strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		}
	}
	return ""
}

// isSupportedOS checks if the given OS ID is in the supported list
func isSupportedOS(osID string) bool {
	supportedOS := getSupportedOSList()
	for _, supported := range supportedOS {
		if strings.EqualFold(osID, supported) {
			return true
		}
	}
	return false
}

// getSupportedOSList returns the list of supported operating systems
func getSupportedOSList() []string {
	return []string{"centos", "ubuntu", "rhel", "rocky", "almalinux"}
}
