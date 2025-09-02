package check_version

import (
	"fmt"
	"strings"
)

// CompareVersions compares two version strings, returns true if v1 < v2
func CompareVersions(v1, v2 string) bool {
	// Handle special cases
	if strings.Contains(v1, "rolling") {
		return false // rolling is always "latest"
	}
	if strings.Contains(v2, "rolling") {
		return true
	}
	if strings.Contains(v1, "rc") && !strings.Contains(v2, "rc") {
		return true // rc versions come before stable
	}
	if !strings.Contains(v1, "rc") && strings.Contains(v2, "rc") {
		return false
	}

	// Parse version numbers
	parts1 := strings.Split(strings.Replace(v1, "rc", "", -1), ".")
	parts2 := strings.Split(strings.Replace(v2, "rc", "", -1), ".")

	// Compare each part
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		if i < len(parts1) && parts1[i] != "" {
			fmt.Sscanf(parts1[i], "%d", &p1)
		}
		if i < len(parts2) && parts2[i] != "" {
			fmt.Sscanf(parts2[i], "%d", &p2)
		}

		if p1 < p2 {
			return true
		}
		if p1 > p2 {
			return false
		}
	}

	return false // versions are equal
}

// parseVersionNumber safely parses version number string to int
func parseVersionNumber(versionStr string) (int, error) {
	var num int
	_, err := fmt.Sscanf(versionStr, "%d", &num)
	return num, err
}
