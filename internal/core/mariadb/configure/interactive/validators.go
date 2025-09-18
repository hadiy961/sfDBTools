package interactive

import (
	"fmt"
	"path/filepath"
)

// Common validators untuk Task 2: menghilangkan duplikasi
func ValidateAbsolutePath(path string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("must be absolute path: %s", path)
	}
	return nil
}

func ValidatePortRange(port int) error {
	if port < 1024 || port > 65535 {
		return fmt.Errorf("port must be between 1024 and 65535")
	}
	return nil
}

func ValidateServerIDRange(serverID int) error {
	if serverID < 1 || serverID > 4294967295 {
		return fmt.Errorf("server ID must be between 1 and 4294967295")
	}
	return nil
}

func ValidateBufferPoolInstances(instances int) error {
	if instances < 1 || instances > 64 {
		return fmt.Errorf("buffer pool instances must be between 1 and 64")
	}
	return nil
}

func ValidateMemorySize(size string) error {
	if !isValidMemorySize(size) {
		return fmt.Errorf("invalid memory size format: %s (use format like 1G, 512M)", size)
	}
	return nil
}

// isValidMemorySize memeriksa apakah format ukuran memory valid
func isValidMemorySize(size string) bool {
	if len(size) < 2 {
		return false
	}

	// Check suffix
	suffix := size[len(size)-1:]
	if suffix != "M" && suffix != "G" && suffix != "K" {
		return false
	}

	// Check numeric part
	numPart := size[:len(size)-1]
	for _, char := range numPart {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}
