package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FileSystem interface provides abstraction for safe file system operations
type FileSystem interface {
	SafeRemove(path string, validators ...Validator) error
	CalculateSize(path string) (int64, error)
	Exists(path string) bool
	IsDirectory(path string) bool
	CreateBackup(path, backupPath string) error
}

// Validator interface for path validation
type Validator interface {
	Validate(path string) error
}

// ValidatorFunc is a function adapter for Validator interface
type ValidatorFunc func(path string) error

func (f ValidatorFunc) Validate(path string) error {
	return f(path)
}

// fileSystem implements FileSystem interface
type fileSystem struct{}

// NewSafeFileSystem creates a new safe file system manager
func NewSafeFileSystem() FileSystem {
	return &fileSystem{}
}

// SafeRemove removes a path after validation
func (fs *fileSystem) SafeRemove(path string, validators ...Validator) error {
	// Basic safety checks
	if err := fs.validateBasicSafety(path); err != nil {
		return err
	}

	// Run custom validators
	for _, v := range validators {
		if err := v.Validate(path); err != nil {
			return fmt.Errorf("validation failed for %s: %w", path, err)
		}
	}

	// Perform removal
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove %s: %w", path, err)
	}

	return nil
}

// CalculateSize calculates the size of a file or directory in bytes
func (fs *fileSystem) CalculateSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// Exists checks if a path exists
func (fs *fileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsDirectory checks if a path is a directory
func (fs *fileSystem) IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// CreateBackup creates a backup of the given path
func (fs *fileSystem) CreateBackup(path, backupPath string) error {
	if !fs.Exists(path) {
		return fmt.Errorf("source path does not exist: %s", path)
	}

	// Create backup directory if it doesn't exist
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory %s: %w", backupDir, err)
	}

	// Copy the path to backup location
	cmd := exec.Command("cp", "-r", path, backupPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create backup of %s to %s: %w\nOutput: %s", path, backupPath, err, string(output))
	}

	return nil
}

// validateBasicSafety performs basic safety checks on paths
func (fs *fileSystem) validateBasicSafety(path string) error {
	// Normalize path
	cleanPath := filepath.Clean(path)

	// Prevent removal of critical system directories
	criticalPaths := []string{
		"/", "/bin", "/sbin", "/usr", "/usr/bin", "/usr/sbin",
		"/etc", "/boot", "/dev", "/proc", "/sys", "/tmp",
		"/home", "/root", "/var", "/opt",
	}

	for _, critical := range criticalPaths {
		if cleanPath == critical {
			return fmt.Errorf("refusing to remove critical system directory: %s", cleanPath)
		}
	}

	// Prevent removal of paths outside reasonable database directories
	allowedPrefixes := []string{
		"/var/lib/mysql",
		"/var/lib/mariadb",
		"/etc/mysql",
		"/etc/mariadb",
		"/var/log/mysql",
		"/var/log/mariadb",
		"/usr/share/mysql",
		"/usr/share/mariadb",
	}

	isAllowed := false
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(cleanPath, prefix) {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return fmt.Errorf("path %s is outside allowed database directories", cleanPath)
	}

	return nil
}

// MariaDBDataValidator validates that a path contains MariaDB data markers
var MariaDBDataValidator = ValidatorFunc(func(path string) error {
	markers := []string{"mysql/", "performance_schema/"}

	for _, marker := range markers {
		markerPath := filepath.Join(path, strings.TrimSuffix(marker, "/"))
		if _, err := os.Stat(markerPath); err == nil {
			return nil // Found at least one marker
		}
	}

	return fmt.Errorf("path %s does not appear to contain MariaDB data directories", path)
})

// MariaDBConfigValidator validates that a path contains MariaDB configuration
var MariaDBConfigValidator = ValidatorFunc(func(path string) error {
	configFiles := []string{"my.cnf", "mariadb.cnf", "mysql.cnf"}

	for _, configFile := range configFiles {
		configPath := filepath.Join(path, configFile)
		if _, err := os.Stat(configPath); err == nil {
			return nil // Found at least one config file
		}
	}

	return fmt.Errorf("path %s does not appear to contain MariaDB configuration files", path)
})
