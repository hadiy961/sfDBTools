package configure

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"sfDBTools/internal/config"
	"sfDBTools/internal/config/model"
	"sfDBTools/internal/core/mariadb"
	"sfDBTools/internal/core/mariadb/remove"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
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

	// Step 5: Create target directories (without setting ownership yet)
	if err := r.createTargetDirectories(); err != nil {
		return fmt.Errorf("target directory creation failed: %w", err)
	}

	// Step 6: Migrate data - now that target directories exist
	if err := r.migrateData(); err != nil {
		return fmt.Errorf("data migration failed: %w", err)
	}

	// Step 7: Setup directories ownership (after migration)
	if err := r.setupDirectories(); err != nil {
		return fmt.Errorf("directory setup failed: %w", err)
	}

	// Step 8: Process configuration file
	if err := r.processConfigFile(appConfig); err != nil {
		return fmt.Errorf("config file processing failed: %w", err)
	}

	// Step 9: Configure systemd
	if err := r.configureSystemd(); err != nil {
		return fmt.Errorf("systemd configuration failed: %w", err)
	}

	// Step 10: Setup firewall
	if err := r.setupFirewall(); err != nil {
		return fmt.Errorf("firewall setup failed: %w", err)
	}

	// Step 11: Configure SELinux (if applicable)
	if err := r.configureSELinux(); err != nil {
		return fmt.Errorf("SELinux configuration failed: %w", err)
	}

	// Skip database initialization and user setup in migration-only mode
	if r.config.MigrationOnly {
		lg.Info("Migration-only mode: skipping database initialization and user setup")
		terminal.PrintInfo("Migration-only mode: starting service with migrated configuration...")

		// Step 13: Start MariaDB service to validate new configuration
		if err := r.startMariaDBService(); err != nil {
			return fmt.Errorf("failed to start MariaDB service with new configuration: %w", err)
		}

		terminal.PrintSuccess("Directory migration completed successfully!")
		terminal.PrintInfo("MariaDB service started with new directory configuration")
		lg.Info("Migration-only mode completed: directories migrated and service started")
		return nil
	}

	// Step 12: Initialize database if needed
	if err := r.initializeDatabaseIfNeeded(); err != nil {
		return fmt.Errorf("database initialization failed: %w", err)
	}

	// Step 13: Start MariaDB service
	if err := r.startMariaDBService(); err != nil {
		return fmt.Errorf("failed to start MariaDB service: %w", err)
	}

	// Step 14: Setup databases and users
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

	spinner := terminal.NewProgressSpinner("Validating operating system...")
	spinner.Start()

	if err := mariadb.ValidateOperatingSystem(); err != nil {
		spinner.Stop()
		lg.Error("OS validation failed", logger.Error(err))
		terminal.PrintError("Operating system validation failed")
		return err
	}

	spinner.Stop()
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

	spinner := terminal.NewProgressSpinner("Stopping MariaDB service...")
	spinner.Start()

	// Stop the service
	if err := serviceManager.StopService(); err != nil {
		spinner.Stop()
		return err
	}

	spinner.Stop()
	terminal.PrintSuccess("MariaDB service stopped")
	return nil
}

// loadConfiguration loads and validates application configuration
func (r *ConfigureRunner) loadConfiguration() (*model.Config, error) {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Loading and validating configuration...")
	spinner.Start()

	appConfig, err := config.Get()
	if err != nil {
		spinner.Stop()
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	if appConfig == nil {
		spinner.Stop()
		return nil, fmt.Errorf("application configuration not loaded")
	}

	// Validate required configuration paths
	if appConfig.ConfigDir.MariaDBConfig == "" {
		spinner.Stop()
		return nil, fmt.Errorf("mariadb_config path is required in configuration")
	}

	if appConfig.ConfigDir.MariaDBKey == "" {
		spinner.Stop()
		return nil, fmt.Errorf("mariadb_key path is required in configuration")
	}

	spinner.Stop()
	lg.Info("Configuration loaded and validated successfully")
	terminal.PrintSuccess("Configuration loaded successfully")
	return appConfig, nil
}

// promptUserSettings prompts user for MariaDB settings
func (r *ConfigureRunner) promptUserSettings(appConfig *model.Config) error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Collecting MariaDB configuration settings...")

	// Create dynamic default settings from existing config or app config
	defaults, err := CreateDynamicDefaults(appConfig)
	if err != nil {
		return fmt.Errorf("failed to create dynamic defaults: %w", err)
	}

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
	spinner := terminal.NewProgressSpinner("Setting up directory ownership and permissions...")
	spinner.Start()

	dirManager := NewDirectoryManager(r.settings)
	if err := dirManager.SetupDirectories(); err != nil {
		spinner.Stop()
		return err
	}

	spinner.Stop()
	terminal.PrintSuccess("Directory ownership configured")
	return nil
}

