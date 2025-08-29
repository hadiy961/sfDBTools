package system

import (
	"fmt"
	"os/exec"
	"strings"
)

// PackageManager interface provides abstraction for package management operations
type PackageManager interface {
	Remove(packages []string) error
	IsInstalled(pkg string) bool
	GetInstalledPackages() ([]string, error)
}

// packageManager implements PackageManager interface
type packageManager struct {
	packageTool string // yum, apt, dnf, etc.
}

// NewPackageManager creates a new package manager based on the system
func NewPackageManager() PackageManager {
	// Detect package manager
	if isCommandAvailable("yum") {
		return &packageManager{packageTool: "yum"}
	} else if isCommandAvailable("apt") {
		return &packageManager{packageTool: "apt"}
	} else if isCommandAvailable("dnf") {
		return &packageManager{packageTool: "dnf"}
	}
	return &packageManager{packageTool: "unknown"}
}

// Remove removes the specified packages
func (pm *packageManager) Remove(packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	var cmd *exec.Cmd
	switch pm.packageTool {
	case "yum":
		args := append([]string{"remove", "-y"}, packages...)
		cmd = exec.Command("yum", args...)
	case "apt":
		args := append([]string{"remove", "-y"}, packages...)
		cmd = exec.Command("apt", args...)
	case "dnf":
		args := append([]string{"remove", "-y"}, packages...)
		cmd = exec.Command("dnf", args...)
	default:
		return fmt.Errorf("unsupported package manager")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove packages %v: %w\nOutput: %s", packages, err, string(output))
	}

	return nil
}

// IsInstalled checks if a package is installed
func (pm *packageManager) IsInstalled(pkg string) bool {
	var cmd *exec.Cmd
	switch pm.packageTool {
	case "yum":
		cmd = exec.Command("rpm", "-q", pkg)
	case "apt":
		cmd = exec.Command("dpkg", "-l", pkg)
	case "dnf":
		cmd = exec.Command("rpm", "-q", pkg)
	default:
		return false
	}

	err := cmd.Run()
	return err == nil
}

// GetInstalledPackages returns a list of MariaDB/MySQL related packages
func (pm *packageManager) GetInstalledPackages() ([]string, error) {
	var cmd *exec.Cmd
	var packages []string

	switch pm.packageTool {
	case "yum", "dnf":
		cmd = exec.Command("rpm", "-qa", "--queryformat", "%{NAME}\n")
	case "apt":
		cmd = exec.Command("dpkg", "-l")
	default:
		return nil, fmt.Errorf("unsupported package manager: %s", pm.packageTool)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get installed packages: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), "mariadb") ||
			strings.Contains(strings.ToLower(line), "mysql") {
			if pm.packageTool == "apt" {
				// For apt, extract package name from dpkg output
				fields := strings.Fields(line)
				if len(fields) >= 2 && (fields[0] == "ii" || fields[0] == "rc") {
					packages = append(packages, fields[1])
				}
			} else {
				// For rpm-based systems
				packages = append(packages, line)
			}
		}
	}

	return packages, nil
}

// isCommandAvailable checks if a command is available in PATH
func isCommandAvailable(name string) bool {
	cmd := exec.Command("which", name)
	err := cmd.Run()
	return err == nil
}
