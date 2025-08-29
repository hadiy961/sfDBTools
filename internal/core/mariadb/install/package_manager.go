package install

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
)

// YumPackageManager implements PackageManager for YUM-based systems
type YumPackageManager struct {
	osInfo *common.OSInfo
}

// NewYumPackageManager creates a new YUM package manager
func NewYumPackageManager(osInfo *common.OSInfo) *YumPackageManager {
	return &YumPackageManager{osInfo: osInfo}
}

// Install installs a package using YUM with optimizations for speed
func (y *YumPackageManager) Install(packageName string, version string) error {
	lg, _ := logger.Get()

	fullPackageName := y.GetPackageName(version)

	// Set timeout for installation (10 minutes max)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Use optimized yum flags for faster installation
	cmd := exec.CommandContext(ctx, "yum", "install", "-y",
		"--nogpgcheck",   // Skip GPG check for speed (we trust MariaDB repo)
		"--skip-broken",  // Skip broken dependencies
		"--downloadonly", // First download only
		fullPackageName)

	lg.Info("Downloading MariaDB package",
		logger.String("package", fullPackageName),
		logger.String("command", cmd.String()))

	// First, download the package
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			lg.Error("Package download timed out after 10 minutes")
			return fmt.Errorf("package download timed out: %w", ctx.Err())
		}
		lg.Error("Failed to download package",
			logger.String("package", fullPackageName),
			logger.String("output", string(output)),
			logger.Error(err))
		return fmt.Errorf("failed to download %s: %w\nOutput: %s", fullPackageName, err, string(output))
	}

	lg.Info("Package downloaded, now installing...", logger.String("package", fullPackageName))

	// Reset timeout for installation
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel2()

	// Now install the downloaded package
	installCmd := exec.CommandContext(ctx2, "yum", "install", "-y",
		"--nogpgcheck",  // Skip GPG check for speed
		"--skip-broken", // Skip broken dependencies
		fullPackageName)

	lg.Info("Installing MariaDB package",
		logger.String("package", fullPackageName),
		logger.String("command", installCmd.String()))

	output, err = installCmd.CombinedOutput()
	if err != nil {
		// Check if it was a timeout
		if ctx2.Err() == context.DeadlineExceeded {
			lg.Error("Package installation timed out after 5 minutes")
			return fmt.Errorf("package installation timed out: %w", ctx2.Err())
		}
		lg.Error("Failed to install package",
			logger.String("package", fullPackageName),
			logger.String("output", string(output)),
			logger.Error(err))
		return fmt.Errorf("failed to install %s: %w\nOutput: %s", fullPackageName, err, string(output))
	}

	lg.Info("Successfully installed package", logger.String("package", fullPackageName))
	return nil
}

// InstallFast installs a package using YUM with maximum speed optimizations
func (y *YumPackageManager) InstallFast(packageName string, version string) error {
	lg, _ := logger.Get()

	fullPackageName := y.GetPackageName(version)

	// Set timeout for installation (8 minutes max)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	// Use highly optimized yum flags for maximum speed
	cmd := exec.CommandContext(ctx, "yum", "install", "-y",
		"--nogpgcheck",         // Skip GPG check completely
		"--skip-broken",        // Skip broken dependencies
		"--assumeyes",          // Assume yes to all questions
		"--quiet",              // Reduce output noise
		"--disablerepo=*",      // Disable all repos
		"--enablerepo=mariadb", // Only enable MariaDB repo
		fullPackageName)

	lg.Info("Installing MariaDB package (fast mode)",
		logger.String("package", fullPackageName))

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			lg.Error("Package installation timed out after 8 minutes")
			return fmt.Errorf("package installation timed out: %w", ctx.Err())
		}

		// If fast install fails, fallback to regular install
		lg.Warn("Fast installation failed, trying regular installation",
			logger.String("output", string(output)),
			logger.Error(err))
		return y.Install(packageName, version)
	}

	lg.Info("Successfully installed package (fast mode)", logger.String("package", fullPackageName))
	return nil
}