// createTargetDirectories creates target directories without setting ownership (for migration)
func (r *ConfigureRunner) createTargetDirectories() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Creating target directories...")
	spinner.Start()

	directories := []string{
		r.settings.DataDir,
		r.settings.BinlogDir,
		r.settings.LogDir,
	}

	// Create directories without setting ownership yet
	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			spinner.Stop()
			lg.Error("Failed to create directory",
				logger.String("directory", dir),
				logger.Error(err))
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		lg.Info("Directory created successfully", logger.String("path", dir))
	}

	spinner.Stop()
	lg.Info("All target directories created successfully")
	terminal.PrintSuccess("Target directories created")
	return nil
}

// processConfigFile processes and deploys the configuration file
func (r *ConfigureRunner) processConfigFile(appConfig *model.Config) error {
	spinner := terminal.NewProgressSpinner("Processing MariaDB configuration file...")
	spinner.Start()

	// Determine target path based on OS (CentOS/RHEL pattern)
	targetPath := "/etc/my.cnf.d/server.cnf"

	configManager := NewConfigFileManager(r.settings, appConfig.ConfigDir.MariaDBConfig, targetPath)
	if err := configManager.ProcessConfigFile(); err != nil {
		spinner.Stop()
		return err
	}

	spinner.Stop()
	terminal.PrintSuccess("Configuration file processed")
	return nil
}

// configureSystemd configures systemd service
func (r *ConfigureRunner) configureSystemd() error {
	spinner := terminal.NewProgressSpinner("Configuring systemd service...")
	spinner.Start()

	systemdManager := NewSystemdManager()
	if err := systemdManager.ConfigureService(); err != nil {
		spinner.Stop()
		return err
	}

	spinner.Stop()
	terminal.PrintSuccess("Systemd service configured")
	return nil
}

// setupFirewall configures firewall for MariaDB port
func (r *ConfigureRunner) setupFirewall() error {
	spinner := terminal.NewProgressSpinner("Setting up firewall rules...")
	spinner.Start()

	firewallManager := NewFirewallManager(r.settings.Port)
	if err := firewallManager.ConfigureFirewall(); err != nil {
		spinner.Stop()
		return err
	}

	spinner.Stop()
	terminal.PrintSuccess("Firewall configured")
	return nil
}

