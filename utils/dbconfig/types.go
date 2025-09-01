package dbconfig

import (
	"fmt"
	"time"
)

// DeletionType represents the type of deletion operation
type DeletionType string

const (
	DeletionSingle   DeletionType = "single"
	DeletionMultiple DeletionType = "multiple"
	DeletionAll      DeletionType = "all"
)

// OperationType represents database config operations
type OperationType string

const (
	OperationShow     OperationType = "show"
	OperationValidate OperationType = "validate"
	OperationDelete   OperationType = "delete"
	OperationEdit     OperationType = "edit"
	OperationGenerate OperationType = "generate"
)

// FileInfo represents configuration file information
type FileInfo struct {
	Name    string
	Path    string
	Size    int64
	ModTime time.Time
	IsValid bool
}

// ConfigInfo represents database configuration details
type ConfigInfo struct {
	Name         string
	Host         string
	Port         int
	User         string
	HasPassword  bool
	FileSize     string
	LastModified time.Time
	IsValid      bool
}

// ValidationResult represents config validation outcome
type ValidationResult struct {
	IsValid     bool
	Errors      []string
	Warnings    []string
	ConfigName  string
	TestResults map[string]bool
}

// GenerationOptions represents options for config generation
type GenerationOptions struct {
	Name         string
	Host         string
	Port         int
	User         string
	Password     string
	Overwrite    bool
	SkipPassword bool
}

// EditOptions represents options for config editing
type EditOptions struct {
	ConfigName  string
	Interactive bool
	Backup      bool
}

// Config represents database configuration for dbconfig operations
type Config struct {
	// File paths and selection
	FilePath   string
	ConfigName string

	// Database connection details (for generate/edit operations)
	Host     string
	Port     int
	User     string
	Password string

	// Operation flags
	ForceDelete bool
	DeleteAll   bool
	AutoMode    bool

	// Authentication
	PasswordType       string
	EncryptionPassword string
}

// DeleteResult represents the result of a delete operation
type DeleteResult struct {
	DeletedCount int
	ErrorCount   int
	DeletedFiles []string
	Errors       []string
}

// ConfigSelection represents a file selection option
type ConfigSelection struct {
	Index    int
	FilePath string
	Name     string
}

// String returns a human-readable description of the operation
func (o OperationType) String() string {
	switch o {
	case OperationShow:
		return "show configuration"
	case OperationValidate:
		return "validate configuration"
	case OperationDelete:
		return "delete configuration"
	case OperationEdit:
		return "edit configuration"
	case OperationGenerate:
		return "generate configuration"
	default:
		return string(o)
	}
}

// String returns a human-readable description of the deletion type
func (d DeletionType) String() string {
	switch d {
	case DeletionSingle:
		return "single configuration"
	case DeletionMultiple:
		return "selected configurations"
	case DeletionAll:
		return "all configurations"
	default:
		return string(d)
	}
}

// GetDisplayName returns formatted display name for the file
func (f *FileInfo) GetDisplayName() string {
	if f.IsValid {
		return fmt.Sprintf("%s (✓)", f.Name)
	}
	return fmt.Sprintf("%s (✗)", f.Name)
}

// GetFormattedSize returns human-readable file size
func (f *FileInfo) GetFormattedSize() string {
	if f.Size < 1024 {
		return fmt.Sprintf("%d B", f.Size)
	} else if f.Size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(f.Size)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(f.Size)/(1024*1024))
}

// HasErrors returns true if validation has errors
func (v *ValidationResult) HasErrors() bool {
	return len(v.Errors) > 0
}

// HasWarnings returns true if validation has warnings
func (v *ValidationResult) HasWarnings() bool {
	return len(v.Warnings) > 0
}