// Remove removes a package using YUM
func (y *YumPackageManager) Remove(packageName string) error {
	lg, _ := logger.Get()

	cmd := exec.Command("yum", "remove", "-y", packageName)
	lg.Info("Removing package", logger.String("package", packageName))

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to remove package",
			logger.String("package", packageName),
			logger.String("output", string(output)),
			logger.Error(err))
		return fmt.Errorf("failed to remove %s: %w", packageName, err)
	}

	return nil
}

// IsInstalled checks if a package is installed
func (y *YumPackageManager) IsInstalled(packageName string) (bool, string, error) {
	cmd := exec.Command("rpm", "-q", packageName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Package not installed
		return false, "", nil
	}

	// Extract version from rpm output
	version := strings.TrimSpace(string(output))
	return true, version, nil
}

// Update updates package repository
func (y *YumPackageManager) Update() error {
	lg, _ := logger.Get()

	cmd := exec.Command("yum", "update", "-y")
	lg.Info("Updating package repository")

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to update repository",
			logger.String("output", string(output)),
			logger.Error(err))
		return fmt.Errorf("failed to update repository: %w", err)
	}

	return nil
}

// AddRepository adds MariaDB repository
func (y *YumPackageManager) AddRepository(repoConfig RepositoryConfig) error {
	lg, _ := logger.Get()

	// Create repository file content
	repoContent := fmt.Sprintf(`[%s]
name=%s
baseurl=%s
gpgkey=%s
gpgcheck=1
enabled=1
priority=%d
`, repoConfig.Name, repoConfig.Name, repoConfig.BaseURL, repoConfig.GPGKey, repoConfig.Priority)

	// Write repository file
	repoFile := fmt.Sprintf("/etc/yum.repos.d/%s.repo", repoConfig.Name)

	cmd := exec.Command("tee", repoFile)
	cmd.Stdin = strings.NewReader(repoContent)

	lg.Info("Adding repository",
		logger.String("name", repoConfig.Name),
		logger.String("file", repoFile))

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to add repository",
			logger.String("output", string(output)),
			logger.Error(err))
		return fmt.Errorf("failed to add repository: %w", err)
	}

	// Import GPG key
	if repoConfig.GPGKey != "" {
		if err := y.importGPGKey(repoConfig.GPGKey); err != nil {
			lg.Warn("Failed to import GPG key", logger.Error(err))
		}
	}

	return nil
}

// GetPackageName returns the full package name for installation
func (y *YumPackageManager) GetPackageName(version string) string {
	// Based on MariaDB documentation, use the official package names
	// For RPM systems: MariaDB-server MariaDB-client MariaDB-backup
	return "MariaDB-server"
}

// importGPGKey imports GPG key for repository
func (y *YumPackageManager) importGPGKey(gpgKey string) error {
	cmd := exec.Command("rpm", "--import", gpgKey)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to import GPG key: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// AptPackageManager implements PackageManager for APT-based systems
type AptPackageManager struct {
	osInfo *common.OSInfo
}

// NewAptPackageManager creates a new APT package manager
func NewAptPackageManager(osInfo *common.OSInfo) *AptPackageManager {
	return &AptPackageManager{osInfo: osInfo}
}

// Install installs a package using APT
func (a *AptPackageManager) Install(packageName string, version string) error {
	lg, _ := logger.Get()

	fullPackageName := a.GetPackageName(version)

	// Update package list first
	if err := a.Update(); err != nil {
		lg.Warn("Failed to update package list", logger.Error(err))
	}

	cmd := exec.Command("apt-get", "install", "-y", fullPackageName)

	lg.Info("Installing MariaDB package",
		logger.String("package", fullPackageName),
		logger.String("command", cmd.String()))

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to install package",
			logger.String("package", fullPackageName),
			logger.String("output", string(output)),
			logger.Error(err))
		return fmt.Errorf("failed to install %s: %w\nOutput: %s", fullPackageName, err, string(output))
	}

	lg.Info("Successfully installed package", logger.String("package", fullPackageName))
	return nil
}

