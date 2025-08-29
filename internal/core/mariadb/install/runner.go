package install

import (
	"fmt"
	"strconv"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/terminal"
)

// InstallRunner orchestrates the MariaDB installation process
type InstallRunner struct {
	config             *InstallConfig
	precheckManager    *PrecheckManager
	versionManager     *VersionManager
	installManager     *InstallManager
	postInstallManager *PostInstallManager
	dataManager        *DataManager
	osInfo             *common.OSInfo
	selectedVersion    *SelectableVersion
}

// NewInstallRunner creates a new installation runner
func NewInstallRunner(config *InstallConfig) *InstallRunner {
	if config == nil {
		config = DefaultInstallConfig()
	}

	return &InstallRunner{
		config:          config,
		precheckManager: NewPrecheckManager(),
		versionManager:  NewVersionManager(),
	}
}

// Run executes the complete MariaDB installation process
func (r *InstallRunner) Run() error {
	lg, _ := logger.Get()

	lg.Info("Starting MariaDB installation process")
	terminal.PrintHeader("MariaDB Installation")

	// Step 1: Check OS compatibility
	if err := r.checkOSCompatibility(); err != nil {
		return fmt.Errorf("OS compatibility check failed: %w", err)
	}

	// Step 2: Check internet connectivity
	if err := r.checkInternetConnectivity(); err != nil {
		return fmt.Errorf("internet connectivity check failed: %w", err)
	}

	// Step 3: Fetch available versions
	if err := r.fetchAvailableVersions(); err != nil {
		return fmt.Errorf("failed to fetch available versions: %w", err)
	}

	// Step 4: Select version
	if err := r.selectVersion(); err != nil {
		return fmt.Errorf("version selection failed: %w", err)
	}

	// Step 5: Check for existing installation
	if err := r.checkExistingInstallation(); err != nil {
		return fmt.Errorf("existing installation check failed: %w", err)
	}

	// Step 6: Configure repository
	if err := r.configureRepository(); err != nil {
		return fmt.Errorf("repository configuration failed: %w", err)
	}

	// Step 7: Install MariaDB
	if err := r.installMariaDB(); err != nil {
		return fmt.Errorf("MariaDB installation failed: %w", err)
	}

	// Step 8: Post-installation setup
	if err := r.postInstallationSetup(); err != nil {
		terminal.PrintWarning("Post-installation setup had issues, but MariaDB was installed successfully")
		lg.Warn("Post-installation setup failed", logger.Error(err))
	}

	terminal.PrintSuccess("MariaDB installation completed successfully!")
	lg.Info("MariaDB installation completed successfully",
		logger.String("version", r.selectedVersion.LatestVersion))

	return nil
}

// checkOSCompatibility checks if the OS is supported
func (r *InstallRunner) checkOSCompatibility() error {
	osInfo, err := r.precheckManager.CheckOSCompatibility()
	if err != nil {
		return err
	}
	r.osInfo = osInfo
	return nil
}

// checkInternetConnectivity verifies internet connection
func (r *InstallRunner) checkInternetConnectivity() error {
	return r.precheckManager.CheckInternetConnectivity()
}

// fetchAvailableVersions retrieves available MariaDB versions
func (r *InstallRunner) fetchAvailableVersions() error {
	return r.versionManager.FetchAvailableVersions()
}

// selectVersion handles version selection
func (r *InstallRunner) selectVersion() error {
	selectedVersion, err := r.versionManager.SelectVersion(r.config.AutoConfirm, r.config.Version)
	if err != nil {
		return err
	}
	r.selectedVersion = selectedVersion
	return nil
}

// checkExistingInstallation checks if MariaDB is already installed
func (r *InstallRunner) checkExistingInstallation() error {
	r.installManager = NewInstallManager(r.config, r.osInfo, r.selectedVersion)
	return r.installManager.CheckExistingInstallation()
}

// configureRepository sets up MariaDB repository
func (r *InstallRunner) configureRepository() error {
	return r.installManager.ConfigureRepository()
}

// installMariaDB performs the actual installation
func (r *InstallRunner) installMariaDB() error {
	return r.installManager.InstallMariaDB()
}

// postInstallationSetup performs post-installation configuration
func (r *InstallRunner) postInstallationSetup() error {
	r.dataManager = NewDataManager(r.selectedVersion)
	r.postInstallManager = NewPostInstallManager(r.config, r.selectedVersion)
	return r.postInstallManager.PerformPostInstallationSetup(r.dataManager)
}

// GetInstallationSteps returns the list of installation steps for progress tracking
func (r *InstallRunner) GetInstallationSteps() []InstallationStep {
	return []InstallationStep{
		{Name: "os_check", Description: "Checking OS compatibility", Required: true},
		{Name: "internet_check", Description: "Checking internet connectivity", Required: true},
		{Name: "fetch_versions", Description: "Fetching available versions", Required: true},
		{Name: "select_version", Description: "Selecting MariaDB version", Required: true},
		{Name: "check_existing", Description: "Checking existing installation", Required: true},
		{Name: "configure_repo", Description: "Configuring repository", Required: true},
		{Name: "install", Description: "Installing MariaDB", Required: true},
		{Name: "post_setup", Description: "Post-installation setup", Required: false},
	}
}

// parseVersionParts extracts major and minor version numbers
func parseVersionParts(parts []string) (int, int) {
	major, minor := 0, 0

	if len(parts) >= 1 {
		if val, err := strconv.Atoi(parts[0]); err == nil {
			major = val
		}
	}
	if len(parts) >= 2 {
		if val, err := strconv.Atoi(parts[1]); err == nil {
			minor = val
		}
	}

	return major, minor
}
