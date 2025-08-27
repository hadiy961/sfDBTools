package configure

import (
	"fmt"

	"sfDBTools/internal/config"
	"sfDBTools/internal/config/model"
	"sfDBTools/internal/core/mariadb"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// ConfigureRunner orchestrates the MariaDB configuration process
type ConfigureRunner struct {
	config   *ConfigureConfig
	settings *MariaDBSettings
}

// NewConfigureRunner creates a new configure runner
func NewConfigureRunner(config *ConfigureConfig) *ConfigureRunner {
	return &ConfigureRunner{
		config: config,
	}
}

// Run executes the complete MariaDB configuration process
func (r *ConfigureRunner) Run() error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Starting MariaDB configuration process...")
	lg.Info("MariaDB configuration process started")

	// Step 1: Validate OS
	if err := r.validateOS(); err != nil {
		return fmt.Errorf("OS validation failed: %w", err)
	}

	// Step 2: Check and stop MariaDB service
	if err := r.stopMariaDBService(); err != nil {
		return fmt.Errorf("failed to stop MariaDB service: %w", err)
	}

	// Step 3: Load and validate configuration
	appConfig, err := r.loadConfiguration()
	if err != nil {
		return fmt.Errorf("configuration loading failed: %w", err)
	}

	// Step 4: Prompt user for settings
	if err := r.promptUserSettings(appConfig); err != nil {
		return fmt.Errorf("user prompt failed: %w", err)
	}

	// Step 5: Setup directories
	if err := r.setupDirectories(); err != nil {
		return fmt.Errorf("directory setup failed: %w", err)
	}

	// Step 6: Process configuration file
	if err := r.processConfigFile(appConfig); err != nil {
		return fmt.Errorf("config file processing failed: %w", err)
	}

	// Step 7: Configure systemd
	if err := r.configureSystemd(); err != nil {
		return fmt.Errorf("systemd configuration failed: %w", err)
	}

	// Step 8: Setup firewall
	if err := r.setupFirewall(); err != nil {
		return fmt.Errorf("firewall setup failed: %w", err)
	}

	// Step 9: Migrate data
	if err := r.migrateData(); err != nil {
		return fmt.Errorf("data migration failed: %w", err)
	}

	// Step 10: Configure SELinux (if applicable)
	if err := r.configureSELinux(); err != nil {
		return fmt.Errorf("SELinux configuration failed: %w", err)
	}

	// Step 11: Start MariaDB service
	if err := r.startMariaDBService(); err != nil {
		return fmt.Errorf("failed to start MariaDB service: %w", err)
	}

	// Step 12: Setup databases and users
	if !r.config.SkipUserSetup && !r.config.SkipDBSetup {
		if err := r.setupDatabasesAndUsers(appConfig); err != nil {
			return fmt.Errorf("database setup failed: %w", err)
		}
	}

	lg.Info("MariaDB configuration completed successfully")
	terminal.PrintSuccess("MariaDB configuration completed successfully!")
	return nil
}

// validateOS validates the operating system
func (r *ConfigureRunner) validateOS() error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Validating operating system...")

	if err := mariadb.ValidateOperatingSystem(); err != nil {
		lg.Error("OS validation failed", logger.Error(err))
		return err
	}

	terminal.PrintSuccess("Operating system validated")
	lg.Info("Operating system validation passed")
	return nil
}

// stopMariaDBService stops the MariaDB service
func (r *ConfigureRunner) stopMariaDBService() error {
	lg, _ := logger.Get()

	serviceManager := NewServiceManager()

	// Check if service is running
	if !serviceManager.IsServiceRunning() {
		lg.Info("MariaDB service is not running")
		terminal.PrintInfo("MariaDB service is not running")
		return nil
	}

	// Stop the service
	if err := serviceManager.StopService(); err != nil {
		return err
	}

	return nil
}

