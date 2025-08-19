package migrate_utils

// MigrationConfig represents the resolved migration configuration
type MigrationConfig struct {
	// Source database configuration
	SourceHost     string
	SourcePort     int
	SourceUser     string
	SourcePassword string
	SourceDBName   string

	// Target database configuration
	TargetHost     string
	TargetPort     int
	TargetUser     string
	TargetPassword string
	TargetDBName   string // For single migration, this will be same as SourceDBName

	// Migration options
	MigrateUsers     bool
	MigrateData      bool
	MigrateStructure bool
	VerifyData       bool
	BackupTarget     bool
	DropTarget       bool
	CreateTarget     bool
}

// MigrationResult represents the result of a migration operation
type MigrationResult struct {
	SourceDBName    string
	TargetDBName    string
	TablesProcessed int
	RecordsMigrated int64
	StartTime       string
	EndTime         string
	Duration        string
	BackupFile      string
	Success         bool
	Error           error
}

// ConfigurationSource represents the source of migration configuration
type ConfigurationSource int

const (
	SourceConfigFile ConfigurationSource = iota
	SourceFlags
	SourceDefaults
	SourceInteractive
)
