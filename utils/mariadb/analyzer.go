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
	var stableVersions []string

	for _, v := range versions {
		if v.Type == "stable" && !strings.Contains(v.Version, "rolling") && !strings.Contains(v.Version, "rc") {
			stableVersions = append(stableVersions, v.Version)
		}
	}

	if len(stableVersions) == 0 {
		return ""
	}

	sort.Slice(stableVersions, func(i, j int) bool {
		return CompareVersions(stableVersions[i], stableVersions[j])
	})

	return stableVersions[len(stableVersions)-1]
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

	var allVersions []string
	for _, v := range versions {
		if !strings.Contains(v.Version, "rc") {
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

		sort.Slice(minorVersions, func(i, j int) bool {
			return CompareVersions(minorVersions[i], minorVersions[j])
		})

		latestMinors = append(latestMinors, minorVersions[len(minorVersions)-1])
	}

	if len(latestMinors) == 0 {
		return ""
	}

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
