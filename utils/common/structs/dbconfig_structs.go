package structs

import "time"

// DBConfig - Database configuration related flags
type DBConfig struct {
	// File paths and selection
	ConfigName        string
	ConnectionOptions ConnectionOptions
	FileInfo          FileInfo

	// Operation flags
	ForceDelete bool
	DeleteAll   bool
	AutoMode    bool

	// Authentication
	EncryptionConfig EncryptionConfig
}

// InputConfig represents configuration input data
type DBConfigInput struct {
	Name     string
	Host     string
	Port     int
	User     string
	Password string
}

// EncryptionConfig database configuration and backup encryption
type EncryptionConfig struct {
	PasswordType       string
	EncryptionPassword string
}

// FileInfo represents configuration file information
type FileInfo struct {
	Name    string
	Path    string
	Size    int64
	ModTime time.Time
	IsValid bool
}

// ConfigInfo represents database configuration details
type DBConfigInfo struct {
	ConnectionOptions ConnectionOptions
	HasPassword       bool
	FileSize          string
	LastModified      time.Time
	IsValid           bool
}

// DBConfigResult represents the result of a database configuration operation
type DBConfigResult struct {
	DBConfigDeleteResult DBConfigDeleteResult
	ValidationResult     DBConfigValidationResult
	GenerationResult     GenerationResult
	EditResult           EditResult
}

type GenerationResult struct {
	ConfigName  string
	FilePath    string
	Overwritten bool
}

type EditResult struct {
	ConfigName string
	FilePath   string
	BackupFile string
}

// DBConfigDeleteResult represents the result of a delete operation
type DBConfigDeleteResult struct {
	DeletedCount int
	ErrorCount   int
	DeletedFiles []string
	Errors       []string
}

// ValidationResult represents config validation outcome
type DBConfigValidationResult struct {
	IsValid     bool
	Errors      []string
	Warnings    []string
	ConfigName  string
	TestResults map[string]bool
}

// ConfigSelection represents a file selection option
type DBConfigSelection struct {
	Index    int
	FilePath string
	Name     string
}
