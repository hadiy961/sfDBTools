package check_version

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

// Pre-compiled regex patterns for parsing fallback
var (
	datePattern    = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
	eolPattern     = regexp.MustCompile(`"eol"\s*:\s*"([0-9]{4}-[0-9]{2}-[0-9]{2})"`)
	eolBoolPattern = regexp.MustCompile(`"eol"\s*:\s*true`)
)

// extractDateFromMap extracts date from a map using common date field names
func extractDateFromMap(m map[string]interface{}) string {
	dateKeys := []string{"release", "release-date", "release_date", "released", "first_release"}
	for _, k := range dateKeys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				if datePattern.MatchString(s) {
					return s
				}
				if t, err := time.Parse(time.RFC3339, s); err == nil {
					return t.Format("2006-01-02")
				}
			}
		}
	}

	if lr, ok := m["latest_release"]; ok {
		if lm, ok := lr.(map[string]interface{}); ok {
			nestedDateKeys := []string{"release", "release_date", "released"}
			for _, k := range nestedDateKeys {
				if v, ok := lm[k]; ok {
					if s, ok := v.(string); ok {
						if datePattern.MatchString(s) {
							return s
						}
					}
				}
			}
		}
	}
	return ""
}

// processJSONResponse processes JSON response and extracts date information
func processJSONResponse(body []byte, _ bool) string {
	var raw interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return ""
	}

	switch t := raw.(type) {
	case map[string]interface{}:
		return extractDateFromMap(t)
	case []interface{}:
		if len(t) > 0 {
			if first, ok := t[0].(map[string]interface{}); ok {
				return extractDateFromMap(first)
			}
		}
	}
	return ""
}

// handleEOLField interprets an object map and extracts an "eol" value if present
func handleEOLField(m map[string]interface{}) string {
	if rawEOL, ok := m["eol"]; ok {
		switch v := rawEOL.(type) {
		case string:
			if datePattern.MatchString(v) {
				return v
			}
		case bool:
			if v {
				return time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			}
		}
	}
	return ""
}

// Version helper utilities
var versionPattern = regexp.MustCompile(`^\d+\.\d+(?:\.\d+)?(?:-rc\d*|\.rolling)?$`)

// IsValidVersion checks if a version string is valid (basic)
func IsValidVersion(version string) bool {
	return versionPattern.MatchString(strings.TrimSpace(version))
}

// DetermineVersionType returns "rolling", "rc" or "stable"
func DetermineVersionType(version string) string {
	v := strings.ToLower(version)
	if strings.Contains(v, "rolling") {
		return "rolling"
	}
	if strings.Contains(v, "rc") {
		return "rc"
	}
	return "stable"
}
