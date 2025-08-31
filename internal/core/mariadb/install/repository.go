package install

import (
	"sfDBTools/utils/common"
	"sfDBTools/utils/repository"
)

// RepositoryManager handles MariaDB repository configuration
type RepositoryManager struct {
	manager *repository.Manager
}

// NewRepositoryManager creates a new repository manager
func NewRepositoryManager(osInfo *common.OSInfo) *RepositoryManager {
	return &RepositoryManager{
		manager: repository.NewManager(osInfo),
	}
}

// SetupRepository sets up MariaDB repository using official script
func (r *RepositoryManager) SetupRepository(version string) error {
	return r.manager.SetupOfficial(version)
}

// IsScriptAvailable checks if MariaDB setup script is accessible
func (r *RepositoryManager) IsScriptAvailable() (bool, error) {
	return r.manager.IsAvailable()
}

// CleanRepositories removes existing MariaDB repositories
func (r *RepositoryManager) CleanRepositories() error {
	return r.manager.Clean()
}

// UpdatePackageCache updates package manager cache
func (r *RepositoryManager) UpdatePackageCache() error {
	return r.manager.UpdateCache()
}
