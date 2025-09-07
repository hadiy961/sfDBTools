package repository

import (
	"fmt"
	"os/exec"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
)

// Manager handles MariaDB repository setup using official script
type Manager struct {
	osInfo *system.OSInfo
}

// NewManager creates a new repository manager
func NewManager(osInfo *system.OSInfo) *Manager {
	return &Manager{osInfo: osInfo}
}

// SetupOfficial sets up MariaDB repository using official script
func (m *Manager) SetupOfficial(version string) error {
	lg, _ := logger.Get()

	lg.Info("Setting up MariaDB repository using official script",
		logger.String("version", version),
		logger.String("os", m.osInfo.ID))

	// Use the correct official MariaDB script URL
	scriptURL := "https://r.mariadb.com/downloads/mariadb_repo_setup"

	// Download and run the official MariaDB repository setup script
	// Format version properly for the script
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf("curl -LsSf %s | sudo bash -s -- --mariadb-server-version=%s", scriptURL, version))

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to setup repository",
			logger.String("output", string(output)),
			logger.Error(err))
		return fmt.Errorf("repository setup failed: %w", err)
	}

	lg.Info("Repository setup completed successfully")
	return nil
}

// IsAvailable checks if the official script is accessible
func (m *Manager) IsAvailable() (bool, error) {
	// Test if we can reach the MariaDB script using the correct URL
	cmd := exec.Command("curl", "-I", "-s", "--connect-timeout", "10",
		"https://r.mariadb.com/downloads/mariadb_repo_setup")

	err := cmd.Run()
	return err == nil, err
}

// Clean removes existing MariaDB repositories
func (m *Manager) Clean() error {
	lg, _ := logger.Get()

	switch m.osInfo.PackageType {
	case "rpm":
		return m.cleanRPMRepos()
	case "deb":
		return m.cleanDEBRepos()
	default:
		lg.Warn("Unknown package type, skipping repository cleanup",
			logger.String("type", m.osInfo.PackageType))
		return nil
	}
}

// cleanRPMRepos removes RPM MariaDB repositories
func (m *Manager) cleanRPMRepos() error {
	// Remove MariaDB repository files
	cmd := exec.Command("rm", "-f", "/etc/yum.repos.d/mariadb*.repo")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clean RPM repositories: %w", err)
	}
	return nil
}

// cleanDEBRepos removes DEB MariaDB repositories
func (m *Manager) cleanDEBRepos() error {
	// Remove MariaDB sources
	cmd := exec.Command("rm", "-f", "/etc/apt/sources.list.d/mariadb*.list")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clean DEB repositories: %w", err)
	}

	// Remove GPG key
	exec.Command("rm", "-f", "/etc/apt/keyrings/mariadb-keyring.gpg").Run()
	return nil
}

// UpdateCache updates package manager cache
func (m *Manager) UpdateCache() error {
	lg, _ := logger.Get()

	switch m.osInfo.PackageType {
	case "rpm":
		cmd := exec.Command("yum", "makecache")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update YUM cache: %w", err)
		}
	case "deb":
		cmd := exec.Command("apt-get", "update")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update APT cache: %w", err)
		}
	}

	lg.Info("Package cache updated successfully")
	return nil
}
