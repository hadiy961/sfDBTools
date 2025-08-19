package connection

import (
	"strings"
)

// Helper function to check if a string contains a substring (case insensitive)
func contains(s, substr string) bool {
	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
	return strings.Contains(s, substr)
}
