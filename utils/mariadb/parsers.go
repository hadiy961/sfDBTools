package mariadb

import (
	"regexp"
	"strings"
)

// GitHubAPIParser parses MariaDB versions from GitHub releases API
type GitHubAPIParser struct{}

// NewGitHubAPIParser creates a new GitHub API parser
func NewGitHubAPIParser() *GitHubAPIParser {
	return &GitHubAPIParser{}
}

// ParseVersions implements VersionParser interface for GitHub API
func (p *GitHubAPIParser) ParseVersions(content string) ([]VersionInfo, error) {
	var versions []VersionInfo
	seenVersions := make(map[string]bool)

	// Look for tag names in the GitHub API response
	tagRegex := regexp.MustCompile(`"tag_name":\s*"([^"]*)"`)
	matches := tagRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			tag := strings.TrimSpace(match[1])
			// Extract version from tag (e.g., "mariadb-10.6.15" -> "10.6.15")
			versionMatch := regexp.MustCompile(`mariadb-(\d+\.\d+(?:\.\d+)?(?:-rc\d*)?)`).FindStringSubmatch(tag)
			if len(versionMatch) > 1 {
				version := versionMatch[1]
				if !seenVersions[version] && IsValidVersion(version) {
					versionType := DetermineVersionType(version)
					versions = append(versions, VersionInfo{
						Version: version,
						Type:    versionType,
					})
					seenVersions[version] = true
				}
			}
		}
	}

	return versions, nil
}

// DownloadsPageParser parses MariaDB versions from downloads page
type DownloadsPageParser struct{}

// NewDownloadsPageParser creates a new downloads page parser
func NewDownloadsPageParser() *DownloadsPageParser {
	return &DownloadsPageParser{}
}

// ParseVersions implements VersionParser interface for downloads page
func (p *DownloadsPageParser) ParseVersions(content string) ([]VersionInfo, error) {
	var versions []VersionInfo
	seenVersions := make(map[string]bool)

	// Look for version patterns in the downloads page
	versionRegex := regexp.MustCompile(`(?i)(?:mariadb|version)\s*(\d+\.\d+(?:\.\d+)?(?:-rc\d*|\.rolling)?)`)
	matches := versionRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			version := strings.TrimSpace(match[1])
			if !seenVersions[version] && IsValidVersion(version) {
				versionType := DetermineVersionType(version)
				versions = append(versions, VersionInfo{
					Version: version,
					Type:    versionType,
				})
				seenVersions[version] = true
			}
		}
	}

	// Also look for more specific patterns
	stableVersionRegex := regexp.MustCompile(`(?i)stable.*?(\d+\.\d+(?:\.\d+)?)`)
	stableMatches := stableVersionRegex.FindAllStringSubmatch(content, -1)

	for _, match := range stableMatches {
		if len(match) > 1 {
			version := strings.TrimSpace(match[1])
			if !seenVersions[version] && IsValidVersion(version) {
				versions = append(versions, VersionInfo{
					Version: version,
					Type:    "stable",
				})
				seenVersions[version] = true
			}
		}
	}

	return versions, nil
}

// RepositoryScriptParser parses MariaDB versions from repository setup script
type RepositoryScriptParser struct{}

// NewRepositoryScriptParser creates a new repository script parser
func NewRepositoryScriptParser() *RepositoryScriptParser {
	return &RepositoryScriptParser{}
}

// ParseVersions implements VersionParser interface for repository script
func (p *RepositoryScriptParser) ParseVersions(content string) ([]VersionInfo, error) {
	var versions []VersionInfo
	seenVersions := make(map[string]bool)

	// Look for mariadb version patterns in the script
	patterns := []string{
		`mariadb-(\d+\.\d+(?:\.\d+)?(?:-rc\d*)?(?:\.rolling)?)`,                                   // mariadb-10.6, mariadb-11.4
		`"(\d+\.\d+(?:\.\d+)?(?:-rc\d*)?(?:\.rolling)?)"`,                                         // quoted versions
		`--mariadb-server-version[=\s]+["\']?(\d+\.\d+(?:\.\d+)?(?:-rc\d*)?(?:\.rolling)?)["\']?`, // script parameters
		`version[=\s]*["\']?(\d+\.\d+(?:\.\d+)?(?:-rc\d*)?(?:\.rolling)?)["\']?`,                  // version parameters
	}

	for _, pattern := range patterns {
		versionRegex := regexp.MustCompile(pattern)
		matches := versionRegex.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) > 1 {
				version := strings.TrimSpace(match[1])
				if !seenVersions[version] && IsValidVersion(version) {
					versionType := DetermineVersionType(version)
					versions = append(versions, VersionInfo{
						Version: version,
						Type:    versionType,
					})
					seenVersions[version] = true
				}
			}
		}
	}

	// Look for supported versions list in comments or documentation within the script
	supportedRegex := regexp.MustCompile(`(?i)(?:supported|available).*?versions?.*?:?\s*((?:\d+\.\d+(?:\.\d+)?(?:-rc\d*)?(?:\.rolling)?\s*,?\s*)+)`)
	supportedMatches := supportedRegex.FindAllStringSubmatch(content, -1)

	for _, match := range supportedMatches {
		if len(match) > 1 {
			versionList := match[1]
			// Extract individual versions from the list
			individualVersions := regexp.MustCompile(`(\d+\.\d+(?:\.\d+)?(?:-rc\d*)?(?:\.rolling)?)`).FindAllString(versionList, -1)
			for _, version := range individualVersions {
				version = strings.TrimSpace(version)
				if !seenVersions[version] && IsValidVersion(version) {
					versionType := DetermineVersionType(version)
					versions = append(versions, VersionInfo{
						Version: version,
						Type:    versionType,
					})
					seenVersions[version] = true
				}
			}
		}
	}

	return versions, nil
}
