package mariadb

import (
	"sort"
	"strings"
)

// VersionAnalyzer analyzes MariaDB versions
type VersionAnalyzer struct{}

// NewVersionAnalyzer creates a new version analyzer
func NewVersionAnalyzer() *VersionAnalyzer {
	return &VersionAnalyzer{}
}

// FindCurrentStable finds the current stable version
func (a *VersionAnalyzer) FindCurrentStable(versions []VersionInfo) string {
	stableVersions := filterVersionStrings(versions, func(v VersionInfo) bool {
		return v.Type == "stable" && !strings.Contains(v.Version, "rolling") && !strings.Contains(v.Version, "rc")
	})

	return latestFromStrings(stableVersions)
}

// FindLatestVersion finds the absolute latest version
func (a *VersionAnalyzer) FindLatestVersion(versions []VersionInfo) string {
	if len(versions) == 0 {
		return ""
	}
	for _, v := range versions {
		if strings.Contains(v.Version, "rolling") {
			return v.Version
		}
	}

	allVersions := filterVersionStrings(versions, func(v VersionInfo) bool {
		return !strings.Contains(v.Version, "rc")
	})

	if len(allVersions) == 0 {
		return versions[len(versions)-1].Version
	}

	return latestFromStrings(allVersions)
}

// FindLatestMinor finds the latest minor version across all major versions
func (a *VersionAnalyzer) FindLatestMinor(versions []VersionInfo) string {
	if len(versions) == 0 {
		return ""
	}
	// collect stable, non-rolling, non-rc versions grouped by major
	majorVersions := make(map[string][]string)

	for _, v := range versions {
		if v.Type != "stable" || strings.Contains(v.Version, "rolling") || strings.Contains(v.Version, "rc") {
			continue
		}

		parts := strings.Split(v.Version, ".")
		if len(parts) >= 2 {
			majorVersions[parts[0]] = append(majorVersions[parts[0]], v.Version)
		}
	}

	var latestMinors []string
	for _, minorVersions := range majorVersions {
		if len(minorVersions) == 0 {
			continue
		}

		latestMinors = append(latestMinors, latestFromStrings(minorVersions))
	}

	return latestFromStrings(latestMinors)
}

// SortVersions sorts versions in ascending order
func (a *VersionAnalyzer) SortVersions(versions []VersionInfo) []VersionInfo {
	sorted := make([]VersionInfo, len(versions))
	copy(sorted, versions)

	sortVersionInfoSlice(sorted)
	return sorted
}

// GroupVersionsByType groups versions by their type
func (a *VersionAnalyzer) GroupVersionsByType(versions []VersionInfo) map[string][]VersionInfo {
	groups := make(map[string][]VersionInfo)

	for _, version := range versions {
		groups[version.Type] = append(groups[version.Type], version)
	}

	for versionType, versionList := range groups {
		sortVersionInfoSlice(versionList)
		groups[versionType] = versionList
	}

	return groups
}

// FilterVersionsByMajor filters versions by major version number
func (a *VersionAnalyzer) FilterVersionsByMajor(versions []VersionInfo, majorVersion string) []VersionInfo {
	var filtered []VersionInfo

	for _, version := range versions {
		parts := strings.Split(version.Version, ".")
		if len(parts) > 0 && parts[0] == majorVersion {
			filtered = append(filtered, version)
		}
	}

	return filtered
}

// --- helper reusable functions ---

// filterVersionStrings returns a slice of version strings for which pred returns true
func filterVersionStrings(versions []VersionInfo, pred func(VersionInfo) bool) []string {
	var out []string
	for _, v := range versions {
		if pred(v) {
			out = append(out, v.Version)
		}
	}
	return out
}

// sortVersionStringsAsc sorts a slice of version strings in ascending order using CompareVersions
func sortVersionStringsAsc(list []string) {
	sort.Slice(list, func(i, j int) bool {
		return CompareVersions(list[i], list[j])
	})
}

// latestFromStrings returns the latest (largest) version string from the slice or empty string
func latestFromStrings(list []string) string {
	if len(list) == 0 {
		return ""
	}
	sortVersionStringsAsc(list)
	return list[len(list)-1]
}

// sortVersionInfoSlice sorts a slice of VersionInfo in-place by their Version field
func sortVersionInfoSlice(list []VersionInfo) {
	sort.Slice(list, func(i, j int) bool {
		return CompareVersions(list[i].Version, list[j].Version)
	})
}
