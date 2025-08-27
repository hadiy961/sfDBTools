package upgrade

// UpgradeConfig holds configuration for MariaDB upgrade
type UpgradeConfig struct {
	// Target version to upgrade to
	TargetVersion string

	// AutoConfirm skips confirmation prompts
	AutoConfirm bool

	// BackupData indicates whether to backup data before upgrade
	BackupData bool

	// BackupPath specifies where to store backup if BackupData is true
	BackupPath string

	// SkipBackup skips data backup (dangerous)
	SkipBackup bool

	// ForceUpgrade allows upgrading even if target version is lower
	ForceUpgrade bool

	// SkipPostUpgrade skips mysql_upgrade execution
	SkipPostUpgrade bool

	// TestMode performs dry-run without actual upgrade
	TestMode bool

	// RemoveExisting removes existing installation before upgrade
	RemoveExisting bool

	// StartService starts MariaDB service after upgrade
	StartService bool

	// EnableSecurity enables security setup after upgrade
	EnableSecurity bool
}

// DefaultUpgradeConfig returns default upgrade configuration
func DefaultUpgradeConfig() *UpgradeConfig {
	return &UpgradeConfig{
		TargetVersion:   "",
		AutoConfirm:     false,
		BackupData:      true,
		BackupPath:      "", // Use default
		SkipBackup:      false,
		ForceUpgrade:    false,
		SkipPostUpgrade: false,
		TestMode:        false,
		RemoveExisting:  false,
		StartService:    true,
		EnableSecurity:  true,
	}
}

// CurrentInstallation represents current MariaDB installation
type CurrentInstallation struct {
	IsInstalled    bool
	Version        string
	PackageName    string
	DataDirectory  string
	ServiceName    string
	ServiceRunning bool
	ServiceEnabled bool
	ConfigFiles    []string
}

// UpgradeStep represents a step in the upgrade process
type UpgradeStep struct {
	Name        string
	Description string
	Required    bool
	Completed   bool
	Error       error
}

// UpgradePlan contains the execution plan for upgrade
type UpgradePlan struct {
	CurrentVersion string
	TargetVersion  string
	UpgradeType    UpgradeType
	Steps          []UpgradeStep
	BackupPath     string
	EstimatedTime  string
	Risks          []string
	Prerequisites  []string
}

// UpgradeType defines the type of upgrade
type UpgradeType string

const (
	UpgradeTypeMajor   UpgradeType = "major"   // e.g., 10.6 -> 11.4
	UpgradeTypeMinor   UpgradeType = "minor"   // e.g., 10.6.19 -> 10.6.23
	UpgradeTypePatch   UpgradeType = "patch"   // e.g., 10.6.23-1 -> 10.6.23-2
	UpgradeTypeNone    UpgradeType = "none"    // Already latest
	UpgradeTypeInvalid UpgradeType = "invalid" // Downgrade or invalid
)

// UpgradeResult contains the result of upgrade operation
type UpgradeResult struct {
	Success         bool
	PreviousVersion string
	NewVersion      string
	BackupPath      string
	Duration        string
	StepsCompleted  int
	StepsTotal      int
	Error           error
	RollbackInfo    *RollbackInfo
}

// RollbackInfo contains information needed for rollback
type RollbackInfo struct {
	AvailableBackup string
	PreviousVersion string
	RollbackSteps   []string
}

// ValidationResult contains validation results
type ValidationResult struct {
	Valid       bool
	Errors      []string
	Warnings    []string
	Suggestions []string
}
