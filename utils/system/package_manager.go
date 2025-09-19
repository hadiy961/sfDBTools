package system

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"sfDBTools/utils/terminal"
)

// PackageManager interface provides abstraction for package management operations
type PackageManager interface {
	Install(packages []string) error
	Remove(packages []string) error
	IsInstalled(pkg string) bool
	GetInstalledPackages() ([]string, error)
	UpdateCache() error
	Upgrade() error
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

// Install installs the specified packages
func (pm *packageManager) Install(packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	var cmd *exec.Cmd
	switch pm.packageTool {
	case "yum":
		args := append([]string{"install", "-y"}, packages...)
		cmd = exec.Command("yum", args...)
	case "apt":
		args := append([]string{"install", "-y"}, packages...)
		cmd = exec.Command("apt", args...)
	case "dnf":
		args := append([]string{"install", "-y"}, packages...)
		cmd = exec.Command("dnf", args...)
	default:
		return fmt.Errorf("unsupported package manager")
	}

	// Stream stdout and stderr so callers can see live progress (like UpdateCache)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start install command: %w", err)
	}

	// stream stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			terminal.SafePrintln(scanner.Text())
		}
	}()

	// stream stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			terminal.SafePrintln(scanner.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to install packages %v: %w", packages, err)
	}

	return nil
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

	// Stream stdout and stderr so callers can see live progress
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start remove command: %w", err)
	}

	// stream stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			terminal.SafePrintln(scanner.Text())
		}
	}()

	// stream stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			terminal.SafePrintln(scanner.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to remove packages %v: %w", packages, err)
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

// UpdateCache updates the package manager cache
func (pm *packageManager) UpdateCache() error {
	var cmd *exec.Cmd
	switch pm.packageTool {
	case "yum":
		cmd = exec.Command("yum", "makecache")
	case "apt":
		cmd = exec.Command("apt", "update")
	case "dnf":
		cmd = exec.Command("dnf", "makecache")
	default:
		return fmt.Errorf("unsupported package manager: %s", pm.packageTool)
	}

	// Stream stdout and stderr and print lines using terminal.SafePrintln so
	// active spinner (if any) is paused/resumed properly.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start update cache command: %w", err)
	}

	// stream stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			terminal.SafePrintln(scanner.Text())
		}
	}()

	// stream stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			terminal.SafePrintln(scanner.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to update package cache: %w", err)
	}

	return nil
}

// Upgrade performs a system package upgrade (distribution-specific) and streams output
func (pm *packageManager) Upgrade() error {
	var cmd *exec.Cmd
	switch pm.packageTool {
	case "yum":
		// yum update will update packages
		cmd = exec.Command("yum", "update", "-y")
	case "apt":
		// apt upgrade with -y to auto confirm
		cmd = exec.Command("apt", "upgrade", "-y")
	case "dnf":
		cmd = exec.Command("dnf", "upgrade", "-y")
	default:
		return fmt.Errorf("unsupported package manager: %s", pm.packageTool)
	}

	// Stream stdout and stderr and print lines using terminal.SafePrintln so
	// active spinner (if any) is paused/resumed properly.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start upgrade command: %w", err)
	}

	// stream stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			terminal.SafePrintln(scanner.Text())
		}
	}()

	// stream stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			terminal.SafePrintln(scanner.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to upgrade packages: %w", err)
	}

	return nil
}

// isCommandAvailable checks if a command is available in PATH
func isCommandAvailable(name string) bool {
	cmd := exec.Command("which", name)
	err := cmd.Run()
	return err == nil
}
