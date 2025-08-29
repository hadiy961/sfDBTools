package install

import (
	"fmt"
	"strings"

	"sfDBTools/internal/core/mariadb/configure"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// PostInstallManager handles post-installation setup
type PostInstallManager struct {
	config          *InstallConfig
	selectedVersion *SelectableVersion
	serviceManager  *ServiceManager
	configFixer     *ConfigFixer
}

// NewPostInstallManager creates a new post-install manager
func NewPostInstallManager(config *InstallConfig, selectedVersion *SelectableVersion) *PostInstallManager {
	return &PostInstallManager{
		config:          config,
		selectedVersion: selectedVersion,
		serviceManager:  NewServiceManager(),
		configFixer:     NewConfigFixer(),
	}
}

// PerformPostInstallationSetup performs post-installation configuration
func (p *PostInstallManager) PerformPostInstallationSetup(dataManager *DataManager) error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Performing post-installation setup...")

	// Step 1: Fix common configuration issues after installation
	if err := p.configFixer.FixPostInstallationIssues("/var/lib/mysql"); err != nil {
		lg.Warn("Failed to fix post-installation issues", logger.Error(err))
	}

	// Step 2: Check for data directory conflicts before starting service
	if err := dataManager.CheckDataDirectoryCompatibility(); err != nil {
		lg.Warn("Data directory compatibility issue detected", logger.Error(err))
		terminal.PrintWarning("Data directory compatibility issue detected")

		if err := dataManager.HandleDataDirectoryConflict(p.config.AutoConfirm); err != nil {
			return fmt.Errorf("failed to resolve data directory conflict: %w", err)
		}
	}

	// Start MariaDB service
	if p.config.StartService {
		if err := p.serviceManager.StartMariaDBService(); err != nil {
			lg.Error("Failed to start MariaDB service", logger.Error(err))
			return fmt.Errorf("failed to start MariaDB service: %w", err)
		}
		terminal.PrintSuccess("MariaDB service started successfully")
	}

	// Enable service on boot
	if err := p.serviceManager.EnableMariaDBService(); err != nil {
		lg.Warn("Failed to enable MariaDB service on boot", logger.Error(err))
		terminal.PrintWarning("MariaDB service may not start automatically on boot")
	} else {
		terminal.PrintSuccess("MariaDB service enabled on boot")
	}

	// Run security setup if enabled
	if p.config.EnableSecurity {
		terminal.PrintInfo("Security setup will need to be run manually using: mysql_secure_installation")
	}

	// Ask if user wants to configure MariaDB with custom settings
	if p.shouldRunConfiguration() {
		if err := p.runMariaDBConfiguration(); err != nil {
			lg.Warn("MariaDB configuration failed", logger.Error(err))
			terminal.PrintWarning("MariaDB configuration had issues, but installation was successful")
			terminal.PrintInfo("You can run configuration manually using: sfdbtools mariadb configure")
		} else {
			terminal.PrintSuccess("MariaDB configuration completed successfully")
		}
	} else {
		terminal.PrintInfo("MariaDB configuration skipped")
		terminal.PrintInfo("You can run configuration later using: sfdbtools mariadb configure")
	}

	lg.Info("Post-installation setup completed")
	return nil
}

// shouldRunConfiguration asks user if they want to configure MariaDB
func (p *PostInstallManager) shouldRunConfiguration() bool {
	// Skip prompt if auto-confirm is enabled (don't auto-configure)
	if p.config.AutoConfirm {
		return false // Let user run configuration manually later
	}

	terminal.PrintInfo("\nMariaDB has been installed successfully!")
	terminal.PrintInfo("Would you like to configure MariaDB with custom settings now?")
	terminal.PrintInfo("This will setup:")
	terminal.PrintInfo("  • Custom data and binlog directories")
	terminal.PrintInfo("  • Configuration file optimization")
	terminal.PrintInfo("  • Firewall and SELinux settings")
	terminal.PrintInfo("  • Default databases and users")

	fmt.Print("Run MariaDB configuration now? (y/N): ")

	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// runMariaDBConfiguration executes the MariaDB configuration process
func (p *PostInstallManager) runMariaDBConfiguration() error {
	lg, _ := logger.Get()

	lg.Info("Starting MariaDB configuration from install process")
	terminal.PrintInfo("Configuring MariaDB with optimized settings...")

	// Create configure config with auto-confirm mode to avoid double prompting
	configureConfig := &configure.ConfigureConfig{
		AutoConfirm:   false, // Auto-confirm since user already agreed
		SkipUserSetup: false, // Include user setup
		SkipDBSetup:   false, // Include database setup
	}

	// Create configure runner
	configRunner := configure.NewConfigureRunner(configureConfig)

	// Run configuration
	if err := configRunner.Run(); err != nil {
		return fmt.Errorf("MariaDB configuration failed: %w", err)
	}

	lg.Info("MariaDB configuration completed from install process")
	return nil
}
