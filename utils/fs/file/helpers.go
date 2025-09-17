package file

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureParentDir makes sure the parent directory of filePath exists.
// It creates parents with mode 0755 if necessary.
func EnsureParentDir(filePath string) error {
	parent := filepath.Dir(filePath)
	if parent == "" || parent == "." {
		return nil
	}
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("failed to create parent directory '%s': %w", parent, err)
	}
	return nil
}

// TestWrite attempts to create a small test file at filePath to verify writability.
// It uses the provided mode when creating the file and removes it afterwards.
func TestWrite(filePath string, mode os.FileMode) error {
	if err := EnsureParentDir(filePath); err != nil {
		return err
	}

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("failed to create test file '%s': %w", filePath, err)
	}
	f.Close()
	if err := os.Remove(filePath); err != nil {
		// Not fatal, but warn via error so caller can decide
		return fmt.Errorf("failed to cleanup test file '%s': %w", filePath, err)
	}
	return nil
}
