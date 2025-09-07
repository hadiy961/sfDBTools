package remove

import (
	"fmt"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/repository"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// Config for remover
type Config struct {
	SkipConfirm bool
}

// RemoveResult contains outcome
type RemoveResult struct {
	Success   bool
	Message   string
	RemovedAt time.Time
}

// Remover performs MariaDB removal using modular components
type Remover struct {
	cfg            *Config
	serviceManager *ServiceManager
	packageManager *PackageManager
	fileManager    *FileManager
	configParser   *ConfigParser
	validator      *Validator
	repoMgr        *repository.Manager
}

// DefaultConfig returns default configuration for removal
func DefaultConfig() *Config {
	return &Config{
		SkipConfirm: false,
	}
}

// NewRemover constructs a Remover with all necessary components
func NewRemover(cfg *Config) (*Remover, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Use existing helpers to detect OS
	osInfo, err := system.DetectOS()
	if err != nil {
		return nil, fmt.Errorf("failed to detect OS: %w", err)
	}

	return &Remover{
		cfg:            cfg,
		serviceManager: NewServiceManager(),
		packageManager: NewPackageManager(),
		fileManager:    NewFileManager(),
		configParser:   NewConfigParser(),
		validator:      NewValidator(),
		repoMgr:        repository.NewManager(osInfo),
	}, nil
}

// Remove performs destructive removal following safe prompts
func (r *Remover) Remove() (*RemoveResult, error) {
	lg, _ := logger.Get()

	terminal.ClearAndShowHeader("MariaDB Remove")

	// Step 1: Validate that MariaDB services exist
	servicesFound, err := r.validator.ValidateMariaDBServices()
	if err != nil {
		return r.validator.CreateResult(false, "validation failed"), err
	}
	if !servicesFound {
		return r.validator.CreateResult(false, "no MariaDB services found"), nil
	}

	// Step 2: Get user confirmation
	confirmed, err := r.validator.ConfirmRemoval(r.cfg.SkipConfirm)
	if err != nil {
		return r.validator.CreateResult(false, "confirmation failed"), err
	}
	if !confirmed {
		return r.validator.CreateResult(false, "cancelled by user"), nil
	}

	// Step 3: Stop and disable MariaDB services
	terminal.PrintInfo("Checking MariaDB services...")
	r.serviceManager.StopAndDisableServices()

	// Step 4: Remove packages
	if err := r.packageManager.RemoveMariaDBPackages(); err != nil {
		lg.Warn("Package removal failed but continuing with cleanup", logger.Error(err))
	}

	// Step 5: Remove standard directories
	r.fileManager.RemoveDefaultDirectories()

	// Step 6: Handle custom configuration files and directories
	customConfigs := r.configParser.FindCustomConfigFiles()
	r.fileManager.RemoveCustomDirectories(customConfigs, r.cfg.SkipConfirm)

	// Step 7: Remove system user and group
	r.fileManager.RemoveUserAndGroup()

	// Step 8: Clean repository entries
	if err := r.repoMgr.Clean(); err != nil {
		lg.Warn("Repository cleanup failed", logger.Error(err))
	}

	terminal.PrintSuccess("MariaDB cleanup finished")
	lg.Info("MariaDB removal completed")

	return r.validator.CreateResult(true, "completed"), nil
}