// InstallFast installs a package using APT with optimizations for speed
func (a *AptPackageManager) InstallFast(packageName string, version string) error {
	lg, _ := logger.Get()

	fullPackageName := a.GetPackageName(version)

	// Set timeout for installation (8 minutes max)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	// Use optimized apt flags for faster installation
	cmd := exec.CommandContext(ctx, "apt-get", "install", "-y",
		"--no-install-recommends", // Don't install recommended packages
		"--assume-yes",            // Assume yes to all questions
		"--quiet",                 // Reduce output noise
		"--allow-unauthenticated", // Skip authentication for speed
		fullPackageName)

	lg.Info("Installing MariaDB package (fast mode)",
		logger.String("package", fullPackageName))

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			lg.Error("Package installation timed out after 8 minutes")
			return fmt.Errorf("package installation timed out: %w", ctx.Err())
		}

		// If fast install fails, fallback to regular install
		lg.Warn("Fast installation failed, trying regular installation",
			logger.String("output", string(output)),
			logger.Error(err))
		return a.Install(packageName, version)
	}

	lg.Info("Successfully installed package (fast mode)", logger.String("package", fullPackageName))
	return nil
}

// Remove removes a package using APT
func (a *AptPackageManager) Remove(packageName string) error {
	lg, _ := logger.Get()

	cmd := exec.Command("apt-get", "remove", "-y", packageName)
	lg.Info("Removing package", logger.String("package", packageName))

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to remove package",
			logger.String("package", packageName),
			logger.String("output", string(output)),
			logger.Error(err))
		return fmt.Errorf("failed to remove %s: %w", packageName, err)
	}

	return nil
}

// IsInstalled checks if a package is installed
func (a *AptPackageManager) IsInstalled(packageName string) (bool, string, error) {
	cmd := exec.Command("dpkg", "-l", packageName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return false, "", nil
	}

	// Parse dpkg output to get version
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ii") && strings.Contains(line, packageName) {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				return true, fields[2], nil
			}
		}
	}

	return false, "", nil
}

// Update updates package repository
func (a *AptPackageManager) Update() error {
	lg, _ := logger.Get()

	cmd := exec.Command("apt-get", "update")
	lg.Info("Updating package repository")

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to update repository",
			logger.String("output", string(output)),
			logger.Error(err))
		return fmt.Errorf("failed to update repository: %w", err)
	}

	return nil
}

// AddRepository adds MariaDB repository for APT
func (a *AptPackageManager) AddRepository(repoConfig RepositoryConfig) error {
	lg, _ := logger.Get()

	// Install required packages for adding repositories
	prereqCmd := exec.Command("apt-get", "install", "-y", "software-properties-common", "dirmngr", "apt-transport-https")
	if output, err := prereqCmd.CombinedOutput(); err != nil {
		lg.Warn("Failed to install prerequisites",
			logger.String("output", string(output)),
			logger.Error(err))
	}

	// Add GPG key
	if repoConfig.GPGKey != "" {
		if err := a.addGPGKey(repoConfig.GPGKey); err != nil {
			return fmt.Errorf("failed to add GPG key: %w", err)
		}
	}

	// Add repository
	cmd := exec.Command("add-apt-repository", repoConfig.BaseURL)

	lg.Info("Adding repository",
		logger.String("name", repoConfig.Name),
		logger.String("url", repoConfig.BaseURL))

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to add repository",
			logger.String("output", string(output)),
			logger.Error(err))
		return fmt.Errorf("failed to add repository: %w", err)
	}

	return nil
}

// GetPackageName returns the full package name for installation
func (a *AptPackageManager) GetPackageName(version string) string {
	// Based on MariaDB documentation, use the official package names
	// For DEB systems: mariadb-server mariadb-client mariadb-backup
	return "mariadb-server"
}

// addGPGKey adds GPG key for repository
func (a *AptPackageManager) addGPGKey(gpgKey string) error {
	cmd := exec.Command("wget", "-qO-", gpgKey)
	wgetOutput, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to download GPG key: %w", err)
	}

	aptKeyCmd := exec.Command("apt-key", "add", "-")
	aptKeyCmd.Stdin = strings.NewReader(string(wgetOutput))

	output, err := aptKeyCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add GPG key: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// NewPackageManager creates appropriate package manager based on OS
func NewPackageManager(osInfo *common.OSInfo) PackageManager {
	switch osInfo.PackageType {
	case "rpm":
		return NewYumPackageManager(osInfo)
	case "deb":
		return NewAptPackageManager(osInfo)
	default:
		return nil
	}
}
