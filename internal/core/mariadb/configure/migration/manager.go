package migration

import (
	"sfDBTools/internal/logger"
	fsutil "sfDBTools/utils/fs"
)

// MigrationManager manages all migration operations with consistent filesystem access
type MigrationManager struct {
	fsMgr  *fsutil.Manager
	logger *logger.Logger
}

// MigrationOperations defines the interface for migration operations
type MigrationOperations interface {
	// Directory operations
	CopyDirectory(source, destination string) error
	CopyLogFilesOnly(source, destination string) error

	// File operations
	CopyFile(srcPath, destPath string, info interface{}) error

	// Utility operations
	IsLogFile(path string) bool
	IsDataDirectory(path, sourceRoot string) bool
}

// NewMigrationManager creates a new migration manager instance
func NewMigrationManager() *MigrationManager {
	lg, _ := logger.Get()
	return &MigrationManager{
		fsMgr:  fsutil.NewManager(),
		logger: lg,
	}
}

// FileSystem returns the filesystem manager
func (m *MigrationManager) FileSystem() *fsutil.Manager {
	return m.fsMgr
}

// Logger returns the logger instance
func (m *MigrationManager) Logger() *logger.Logger {
	return m.logger
}
