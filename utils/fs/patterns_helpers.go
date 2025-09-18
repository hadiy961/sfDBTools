package fs

import (
	"path/filepath"
	"strings"
)

// normalize returns lowercased filename for case-insensitive comparisons
func normalize(s string) string {
	return strings.ToLower(s)
}

// baseName returns the base name of a path
func baseName(path string) string {
	return filepath.Base(path)
}

// hasSuffixCI checks suffix case-insensitively
func hasSuffixCI(name string, suffix string) bool {
	return strings.HasSuffix(normalize(name), normalize(suffix))
}

// hasPrefixCI checks prefix case-insensitively
func hasPrefixCI(name string, prefix string) bool {
	return strings.HasPrefix(normalize(name), normalize(prefix))
}

// equalsCI checks equality case-insensitively
func equalsCI(a, b string) bool {
	return normalize(a) == normalize(b)
}
