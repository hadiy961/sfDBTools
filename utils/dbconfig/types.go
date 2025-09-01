package dbconfig

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
	EncryptionPassword string
}

// DeleteResult represents the result of a delete operation
type DeleteResult struct {
	DeletedCount int
	ErrorCount   int
	DeletedFiles []string
	Errors       []string
}

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	FileValid        bool
	DecryptionValid  bool
	ConnectionValid  bool
	ServerVersion    string
	ValidationErrors []string
}

// ConfigSelection represents a file selection option
type ConfigSelection struct {
	Index    int
	FilePath string
	Name     string
}
