package install

import (
	"fmt"
	"os/exec"
	"strings"

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

// UpdatePackageCache updates the package manager cache
func (r *RepoSetupManager) UpdatePackageCache() error {
	lg, _ := logger.Get()

	var cmd *exec.Cmd
	switch r.osInfo.PackageType {
	case "rpm":
		cmd = exec.Command("yum", "clean", "all")
		if output, err := cmd.CombinedOutput(); err != nil {
			lg.Warn("Failed to clean yum cache",
				logger.String("output", string(output)),
				logger.Error(err))
		}

		cmd = exec.Command("yum", "makecache")
	case "deb":
		cmd = exec.Command("apt-get", "update")
	default:
		return fmt.Errorf("unsupported package type: %s", r.osInfo.PackageType)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to update package cache",
			logger.String("output", string(output)),
			logger.Error(err))
		return fmt.Errorf("failed to update package cache: %w", err)
	}

	lg.Info("Package cache updated successfully")
	return nil
}
