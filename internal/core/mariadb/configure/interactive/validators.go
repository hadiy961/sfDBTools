package interactive

import (
	"fmt"

	"sfDBTools/utils/fs/dir"
	"sfDBTools/utils/validation"
)

// Common validators untuk Task 2: menghilangkan duplikasi
func ValidateAbsolutePath(path string) error {
	// Use the repository's fs directory validation which performs
	// platform-aware checks and optionally permission/access checks.
	if err := dir.Validate(path); err != nil {
		return fmt.Errorf("must be absolute path or valid directory: %w", err)
	}
	return nil
}

func ValidateServerIDRange(serverID int) error {
	return validation.ValidateServerIDRange(serverID)
}

func ValidateBufferPoolInstances(instances int) error {
	return validation.ValidateBufferPoolInstances(instances)
}

func ValidateMemorySize(size string) error {
	return validation.ValidateMemorySize(size)
}