// migrateData migrates data from current active location to new location and cleans up old directories
func (r *ConfigureRunner) migrateData() error {
	spinner := terminal.NewProgressSpinner("Detecting and migrating MariaDB directories...")
	spinner.Start()

	// Use detection service to find current MariaDB directories
	osDetector := common.NewOSDetector()
	osInfo, err := osDetector.DetectOS()
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to detect OS: %w", err)
	}

	detectionService := remove.NewDetectionService(osInfo)
	installation, err := detectionService.DetectInstallation()
	if err != nil {
		// If detection fails, fallback to config defaults
		spinner.Stop()
		terminal.PrintWarning("Could not detect existing MariaDB installation, using config defaults")
		return r.migrateFromConfigDefaults()
	}

	// Get actual directories from detection
	sourceDataDir := installation.ActualDataDir
	sourceBinlogDir := installation.ActualBinlogDir
	sourceLogDir := installation.ActualLogDir

	// Fallback to standard paths if detection didn't find specific paths or found invalid paths
	if sourceDataDir == "" || sourceDataDir == "." || !filepath.IsAbs(sourceDataDir) {
		sourceDataDir = "/var/lib/mysql"
	}
	if sourceBinlogDir == "" || sourceBinlogDir == "." || !filepath.IsAbs(sourceBinlogDir) {
		sourceBinlogDir = "/var/lib/mysqlbinlogs"
	}
	if sourceLogDir == "" || sourceLogDir == "." || !filepath.IsAbs(sourceLogDir) {
		sourceLogDir = "/var/log/mysql"
	}

	// Migrate data directory
	if sourceDataDir != r.settings.DataDir {
		dataMigrator := NewDataMigratorWithCleanup(sourceDataDir, r.settings.DataDir)
		if err := dataMigrator.MigrateData(); err != nil {
			spinner.Stop()
			return fmt.Errorf("failed to migrate data directory: %w", err)
		}
	}

	// Migrate binlog directory
	if sourceBinlogDir != r.settings.BinlogDir {
		binlogMigrator := NewDataMigratorWithCleanup(sourceBinlogDir, r.settings.BinlogDir)
		if err := binlogMigrator.MigrateBinlogData(); err != nil {
			// Log warning but don't fail - binlog directory might not exist
			terminal.PrintWarning("Failed to migrate binlog directory (this is usually okay)")
		}
	}

	// Migrate log directory (if it contains important logs)
	if sourceLogDir != r.settings.LogDir {
		logMigrator := NewDataMigratorWithCleanup(sourceLogDir, r.settings.LogDir)
		if err := logMigrator.MigrateData(); err != nil {
			// Log warning but don't fail - log directory might not exist or be empty
			terminal.PrintWarning("Failed to migrate log directory (this is usually okay)")
		}
	}

	spinner.Stop()
	terminal.PrintSuccess("Data migration completed")
	return nil
}

// migrateFromConfigDefaults migrates using configuration defaults when detection fails
func (r *ConfigureRunner) migrateFromConfigDefaults() error {
	// Get default directories from config or fallback to standard paths
	appConfig, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to get configuration: %w", err)
	}

	sourceDataDir := "/var/lib/mysql"
	sourceBinlogDir := "/var/lib/mysqlbinlogs"
	sourceLogDir := "/var/log/mysql"

	if appConfig.MariaDB.Installation.DataDir != "" {
		sourceDataDir = appConfig.MariaDB.Installation.DataDir
	}
	if appConfig.MariaDB.Installation.BinlogDir != "" {
		sourceBinlogDir = appConfig.MariaDB.Installation.BinlogDir
	}
	if appConfig.MariaDB.Installation.LogDir != "" {
		sourceLogDir = appConfig.MariaDB.Installation.LogDir
	}

	// Migrate directories (same logic as above)
	if sourceDataDir != r.settings.DataDir {
		terminal.PrintInfo(fmt.Sprintf("Migrating data: %s → %s", sourceDataDir, r.settings.DataDir))
		dataMigrator := NewDataMigratorWithCleanup(sourceDataDir, r.settings.DataDir)
		if err := dataMigrator.MigrateData(); err != nil {
			return fmt.Errorf("failed to migrate data directory: %w", err)
		}
	}

	if sourceBinlogDir != r.settings.BinlogDir {
		terminal.PrintInfo(fmt.Sprintf("Migrating binlog: %s → %s", sourceBinlogDir, r.settings.BinlogDir))
		binlogMigrator := NewDataMigratorWithCleanup(sourceBinlogDir, r.settings.BinlogDir)
		if err := binlogMigrator.MigrateBinlogData(); err != nil {
			terminal.PrintWarning("Failed to migrate binlog directory (this is usually okay)")
		}
	}

	if sourceLogDir != r.settings.LogDir {
		terminal.PrintInfo(fmt.Sprintf("Migrating logs: %s → %s", sourceLogDir, r.settings.LogDir))
		logMigrator := NewDataMigratorWithCleanup(sourceLogDir, r.settings.LogDir)
		if err := logMigrator.MigrateData(); err != nil {
			terminal.PrintWarning("Failed to migrate log directory (this is usually okay)")
		}
	}

	return nil
}

