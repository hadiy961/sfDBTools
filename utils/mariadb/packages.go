package mariadb

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"sfDBTools/internal/logger"
)

// RemovePackages removes MariaDB/MySQL packages based on OS
func RemovePackages(osInfo *OSInfo) (int, []string, error) {
	lg, _ := logger.Get()

	if IsRHELBased(osInfo.ID) {
		return removeRHELPackages(lg)
	} else if IsDebianBased(osInfo.ID) {
		return removeDebianPackages(lg)
	}

	return 0, nil, fmt.Errorf("unsupported operating system: %s", osInfo.ID)
}

// GetInstalledPackageCount returns the count of installed MariaDB/MySQL packages
func GetInstalledPackageCount(osInfo *OSInfo) (int, []string, error) {
	var packages []string
	var err error

	if IsRHELBased(osInfo.ID) {
		packages, err = getRHELInstalledPackages()
	} else if IsDebianBased(osInfo.ID) {
		packages, err = getDebianInstalledPackages()
	} else {
		return 0, nil, fmt.Errorf("unsupported operating system: %s", osInfo.ID)
	}

	if err != nil {
		return 0, nil, fmt.Errorf("failed to get installed packages: %w", err)
	}

	return len(packages), packages, nil
}

// removeRHELPackages removes packages on RHEL-based systems
func removeRHELPackages(lg *logger.Logger) (int, []string, error) {
	// Get list of installed MariaDB/MySQL packages
	packages, err := getRHELInstalledPackages()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get installed packages: %w", err)
	}

	if len(packages) == 0 {
		lg.Info("No MariaDB/MySQL packages found to remove")
		return 0, []string{}, nil
	}

	lg.Info("Found packages to remove", logger.Int("count", len(packages)))

	// Remove packages using dnf/yum
	var cmd *exec.Cmd
	if CommandExists("dnf") {
		args := append([]string{"remove", "-y"}, packages...)
		cmd = exec.Command("dnf", args...)
	} else if CommandExists("yum") {
		args := append([]string{"remove", "-y"}, packages...)
		cmd = exec.Command("yum", args...)
	} else {
		return 0, packages, fmt.Errorf("neither dnf nor yum found")
	}

	lg.Debug("Executing package removal command",
		logger.String("command", cmd.String()))

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Warn("Package removal completed with warnings",
			logger.Error(err),
			logger.String("output", string(output)))
		// Don't treat as fatal error, some packages might not exist
	} else {
		lg.Info("Packages removed successfully")
	}

	return len(packages), packages, nil
}

// removeDebianPackages removes packages on Debian-based systems
func removeDebianPackages(lg *logger.Logger) (int, []string, error) {
	// Get list of installed MariaDB/MySQL packages
	packages, err := getDebianInstalledPackages()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get installed packages: %w", err)
	}

	if len(packages) == 0 {
		lg.Info("No MariaDB/MySQL packages found to remove")
		return 0, []string{}, nil
	}

	lg.Info("Found packages to remove", logger.Int("count", len(packages)))

	// Remove packages using apt
	args := append([]string{"remove", "-y", "--purge"}, packages...)
	cmd := exec.Command("apt", args...)

	lg.Debug("Executing package removal command",
		logger.String("command", cmd.String()))

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Warn("Package removal completed with warnings",
			logger.Error(err),
			logger.String("output", string(output)))
		// Don't treat as fatal error, some packages might not exist
	} else {
		lg.Info("Packages removed successfully")
	}

	// Clean up package cache
	cleanCmd := exec.Command("apt", "autoremove", "-y")
	cleanCmd.Run()

	return len(packages), packages, nil
}

// getRHELInstalledPackages gets list of installed MariaDB/MySQL packages on RHEL
func getRHELInstalledPackages() ([]string, error) {
	cmd := exec.Command("rpm", "-qa")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var packages []string
	lines := strings.Split(string(output), "\n")

	// Patterns to match MariaDB/MySQL packages
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`^[Mm]aria[Dd][Bb].*`),
		regexp.MustCompile(`^[Mm]y[Ss][Qq][Ll].*`),
		regexp.MustCompile(`^mysql.*`),
		regexp.MustCompile(`^mariadb.*`),
		regexp.MustCompile(`^galera.*`),
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		for _, pattern := range patterns {
			if pattern.MatchString(line) {
				// Extract package name properly (remove version and architecture info)
				// Example: mariadb-connector-c-config-3.4.4-1.el10.noarch -> mariadb-connector-c-config
				parts := strings.Split(line, "-")
				if len(parts) >= 2 {
					// Find the last numeric part to separate version from package name
					var packageParts []string
					for _, part := range parts {
						// Stop at first part that looks like version (starts with digit)
						if regexp.MustCompile(`^\d`).MatchString(part) {
							break
						}
						packageParts = append(packageParts, part)
					}

					if len(packageParts) > 0 {
						packages = append(packages, strings.Join(packageParts, "-"))
					} else if len(parts) > 0 {
						// Fallback to first part only
						packages = append(packages, parts[0])
					}
				}
				break
			}
		}
	}

	// Remove duplicates
	packages = removeDuplicates(packages)
	return packages, nil
}

// getDebianInstalledPackages gets list of installed MariaDB/MySQL packages on Debian
func getDebianInstalledPackages() ([]string, error) {
	cmd := exec.Command("dpkg", "-l")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var packages []string
	lines := strings.Split(string(output), "\n")

	// Patterns to match MariaDB/MySQL packages
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`^ii\s+([Mm]aria[Dd][Bb]\S+)`),
		regexp.MustCompile(`^ii\s+([Mm]y[Ss][Qq][Ll]\S+)`),
		regexp.MustCompile(`^ii\s+(mysql\S+)`),
		regexp.MustCompile(`^ii\s+(mariadb\S+)`),
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		for _, pattern := range patterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) > 1 {
				packages = append(packages, matches[1])
				break
			}
		}
	}

	// Remove duplicates
	packages = removeDuplicates(packages)
	return packages, nil
}

// removeDuplicates removes duplicate strings from slice
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
