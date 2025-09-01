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

// FindCurrentStable finds the current stable version (typically the latest non-rolling, non-rc)
func (a *VersionAnalyzer) FindCurrentStable(versions []VersionInfo) string {
	var stableVersions []string

	for _, v := range versions {
		if v.Type == "stable" && !strings.Contains(v.Version, "rolling") && !strings.Contains(v.Version, "rc") {
			stableVersions = append(stableVersions, v.Version)
		}
	}

	if len(stableVersions) == 0 {
		return ""
	}

	// Sort and return the latest stable
	sort.Slice(stableVersions, func(i, j int) bool {
		return CompareVersions(stableVersions[i], stableVersions[j])
	})

	return stableVersions[len(stableVersions)-1]
}

// FindLatestVersion finds the absolute latest version (including rolling/rc)
func (a *VersionAnalyzer) FindLatestVersion(versions []VersionInfo) string {
	if len(versions) == 0 {
		return ""
	}

	// Check for rolling first (it's usually the latest)
	for _, v := range versions {
		if strings.Contains(v.Version, "rolling") {
			return v.Version
		}
	}

	// Otherwise find the highest numbered version
	var allVersions []string
	for _, v := range versions {
		if !strings.Contains(v.Version, "rc") { // Exclude RC versions from "latest"
			allVersions = append(allVersions, v.Version)
		}
	}

	if len(allVersions) == 0 {
		return versions[len(versions)-1].Version
	}

	sort.Slice(allVersions, func(i, j int) bool {
		return CompareVersions(allVersions[i], allVersions[j])
	})

	return allVersions[len(allVersions)-1]
}

// FindLatestMinor finds the latest minor version across all major versions
func (a *VersionAnalyzer) FindLatestMinor(versions []VersionInfo) string {
	if len(versions) == 0 {
		return ""
	}

	// Group versions by major version
	majorVersions := make(map[string][]string)

	for _, v := range versions {
		if v.Type != "stable" || strings.Contains(v.Version, "rolling") || strings.Contains(v.Version, "rc") {
			continue // Only consider stable versions
		}

		// Extract major version (e.g., "10" from "10.5", "11" from "11.4")
		parts := strings.Split(v.Version, ".")
		if len(parts) >= 2 {
			major := parts[0]
			majorVersions[major] = append(majorVersions[major], v.Version)
		}
	}

	// Find the latest minor version for each major version
	var latestMinors []string
	for _, minorVersions := range majorVersions {
		if len(minorVersions) == 0 {
			continue
		}

		// Sort minor versions within this major version
		sort.Slice(minorVersions, func(i, j int) bool {
			return CompareVersions(minorVersions[i], minorVersions[j])
		})

		// Get the latest (last) minor version for this major
		latestMinor := minorVersions[len(minorVersions)-1]
		latestMinors = append(latestMinors, latestMinor)
	}

	if len(latestMinors) == 0 {
		return ""
	}

	// Sort all latest minor versions and return the absolute latest
	sort.Slice(latestMinors, func(i, j int) bool {
		return CompareVersions(latestMinors[i], latestMinors[j])
	})

	return latestMinors[len(latestMinors)-1]
}

// SortVersions sorts versions in ascending order
func (a *VersionAnalyzer) SortVersions(versions []VersionInfo) []VersionInfo {
	sorted := make([]VersionInfo, len(versions))
	copy(sorted, versions)

	sort.Slice(sorted, func(i, j int) bool {
		return CompareVersions(sorted[i].Version, sorted[j].Version)
	})

	return sorted
}

// GroupVersionsByType groups versions by their type
func (a *VersionAnalyzer) GroupVersionsByType(versions []VersionInfo) map[string][]VersionInfo {
	groups := make(map[string][]VersionInfo)

	for _, version := range versions {
		groups[version.Type] = append(groups[version.Type], version)
	}

	// Sort each group
	for versionType, versionList := range groups {
		sort.Slice(versionList, func(i, j int) bool {
			return CompareVersions(versionList[i].Version, versionList[j].Version)
		})
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
