package validation

import (
	"fmt"
)

// Centralized constants
const (
	MinServerID = 1
	MaxServerID = 4294967295

	MinBufferPoolInstances = 1
	MaxBufferPoolInstances = 64
)

// ValidateServerIDRange validates server id range
func ValidateServerIDRange(serverID int) error {
	if serverID < MinServerID || serverID > MaxServerID {
		return fmt.Errorf("server ID must be between %d and %d", MinServerID, MaxServerID)
	}
	return nil
}

// ValidateBufferPoolInstances validates buffer pool instances range
func ValidateBufferPoolInstances(instances int) error {
	if instances < MinBufferPoolInstances || instances > MaxBufferPoolInstances {
		return fmt.Errorf("buffer pool instances must be between %d and %d", MinBufferPoolInstances, MaxBufferPoolInstances)
	}
	return nil
}

// IsValidMemorySize checks if memory size format like 1G, 512M, 1024K
func IsValidMemorySize(size string) bool {
	if len(size) < 2 {
		return false
	}

	suffix := size[len(size)-1:]
	if suffix != "M" && suffix != "G" && suffix != "K" {
		return false
	}

	numPart := size[:len(size)-1]
	for _, c := range numPart {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// ValidateMemorySize returns error if memory size is invalid
func ValidateMemorySize(size string) error {
	if !IsValidMemorySize(size) {
		return fmt.Errorf("invalid memory size format: %s (use format like 1G, 512M)", size)
	}
	return nil
}
