package mariadb

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"sfDBTools/internal/logger"
)

// DetectOS detects the operating system and distribution
func DetectOS() (*OSInfo, error) {
	lg, _ := logger.Get()

	osInfo := &OSInfo{}

	// Try to read /etc/os-release first (most modern distributions)
	if err := parseOSRelease(osInfo); err != nil {
		lg.Debug("Failed to parse /etc/os-release", logger.Error(err))

		// Fallback to other methods
		if err := detectOSFallback(osInfo); err != nil {
			return nil, fmt.Errorf("failed to detect operating system: %w", err)
		}
	}

	lg.Debug("OS detected",
		logger.String("id", osInfo.ID),
		logger.String("version", osInfo.Version))

	return osInfo, nil
}

// parseOSRelease parses /etc/os-release file
func parseOSRelease(osInfo *OSInfo) error {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return fmt.Errorf("failed to open /etc/os-release: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := parts[0]
			value := strings.Trim(parts[1], `"`)

			switch key {
			case "ID":
				osInfo.ID = value
			case "NAME":
				osInfo.Name = value
			case "VERSION_ID":
				osInfo.Version = value
			case "VERSION_CODENAME":
				osInfo.Codename = value
			}
		}
	}

	if osInfo.ID == "" {
		return fmt.Errorf("could not determine OS ID from /etc/os-release")
	}

	return scanner.Err()
}

// detectOSFallback uses fallback methods to detect OS
func detectOSFallback(osInfo *OSInfo) error {
	// Try lsb_release command
	if err := tryLSBRelease(osInfo); err == nil {
		return nil
	}

	// Try /etc/redhat-release for RHEL-based systems
	if err := tryRedHatRelease(osInfo); err == nil {
		return nil
	}

	// Try /etc/debian_version for Debian-based systems
	if err := tryDebianVersion(osInfo); err == nil {
		return nil
	}

	return fmt.Errorf("could not detect operating system using fallback methods")
}

// tryLSBRelease tries to use lsb_release command
func tryLSBRelease(osInfo *OSInfo) error {
	cmd := exec.Command("lsb_release", "-si")
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	osInfo.ID = strings.ToLower(strings.TrimSpace(string(out)))

	cmd = exec.Command("lsb_release", "-sr")
	out, err = cmd.Output()
	if err != nil {
		return err
	}
	osInfo.Version = strings.TrimSpace(string(out))

	return nil
}

// tryRedHatRelease tries to parse /etc/redhat-release
func tryRedHatRelease(osInfo *OSInfo) error {
	content, err := os.ReadFile("/etc/redhat-release")
	if err != nil {
		return err
	}

	line := strings.ToLower(string(content))

	if strings.Contains(line, "centos") {
		osInfo.ID = "centos"
		osInfo.Name = "CentOS"
	} else if strings.Contains(line, "red hat") || strings.Contains(line, "rhel") {
		osInfo.ID = "rhel"
		osInfo.Name = "Red Hat Enterprise Linux"
	} else if strings.Contains(line, "almalinux") {
		osInfo.ID = "almalinux"
		osInfo.Name = "AlmaLinux"
	} else if strings.Contains(line, "rocky") {
		osInfo.ID = "rocky"
		osInfo.Name = "Rocky Linux"
	}

	// Extract version number
	re := regexp.MustCompile(`(\d+)\.?(\d*)?`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		osInfo.Version = matches[1]
		if len(matches) > 2 && matches[2] != "" {
			osInfo.Version += "." + matches[2]
		}
	}

	return nil
}

// tryDebianVersion tries to parse /etc/debian_version
func tryDebianVersion(osInfo *OSInfo) error {
	content, err := os.ReadFile("/etc/debian_version")
	if err != nil {
		return err
	}

	osInfo.Version = strings.TrimSpace(string(content))

	// Check if it's Ubuntu by looking for lsb-release
	if _, err := os.Stat("/etc/lsb-release"); err == nil {
		osInfo.ID = "ubuntu"
		osInfo.Name = "Ubuntu"
	} else {
		osInfo.ID = "debian"
		osInfo.Name = "Debian"
	}

	return nil
}

// IsRHELBased checks if the OS is RHEL-based
func IsRHELBased(osID string) bool {
	rhelBased := []string{"centos", "rhel", "almalinux", "rocky", "fedora"}
	for _, id := range rhelBased {
		if osID == id {
			return true
		}
	}
	return false
}

// IsDebianBased checks if the OS is Debian-based
func IsDebianBased(osID string) bool {
	debianBased := []string{"ubuntu", "debian", "linuxmint"}
	for _, id := range debianBased {
		if osID == id {
			return true
		}
	}
	return false
}
