package check_version

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// parsing helpers and regex moved to parser.go

// GetMariaDBEOLDate returns EOL date for MariaDB version using dynamic approach
func GetMariaDBEOLDate(version string) string {
	// For rolling/rc versions, return "No LTS"
	if strings.Contains(version, "rolling") || strings.Contains(version, "rc") {
		return NoLTS
	}

	// Try external source first (fast timeout)
	if eolDate := tryFetchEOLFromExternal(version); eolDate != "" {
		return eolDate
	}

	// Fall back to calculation based on lifecycle
	return calculateEOLFromLifecycle(version)
}

// tryFetchEOLFromExternal attempts to fetch EOL from external sources
func tryFetchEOLFromExternal(version string) string {
	// Normalize version and extract major.minor
	major, minor, ok := normalizeVersion(version)
	if !ok {
		return ""
	}
	majorMinor := major + "." + minor

	// check cache
	eolCacheMu.RLock()
	if ent, found := eolCache[majorMinor]; found {
		if time.Now().Before(ent.expires) {
			eolCacheMu.RUnlock()
			return ent.value
		}
	}
	eolCacheMu.RUnlock()

	// Create Resty client with timeout
	client := getHTTPClient(DefaultEOLTimeout)

	// Try endoflife.date API
	url := fmt.Sprintf(DefaultEndOfLifeAPI, majorMinor)
	resp, err := client.R().Get(url)
	if err != nil {
		return "" // Fail silently for external sources
	}

	if resp.StatusCode() != 200 {
		return ""
	}

	body := resp.Body()

	// Decode JSON robustly instead of regex
	var raw interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		// fallback to previous regex approach if JSON decode fails
		if matches := eolPattern.FindStringSubmatch(string(body)); len(matches) > 1 {
			return matches[1]
		}
		if eolBoolPattern.MatchString(string(body)) {
			return time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		}
		return ""
	}

	// helper to interpret "eol" field when present
	handleMap := func(m map[string]interface{}) string {
		if rawEOL, ok := m["eol"]; ok {
			switch v := rawEOL.(type) {
			case string:
				// basic validation: YYYY-MM-DD
				if datePattern.MatchString(v) {
					return v
				}
			case bool:
				if v {
					// already EOL -> return yesterday
					return time.Now().AddDate(0, 0, -1).Format("2006-01-02")
				}
			}
		}
		return ""
	}

	switch t := raw.(type) {
	case map[string]interface{}:
		if res := handleMap(t); res != "" {
			// cache and return
			eolCacheMu.Lock()
			eolCache[majorMinor] = cacheEntry{value: res, expires: time.Now().Add(eolCacheTTL)}
			eolCacheMu.Unlock()
			return res
		}
	case []interface{}:
		if len(t) > 0 {
			if first, ok := t[0].(map[string]interface{}); ok {
				if res := handleMap(first); res != "" {
					// cache and return
					eolCacheMu.Lock()
					eolCache[majorMinor] = cacheEntry{value: res, expires: time.Now().Add(eolCacheTTL)}
					eolCacheMu.Unlock()
					return res
				}
			}
		}
	}

	return ""
}

// calculateEOLFromLifecycle calculates EOL based on MariaDB lifecycle policy
func calculateEOLFromLifecycle(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return TBD
	}

	major := parts[0]
	minor := parts[1]

	// Determine if LTS based on known pattern
	isLTS := isLTSVersion(major, minor)

	if isLTS {
		return estimateLTSEOL(major, minor)
	}

	return estimateStableEOL(major, minor)
}

// isLTSVersion determines if a version is LTS based on MariaDB's pattern
func isLTSVersion(major, minor string) bool {
	// Known LTS pattern: typically every 1.5-2 years
	ltsVersions := map[string][]string{
		"10": {"5", "6", "11"}, // Known LTS
		"11": {"4"},            // Known LTS
		// Future pattern: likely 12.4, 13.4, etc.
	}

	if minors, exists := ltsVersions[major]; exists {
		for _, ltsMinor := range minors {
			if minor == ltsMinor {
				return true
			}
		}
	}

	// Pattern-based detection for future versions
	if majorNum, err := parseVersionNumber(major); err == nil && majorNum >= 12 {
		if minorNum, err := parseVersionNumber(minor); err == nil && minorNum == 4 {
			return true // Assume X.4 pattern continues
		}
	}

	return false
}

