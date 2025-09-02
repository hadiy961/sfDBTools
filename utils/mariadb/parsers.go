package mariadb

import (
	"regexp"
	"strings"
	"time"
)

type GitHubAPIParser struct{}

func NewGitHubAPIParser() *GitHubAPIParser {
	return &GitHubAPIParser{}
}

func (p *GitHubAPIParser) ParseVersions(content string) ([]VersionInfo, error) {
	versions := make([]VersionInfo, 0)
	seen := make(map[string]bool)

	// Primary pattern: extract tag_name and published_at
	pattern := regexp.MustCompile(`"tag_name":\s*"mariadb-([^"]+)"[^}]*"published_at":\s*"([^"]*)"`)
	matches := pattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) != 3 {
			continue
		}

		version := strings.TrimSpace(match[1])
		if !seen[version] && IsValidVersion(version) {
			versions = append(versions, VersionInfo{
				Version:     version,
				Type:        DetermineVersionType(version),
				ReleaseDate: parseDate(match[2]),
			})
			seen[version] = true
		}
	}

	// Fallback: tag_name only
	if len(versions) == 0 {
		fallbackPattern := regexp.MustCompile(`"tag_name":\s*"mariadb-([^"]+)"`)
		matches = fallbackPattern.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) != 2 {
				continue
			}

			version := strings.TrimSpace(match[1])
			if !seen[version] && IsValidVersion(version) {
				versions = append(versions, VersionInfo{
					Version: version,
					Type:    DetermineVersionType(version),
				})
				seen[version] = true
			}
		}
	}

	return versions, nil
}

type DownloadsPageParser struct{}

func NewDownloadsPageParser() *DownloadsPageParser {
	return &DownloadsPageParser{}
}

func (p *DownloadsPageParser) ParseVersions(content string) ([]VersionInfo, error) {
	versions := make([]VersionInfo, 0)
	seen := make(map[string]bool)

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)mariadb\s*(\d+\.\d+(?:\.\d+)?(?:-rc\d*|\.rolling)?)`),
		regexp.MustCompile(`(?i)version\s*(\d+\.\d+(?:\.\d+)?(?:-rc\d*|\.rolling)?)`),
		regexp.MustCompile(`(?i)stable.*?(\d+\.\d+(?:\.\d+)?)`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			version := strings.TrimSpace(match[1])
			if !seen[version] && IsValidVersion(version) {
				versionType := DetermineVersionType(version)
				if len(match) > 2 && strings.Contains(strings.ToLower(match[0]), "stable") {
					versionType = "stable"
				}

				versions = append(versions, VersionInfo{
					Version: version,
					Type:    versionType,
				})
				seen[version] = true
			}
		}
	}

	return versions, nil
}

type RepositoryScriptParser struct{}

func NewRepositoryScriptParser() *RepositoryScriptParser {
	return &RepositoryScriptParser{}
}

func (p *RepositoryScriptParser) ParseVersions(content string) ([]VersionInfo, error) {
	versions := make([]VersionInfo, 0)
	seen := make(map[string]bool)

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`mariadb-(\d+\.\d+(?:\.\d+)?(?:-rc\d*|\.rolling)?)`),
		regexp.MustCompile(`"(\d+\.\d+(?:\.\d+)?(?:-rc\d*|\.rolling)?)"`),
		regexp.MustCompile(`--mariadb-server-version[=\s]+["\']?(\d+\.\d+(?:\.\d+)?(?:-rc\d*|\.rolling)?)["\']?`),
		regexp.MustCompile(`version[=\s]*["\']?(\d+\.\d+(?:\.\d+)?(?:-rc\d*|\.rolling)?)["\']?`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			version := strings.TrimSpace(match[1])
			if !seen[version] && IsValidVersion(version) {
				versions = append(versions, VersionInfo{
					Version: version,
					Type:    DetermineVersionType(version),
				})
				seen[version] = true
			}
		}
	}

	return versions, nil
}

func parseDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Format("2006-01-02")
		}
	}

	return ""
}
