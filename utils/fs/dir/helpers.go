package dir

import (
	"fmt"
	"os"

	"sfDBTools/internal/logger"
)

// Ensure makes sure the directory exists and is writable. It creates the directory if missing.
func Ensure(path string) error {
	m := NewManager()
	if err := m.Validate(path); err != nil {
		// Attempt to create
		if err := m.Create(path); err != nil {
			return fmt.Errorf("failed to ensure directory '%s': %w", path, err)
		}
	}
	return nil
}

// EnsureWithPermissions ensures the directory exists and sets the requested permissions and ownership.
// It will attempt to create the directory, then set permissions via the Manager.
func EnsureWithPermissions(path string, mode os.FileMode, owner, group string) error {
	lg, _ := logger.Get()
	m := NewManager()

	if err := Ensure(path); err != nil {
		return err
	}

	// Try to set permissions/ownership; non-fatal if setting ownership fails when running non-root,
	// but we return the error so callers can decide.
	if err := m.SetPermissions(path, mode, owner, group); err != nil {
		lg.Warn("EnsureWithPermissions: failed to set permissions", logger.String("path", path), logger.Error(err))
		return fmt.Errorf("failed to set permissions for '%s': %w", path, err)
	}

	return nil
}