// estimateLTSEOL estimates EOL for LTS versions
func estimateLTSEOL(major, minor string) string {
	// Try to get release date and add 5 years
	if releaseDate := estimateReleaseDate(major, minor); releaseDate != "" {
		if releaseTime, err := time.Parse("2006-01-02", releaseDate); err == nil {
			eolTime := releaseTime.AddDate(5, 0, 0) // LTS = 5 years support
			return eolTime.Format("2006-01-02")
		}
	}

	// Fallback: conservative estimate
	return TBD
}

// estimateStableEOL estimates EOL for stable versions
func estimateStableEOL(major, minor string) string {
	// Stable versions typically get 18 months support
	if releaseDate := estimateReleaseDate(major, minor); releaseDate != "" {
		if releaseTime, err := time.Parse("2006-01-02", releaseDate); err == nil {
			eolTime := releaseTime.AddDate(1, 6, 0) // 18 months support
			return eolTime.Format("2006-01-02")
		}
	}

	return TBD
}

// estimateReleaseDate estimates release date based on version pattern
func estimateReleaseDate(major, minor string) string {
	// 0) Try external services as primary source
	if d := tryFetchReleaseDateFromExternal(major, minor); d != "" {
		return d
	}

	// 1) Known release dates for reference (local fallback)
	knownReleases := map[string]string{
		"10.5":  "2020-06-24",
		"10.6":  "2021-07-06",
		"10.11": "2023-02-16",
		"11.4":  "2024-05-29",
	}

	versionKey := major + "." + minor
	if date, exists := knownReleases[versionKey]; exists {
		return date
	}

	// 2) Pattern-based estimation for unknown versions
	majorNum, err1 := parseVersionNumber(major)
	minorNum, err2 := parseVersionNumber(minor)

	if err1 != nil || err2 != nil {
		return ""
	}

	// Estimate based on release pattern
	if majorNum >= 11 {
		// MariaDB 11+ typically releases annually with minor releases quarterly
		baseYear := 2024 + (majorNum - 11)
		estimatedMonth := 2 + (minorNum * 3) // Quarterly releases starting Feb
		if estimatedMonth > 12 {
			baseYear++
			estimatedMonth = estimatedMonth - 12
		}

		estimated := time.Date(baseYear, time.Month(estimatedMonth), 15, 0, 0, 0, 0, time.UTC)
		if estimated.Before(time.Now().AddDate(5, 0, 0)) { // Reasonable future limit
			return estimated.Format("2006-01-02")
		}
	}

	return ""
}

// tryFetchReleaseDateFromExternal tries external sources (endoflife.date, GitHub Releases)
// returns YYYY-MM-DD or empty string
func tryFetchReleaseDateFromExternal(major, minor string) string {
	majorMinor := major + "." + minor

	// check cache for release date as well
	eolCacheMu.RLock()
	if ent, found := eolCache[majorMinor]; found {
		if time.Now().Before(ent.expires) {
			eolCacheMu.RUnlock()
			return ent.value
		}
	}
	eolCacheMu.RUnlock()

	// Create Resty client
	client := getHTTPClient(DefaultEOLTimeout)

	// 1) Try endoflife.date first (fast)
	url := fmt.Sprintf(DefaultEndOfLifeAPI, majorMinor)
	resp, err := client.R().Get(url)
	if err == nil && resp.StatusCode() == 200 {
		if date := processJSONResponse(resp.Body(), false); date != "" {
			return date
		}
	}

	// 2) Fall back to GitHub Releases - search for a release that mentions major.minor
	client.SetTimeout(DefaultHTTPTimeout)
	resp, err = client.R().Get(DefaultGitHubReleasesAPI)
	if err == nil && resp.StatusCode() == 200 {
		var releases []map[string]interface{}
		if err := json.Unmarshal(resp.Body(), &releases); err == nil {
			for _, rel := range releases {
				// check tag_name, name, body for majorMinor
				found := false
				for _, k := range []string{"tag_name", "name", "body"} {
					if v, ok := rel[k]; ok {
						if s, ok := v.(string); ok && strings.Contains(s, majorMinor) {
							found = true
							break
						}
					}
				}
				if !found {
					continue
				}
				// get published_at
				if pa, ok := rel["published_at"].(string); ok && pa != "" {
					if t, err := time.Parse(time.RFC3339, pa); err == nil {
						return t.Format("2006-01-02")
					}
				}
				// try created_at
				if ca, ok := rel["created_at"].(string); ok && ca != "" {
					if t, err := time.Parse(time.RFC3339, ca); err == nil {
						return t.Format("2006-01-02")
					}
				}
			}
		}
	}

	return ""
}
