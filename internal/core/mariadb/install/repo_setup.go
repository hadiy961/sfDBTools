package install

import (
	"sfDBTools/utils/common"
	"sfDBTools/utils/repository"
)

// RepoSetupManager handles MariaDB repository setup (simplified)
type RepoSetupManager struct {
	repoManager *repository.Manager
}

// NewRepoSetupManager creates a new repository setup manager
func NewRepoSetupManager(osInfo *common.OSInfo) *RepoSetupManager {
	return &RepoSetupManager{
		repoManager: repository.NewManager(osInfo),
	}
}

// SetupRepository sets up MariaDB repository using official script
func (r *RepoSetupManager) SetupRepository(version string) error {
	return r.repoManager.SetupOfficial(version)
}

// CleanRepositories removes existing MariaDB repositories
func (r *RepoSetupManager) CleanRepositories() error {
	return r.repoManager.Clean()
}

// IsScriptAvailable checks if MariaDB setup script is accessible
func (r *RepoSetupManager) IsScriptAvailable() (bool, error) {
	return r.repoManager.IsAvailable()
}

// UpdatePackageCache updates package manager cache
func (r *RepoSetupManager) UpdatePackageCache() error {
	return r.repoManager.UpdateCache()
}
