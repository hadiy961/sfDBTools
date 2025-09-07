package install

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/repository"
	"sfDBTools/utils/terminal"
)

// RepositorySetupManager handles MariaDB repository setup operations
type RepositorySetupManager struct {
	repoManager *repository.Manager
}

// NewRepositorySetupManager creates a new repository setup manager instance
func NewRepositorySetupManager(repoManager *repository.Manager) *RepositorySetupManager {
	return &RepositorySetupManager{
		repoManager: repoManager,
	}
}

// SetupRepository sets up MariaDB repository for the selected version
func (rsm *RepositorySetupManager) SetupRepository(version string) error {
	lg, _ := logger.Get()

	spinner := terminal.NewProcessingSpinner("Setting up MariaDB repository...")
	spinner.Start()

	// Clean existing repositories first
	if err := rsm.cleanExistingRepositories(spinner); err != nil {
		lg.Warn("Failed to clean existing repositories", logger.Error(err))
	}

	// Setup official repository
	if err := rsm.setupOfficialRepository(spinner, version); err != nil {
		spinner.StopWithError("Failed to setup repository")
		return fmt.Errorf("failed to setup repository: %w", err)
	}

	// Update package cache
	if err := rsm.updatePackageCache(spinner); err != nil {
		spinner.StopWithError("Failed to update package cache")
		return fmt.Errorf("failed to update package cache: %w", err)
	}

	spinner.StopWithSuccess("Repository setup completed")
	return nil
}

// cleanExistingRepositories removes existing MariaDB repositories
func (rsm *RepositorySetupManager) cleanExistingRepositories(spinner *terminal.ProgressSpinner) error {
	spinner.UpdateMessage("Cleaning existing repositories...")
	return rsm.repoManager.Clean()
}

// setupOfficialRepository sets up the official MariaDB repository
func (rsm *RepositorySetupManager) setupOfficialRepository(spinner *terminal.ProgressSpinner, version string) error {
	spinner.UpdateMessage(fmt.Sprintf("Setting up official MariaDB %s repository...", version))
	return rsm.repoManager.SetupOfficial(version)
}

// updatePackageCache updates the package manager cache
func (rsm *RepositorySetupManager) updatePackageCache(spinner *terminal.ProgressSpinner) error {
	spinner.UpdateMessage("Updating package cache...")
	return rsm.repoManager.UpdateCache()
}
