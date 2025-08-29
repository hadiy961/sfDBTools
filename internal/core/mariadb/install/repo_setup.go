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

// RepoSetupManager handles automated repository setup using MariaDB's official script
type RepoSetupManager struct {
	osInfo *common.OSInfo
}

// NewRepoSetupManager creates a new repository setup manager
func NewRepoSetupManager(osInfo *common.OSInfo) *RepoSetupManager {
	return &RepoSetupManager{osInfo: osInfo}
}

// SetupRepository sets up MariaDB repository using the official script
func (r *RepoSetupManager) SetupRepository(version string) error {
	lg, _ := logger.Get()

	lg.Info("Setting up MariaDB repository using official script",
		logger.String("version", version),
		logger.String("os", r.osInfo.ID))

	// Use the correct script URL from MariaDB documentation
	scriptURL := "https://r.mariadb.com/downloads/mariadb_repo_setup"

	// Prepare version string - MariaDB script expects "mariadb-" prefix
	mariadbVersion := fmt.Sprintf("mariadb-%s", version)

	// Download and run the official MariaDB repository setup script
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf("curl -LsS %s | sudo bash -s -- --mariadb-server-version=%s", scriptURL, mariadbVersion))

	lg.Info("Executing MariaDB repository setup",
		logger.String("script_url", scriptURL),
		logger.String("version", mariadbVersion))

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to setup repository using official script",
			logger.String("output", string(output)),
			logger.Error(err))
		return fmt.Errorf("failed to setup MariaDB repository: %w\nOutput: %s", err, string(output))
	}

	lg.Info("Repository setup completed successfully",
		logger.String("output", string(output)))

	return nil
}

// CleanRepositories removes existing MariaDB repositories
func (r *RepoSetupManager) CleanRepositories() error {
	lg, _ := logger.Get()

	switch r.osInfo.PackageType {
	case "rpm":
		// Remove MariaDB repository files
		repoFiles := []string{
			"/etc/yum.repos.d/MariaDB.repo",
			"/etc/yum.repos.d/mariadb.repo",
		}

		for _, file := range repoFiles {
			cmd := exec.Command("rm", "-f", file)
			if output, err := cmd.CombinedOutput(); err != nil {
				lg.Debug("Repository file removal",
					logger.String("file", file),
					logger.String("output", string(output)))
			}
		}

	case "deb":
		// Remove MariaDB sources
		sourceFiles := []string{
			"/etc/apt/sources.list.d/mariadb.list",
		}

		for _, file := range sourceFiles {
			cmd := exec.Command("rm", "-f", file)
			if output, err := cmd.CombinedOutput(); err != nil {
				lg.Debug("Sources file removal",
					logger.String("file", file),
					logger.String("output", string(output)))
			}
		}
	}

	lg.Info("Cleaned existing MariaDB repositories")
	return nil
}

// IsScriptAvailable checks if the MariaDB repository setup script is available
func (r *RepoSetupManager) IsScriptAvailable() (bool, error) {
	lg, _ := logger.Get()

	scriptURL := "https://r.mariadb.com/downloads/mariadb_repo_setup"

	// Use curl with -L to follow redirects and check final response
	cmd := exec.Command("curl", "-LsS", "-o", "/dev/null", "-w", "%{http_code}", scriptURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to check script availability",
			logger.String("url", scriptURL),
			logger.Error(err))
		return false, err
	}

	// Check if response contains "200" for successful response
	responseCode := strings.TrimSpace(string(output))
	if responseCode == "200" {
		return true, nil
	}

	lg.Warn("Setup script not available", logger.String("response_code", responseCode))
	return false, fmt.Errorf("setup script not available: received response code %s", responseCode)
}

// UpdatePackageCache updates the package manager cache with optimizations for speed
func (r *RepoSetupManager) UpdatePackageCache() error {
	lg, _ := logger.Get()

	// Set timeout for cache update operations (1 minute max)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	var cmd *exec.Cmd
	switch r.osInfo.PackageType {
	case "rpm":
		// Skip aggressive cleaning, only refresh metadata
		// This is much faster than "yum clean all"
		cleanCmd := exec.CommandContext(ctx, "yum", "clean", "metadata")
		if output, err := cleanCmd.CombinedOutput(); err != nil {
			lg.Debug("Failed to clean yum metadata",
				logger.String("output", string(output)),
				logger.Error(err))
		}

		// Use makecache with timer for quicker operation
		cmd = exec.CommandContext(ctx, "yum", "makecache", "timer")
	case "deb":
		// Use optimized flags for faster apt update
		// -q: quiet output, --allow-releaseinfo-change: handle release changes
		cmd = exec.CommandContext(ctx, "apt-get", "update", "-q", "--allow-releaseinfo-change")
	default:
		return fmt.Errorf("unsupported package type: %s", r.osInfo.PackageType)
	}

	lg.Info("Updating package cache (optimized for speed)...",
		logger.String("package_type", r.osInfo.PackageType))

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			lg.Error("Package cache update timed out after 1 minute")
			return fmt.Errorf("package cache update timed out: %w", ctx.Err())
		}

		lg.Error("Failed to update package cache",
			logger.String("output", string(output)),
			logger.Error(err))
		return fmt.Errorf("failed to update package cache: %w", err)
	}

	lg.Info("Package cache updated successfully")
	return nil
}

// UpdatePackageCacheFast performs a minimal package cache update - even faster approach
func (r *RepoSetupManager) UpdatePackageCacheFast() error {
	lg, _ := logger.Get()

	// Set shorter timeout for fast update (30 seconds max)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	switch r.osInfo.PackageType {
	case "rpm":
		// For RPM, just run makecache timer which is faster than full makecache
		cmd = exec.CommandContext(ctx, "yum", "makecache", "timer")
	case "deb":
		// For apt, only update specific sources if mariadb.list exists
		// Check if the file exists first
		if _, err := exec.Command("test", "-f", "/etc/apt/sources.list.d/mariadb.list").CombinedOutput(); err != nil {
			// If mariadb.list doesn't exist, fall back to regular update
			lg.Debug("MariaDB sources list not found, using regular update")
			return r.UpdatePackageCache()
		}

		// Update only MariaDB sources
		cmd = exec.CommandContext(ctx, "apt-get", "update", "-q",
			"-o", "Dir::Etc::sourcelist=/etc/apt/sources.list.d/mariadb.list",
			"-o", "Dir::Etc::sourceparts=/dev/null")
	default:
		return fmt.Errorf("unsupported package type: %s", r.osInfo.PackageType)
	}

	lg.Info("Performing fast package cache update...",
		logger.String("package_type", r.osInfo.PackageType))

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			lg.Warn("Fast package cache update timed out, falling back to regular update")
			return r.UpdatePackageCache()
		}

		lg.Debug("Fast update failed, trying regular update",
			logger.String("output", string(output)),
			logger.Error(err))
		// Fallback to regular update if fast update fails
		return r.UpdatePackageCache()
	}

	lg.Info("Fast package cache update completed successfully")
	return nil
}
