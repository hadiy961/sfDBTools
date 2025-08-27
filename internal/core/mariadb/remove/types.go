package remove

import "sfDBTools/utils/common"

// RemovalConfig holds configuration for MariaDB removal
type RemovalConfig struct {
	// RemoveData indicates whether to remove data directories
	RemoveData bool

	// BackupData indicates whether to backup data before removal
	BackupData bool

	// BackupPath specifies where to store backup if BackupData is true
	BackupPath string

	// RemoveRepositories indicates whether to remove MariaDB repositories
	RemoveRepositories bool

	// AutoConfirm skips confirmation prompts
	AutoConfirm bool

	// DataDirectory specifies the MariaDB data directory to remove
	DataDirectory string

	// ConfigDirectory specifies the MariaDB config directory to remove
	ConfigDirectory string

	// LogDirectory specifies the MariaDB log directory to remove
	LogDirectory string

	// ForceRemoval bypasses safety checks
	ForceRemoval bool
}

// DefaultRemovalConfig returns a default removal configuration
func DefaultRemovalConfig() *RemovalConfig {
	return &RemovalConfig{
		RemoveData:         false, // Default to keeping data for safety
		BackupData:         true,  // Default to backing up data
		RemoveRepositories: false, // Keep repositories by default
		AutoConfirm:        false, // Require manual confirmation by default
		DataDirectory:      "/var/lib/mysql",
		ConfigDirectory:    "/etc/mysql",
		LogDirectory:       "/var/log/mysql",
		ForceRemoval:       false,
	}
}

// DetectedInstallation represents a detected MariaDB installation
type DetectedInstallation struct {
	// IsInstalled indicates if MariaDB is installed
	IsInstalled bool

	// Version is the installed version
	Version string

	// PackageName is the name of the installed package
	PackageName string

	// ServiceName is the name of the system service
	ServiceName string

	// ServiceActive indicates if the service is currently running
	ServiceActive bool

	// ServiceEnabled indicates if the service is enabled on boot
	ServiceEnabled bool

	// DataDirectoryExists indicates if data directory exists
	DataDirectoryExists bool

	// DataDirectorySize is the size of data directory in bytes
	DataDirectorySize int64

	// ConfigFiles lists found configuration files
	ConfigFiles []string

	// LogFiles lists found log files
	LogFiles []string

	// OSInfo contains OS information for removal strategy
	OSInfo *common.OSInfo
}

// RemovalSummary contains information about what was removed
type RemovalSummary struct {
	// PackagesRemoved lists packages that were removed
	PackagesRemoved []string

	// ServicesRemoved lists services that were removed
	ServicesRemoved []string

	// DataBackedUp indicates if data was backed up
	DataBackedUp bool

	// BackupLocation is the path where data was backed up
	BackupLocation string

	// DataRemoved indicates if data directories were removed
	DataRemoved bool

	// ConfigRemoved indicates if config files were removed
	ConfigRemoved bool

	// RepositoriesRemoved indicates if repositories were removed
	RepositoriesRemoved bool

	// Errors contains any errors that occurred during removal
	Errors []error
}
