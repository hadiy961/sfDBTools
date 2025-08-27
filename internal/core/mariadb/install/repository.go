package install

import (
	"fmt"
	"os"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
)

// RepositoryManager handles MariaDB repository configuration
type RepositoryManager struct {
	osInfo *common.OSInfo
}

// NewRepositoryManager creates a new repository manager
func NewRepositoryManager(osInfo *common.OSInfo) *RepositoryManager {
	return &RepositoryManager{
		osInfo: osInfo,
	}
}

// GetRepositoryConfig returns repository configuration for the OS and version
func (r *RepositoryManager) GetRepositoryConfig(version string) (*RepositoryConfig, error) {
	lg, _ := logger.Get()

	lg.Info("Getting repository configuration",
		logger.String("os", r.osInfo.ID),
		logger.String("version", version),
		logger.String("arch", r.osInfo.Architecture))

	switch r.osInfo.PackageType {
	case "rpm":
		return r.getRPMRepositoryConfig(version)
	case "deb":
		return r.getDEBRepositoryConfig(version)
	default:
		return nil, fmt.Errorf("unsupported package type: %s", r.osInfo.PackageType)
	}
}

// getRPMRepositoryConfig returns repository configuration for RPM-based systems
func (r *RepositoryManager) getRPMRepositoryConfig(version string) (*RepositoryConfig, error) {
	// Use the official MariaDB setup script URL format
	// This follows the same pattern as: curl -sS https://downloads.mariadb.com/MariaDB/mariadb_repo_setup | sudo bash

	var repoName string
	var baseURL string

	// Different URL format based on the OS and version
	switch r.osInfo.ID {
	case "centos", "rhel":
		repoName = fmt.Sprintf("mariadb-%s", version)
		baseURL = fmt.Sprintf("https://rpm.mariadb.org/%s/el/%s/$basearch", version, r.getELMajorVersion())
	case "rocky", "almalinux":
		repoName = fmt.Sprintf("mariadb-%s", version)
		baseURL = fmt.Sprintf("https://rpm.mariadb.org/%s/el/%s/$basearch", version, r.getELMajorVersion())
	default:
		return nil, fmt.Errorf("unsupported RPM-based OS: %s", r.osInfo.ID)
	}

	return &RepositoryConfig{
		Name:     repoName,
		BaseURL:  baseURL,
		GPGKey:   "https://rpm.mariadb.org/RPM-GPG-KEY-MariaDB",
		Priority: 10,
	}, nil
}

// getDEBRepositoryConfig returns repository configuration for DEB-based systems
func (r *RepositoryManager) getDEBRepositoryConfig(version string) (*RepositoryConfig, error) {
	var codename string

	// Map Ubuntu version to codename
	switch r.osInfo.VersionID {
	case "18.04":
		codename = "bionic"
	case "20.04":
		codename = "focal"
	case "22.04":
		codename = "jammy"
	case "24.04":
		codename = "noble"
	default:
		// Try to detect codename from VERSION_CODENAME or use generic
		codename = "jammy" // fallback to most common LTS
	}

	// Use the official MariaDB apt repository
	baseURL := fmt.Sprintf("deb [arch=%s signed-by=/etc/apt/keyrings/mariadb-keyring.gpg] https://deb.mariadb.org/%s/ubuntu %s main",
		r.getDebArch(), version, codename)

	return &RepositoryConfig{
		Name:     fmt.Sprintf("mariadb-%s", version),
		BaseURL:  baseURL,
		GPGKey:   "https://deb.mariadb.org/PublicKey",
		Priority: 10,
	}, nil
}

// getELMajorVersion extracts major version from Enterprise Linux version
func (r *RepositoryManager) getELMajorVersion() string {
	if r.osInfo.VersionID == "" {
		return "8" // fallback
	}

	// Extract first character (major version)
	if len(r.osInfo.VersionID) > 0 {
		return string(r.osInfo.VersionID[0])
	}

	return "8"
}

// getDebArch converts system architecture to Debian architecture
func (r *RepositoryManager) getDebArch() string {
	switch r.osInfo.Architecture {
	case "x86_64":
		return "amd64"
	case "aarch64":
		return "arm64"
	default:
		return "amd64" // fallback
	}
}

// IsRepositoryConfigured checks if MariaDB repository is already configured
func (r *RepositoryManager) IsRepositoryConfigured(version string) (bool, error) {
	lg, _ := logger.Get()

	switch r.osInfo.PackageType {
	case "rpm":
		return r.isRPMRepositoryConfigured(version)
	case "deb":
		return r.isDEBRepositoryConfigured(version)
	default:
		lg.Warn("Unknown package type for repository check", logger.String("type", r.osInfo.PackageType))
		return false, nil
	}
}

// isRPMRepositoryConfigured checks if RPM repository is configured
func (r *RepositoryManager) isRPMRepositoryConfigured(version string) (bool, error) {
	// Check if repository file exists
	repoFile := fmt.Sprintf("/etc/yum.repos.d/mariadb-%s.repo", version)

	if _, err := os.Stat(repoFile); os.IsNotExist(err) {
		return false, nil
	}

	return true, nil
}

// isDEBRepositoryConfigured checks if DEB repository is configured
func (r *RepositoryManager) isDEBRepositoryConfigured(version string) (bool, error) {
	// This is more complex for APT - we'd need to check sources.list and sources.list.d
	// For now, return false to always configure
	return false, nil
}
