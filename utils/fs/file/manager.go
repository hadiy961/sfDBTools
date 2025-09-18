package file

import (
	"sfDBTools/internal/logger"

	"github.com/spf13/afero"
)

// Manager handles file-level operations and wraps an afero filesystem for testability
type Manager struct {
	fs     afero.Fs
	logger *logger.Logger
}

// NewManager returns a manager using the real OS filesystem
func NewManager() *Manager {
	lg, _ := logger.Get()
	return &Manager{fs: afero.NewOsFs(), logger: lg}
}

// NewManagerWithFs returns a manager using a provided afero filesystem
func NewManagerWithFs(fs afero.Fs) *Manager {
	lg, _ := logger.Get()
	return &Manager{fs: fs, logger: lg}
}