// configureSELinux configures SELinux contexts
func (r *ConfigureRunner) configureSELinux() error {
	spinner := terminal.NewProgressSpinner("Configuring SELinux contexts...")
	spinner.Start()

	selinuxManager := NewSELinuxManager(r.settings.DataDir)
	if err := selinuxManager.ConfigureSELinux(); err != nil {
		spinner.Stop()
		return err
	}

	spinner.Stop()
	terminal.PrintSuccess("SELinux contexts configured")
	return nil
}

// startMariaDBService starts and enables MariaDB service
func (r *ConfigureRunner) startMariaDBService() error {
	spinner := terminal.NewProgressSpinner("Starting and enabling MariaDB service...")
	spinner.Start()

	serviceManager := NewServiceManagerWithSettings(r.settings)

	// Start service
	if err := serviceManager.StartService(); err != nil {
		spinner.Stop()
		return err
	}

	// Enable service on boot
	if err := serviceManager.EnableService(); err != nil {
		spinner.Stop()
		return err
	}

	spinner.Stop()
	terminal.PrintSuccess("MariaDB service started and enabled")

	// Show status
	return serviceManager.GetServiceStatus()
}

// initializeDatabaseIfNeeded checks if system tables exist and initializes database if needed
func (r *ConfigureRunner) initializeDatabaseIfNeeded() error {
	lg, _ := logger.Get()

	// Check if mysql system database exists
	mysqlDbPath := filepath.Join(r.settings.DataDir, "mysql")
	if _, err := os.Stat(mysqlDbPath); err == nil {
		// MySQL system database exists, check if it has required tables
		if r.hasRequiredSystemTables() {
			lg.Info("MariaDB system tables already exist, skipping initialization")
			terminal.PrintInfo("MariaDB system tables already exist")
			return nil
		}
	}

	spinner := terminal.NewProgressSpinner("Initializing MariaDB system database...")
	spinner.Start()

	lg.Info("Initializing MariaDB system database",
		logger.String("data_dir", r.settings.DataDir))

	// Run mysql_install_db to initialize system tables
	cmd := exec.Command("mysql_install_db",
		"--user=mysql",
		"--basedir=/usr",
		"--datadir="+r.settings.DataDir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		spinner.Stop()
		lg.Error("Failed to initialize MariaDB database",
			logger.Error(err),
			logger.String("output", string(output)))
		return fmt.Errorf("failed to initialize MariaDB database: %w", err)
	}

	spinner.Stop()
	lg.Info("MariaDB database initialized successfully",
		logger.String("data_dir", r.settings.DataDir))
	terminal.PrintSuccess("MariaDB database initialized successfully")

	return nil
}

// hasRequiredSystemTables checks if required system tables exist
func (r *ConfigureRunner) hasRequiredSystemTables() bool {
	requiredTables := []string{"mysql/db.frm", "mysql/user.frm", "mysql/plugin.frm"}

	for _, table := range requiredTables {
		tablePath := filepath.Join(r.settings.DataDir, table)
		if _, err := os.Stat(tablePath); err != nil {
			return false
		}
	}

	return true
}

// setupDatabasesAndUsers creates default databases and users
func (r *ConfigureRunner) setupDatabasesAndUsers(appConfig *model.Config) error {
	spinner := terminal.NewProgressSpinner("Setting up default databases and users...")
	spinner.Start()

	clientCode := appConfig.General.ClientCode
	if clientCode == "" {
		clientCode = "default"
	}

	dbManager := NewDatabaseManager(r.settings, clientCode)
	if err := dbManager.SetupDatabasesAndUsers(); err != nil {
		spinner.Stop()
		return err
	}

	spinner.Stop()
	terminal.PrintSuccess("Databases and users created successfully")
	return nil
}