// loadConfiguration loads and validates application configuration
func (r *ConfigureRunner) loadConfiguration() (*model.Config, error) {
	lg, _ := logger.Get()

	terminal.PrintInfo("Loading configuration...")

	appConfig, err := config.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	if appConfig == nil {
		return nil, fmt.Errorf("application configuration not loaded")
	}

	// Validate required configuration paths
	if appConfig.ConfigDir.MariaDBConfig == "" {
		return nil, fmt.Errorf("mariadb_config path is required in configuration")
	}

	if appConfig.ConfigDir.MariaDBKey == "" {
		return nil, fmt.Errorf("mariadb_key path is required in configuration")
	}

	lg.Info("Configuration loaded and validated successfully")
	terminal.PrintSuccess("Configuration loaded successfully")
	return appConfig, nil
}

// promptUserSettings prompts user for MariaDB settings
func (r *ConfigureRunner) promptUserSettings(appConfig *model.Config) error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Collecting MariaDB configuration settings...")

	// Create default settings from config
	defaults := DefaultMariaDBSettings(appConfig)

	// Prompt user for settings
	prompt := NewUserPrompt(defaults)
	settings, err := prompt.PromptForSettings(r.config.AutoConfirm)
	if err != nil {
		return err
	}

	r.settings = settings

	lg.Info("User settings collected successfully")
	terminal.PrintSuccess("Configuration settings collected")
	return nil
}

// setupDirectories creates and configures required directories
func (r *ConfigureRunner) setupDirectories() error {
	terminal.PrintInfo("Setting up directories...")

	dirManager := NewDirectoryManager(r.settings)
	return dirManager.SetupDirectories()
}

// processConfigFile processes and deploys the configuration file
func (r *ConfigureRunner) processConfigFile(appConfig *model.Config) error {
	terminal.PrintInfo("Processing MariaDB configuration file...")

	// Determine target path based on OS (CentOS/RHEL pattern)
	targetPath := "/etc/my.cnf.d/server.cnf"

	configManager := NewConfigFileManager(r.settings, appConfig.ConfigDir.MariaDBConfig, targetPath)
	return configManager.ProcessConfigFile()
}

// configureSystemd configures systemd service
func (r *ConfigureRunner) configureSystemd() error {
	systemdManager := NewSystemdManager()
	return systemdManager.ConfigureService()
}

// setupFirewall configures firewall for MariaDB port
func (r *ConfigureRunner) setupFirewall() error {
	firewallManager := NewFirewallManager(r.settings.Port)
	return firewallManager.ConfigureFirewall()
}

// migrateData migrates data from default location to new location
func (r *ConfigureRunner) migrateData() error {
	terminal.PrintInfo("Migrating MariaDB data...")

	sourceDir := "/var/lib/mysql"
	dataMigrator := NewDataMigrator(sourceDir, r.settings.DataDir)
	return dataMigrator.MigrateData()
}

// configureSELinux configures SELinux contexts
func (r *ConfigureRunner) configureSELinux() error {
	selinuxManager := NewSELinuxManager(r.settings.DataDir)
	return selinuxManager.ConfigureSELinux()
}

// startMariaDBService starts and enables MariaDB service
func (r *ConfigureRunner) startMariaDBService() error {
	serviceManager := NewServiceManager()

	// Start service
	if err := serviceManager.StartService(); err != nil {
		return err
	}

	// Enable service on boot
	if err := serviceManager.EnableService(); err != nil {
		return err
	}

	// Show status
	return serviceManager.GetServiceStatus()
}

// setupDatabasesAndUsers creates default databases and users
func (r *ConfigureRunner) setupDatabasesAndUsers(appConfig *model.Config) error {
	terminal.PrintInfo("Setting up default databases and users...")

	clientCode := appConfig.General.ClientCode
	if clientCode == "" {
		clientCode = "default"
	}

	dbManager := NewDatabaseManager(r.settings, clientCode)
	return dbManager.SetupDatabasesAndUsers()
}
