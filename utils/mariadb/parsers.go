package mariadb

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Precompiled regex patterns to avoid recompilation on each call
var (
	githubPrimaryPattern  = regexp.MustCompile(`"tag_name":\s*"mariadb-([^"]+)"[^}]*"published_at":\s*"([^"]*)"`)
	githubFallbackPattern = regexp.MustCompile(`"tag_name":\s*"mariadb-([^"]+)"`)

	downloadsPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)mariadb\s*(\d+\.\d+(?:\.\d+)?(?:-rc\d*|\.rolling)?)`),
		regexp.MustCompile(`(?i)version\s*(\d+\.\d+(?:\.\d+)?(?:-rc\d*|\.rolling)?)`),
		regexp.MustCompile(`(?i)stable.*?(\d+\.\d+(?:\.\d+)?)`),
	}

	repoPatterns = []*regexp.Regexp{
		regexp.MustCompile(`mariadb-(\d+\.\d+(?:\.\d+)?(?:-rc\d*|\.rolling)?)`),
		regexp.MustCompile(`"(\d+\.\d+(?:\.\d+)?(?:-rc\d*|\.rolling)?)"`),
		regexp.MustCompile(`--mariadb-server-version[=\s]+["']?(\d+\.\d+(?:\.\d+)?(?:-rc\d*|\.rolling)?) ["']?`),
		regexp.MustCompile(`version[=\s]*["']?(\d+\.\d+(?:\.\d+)?(?:-rc\d*|\.rolling)?) ["']?`),
	}
)

type GitHubAPIParser struct{}

func NewGitHubAPIParser() *GitHubAPIParser {
	return &GitHubAPIParser{}
}

func (p *GitHubAPIParser) ParseVersions(content string) ([]VersionInfo, error) {
	versions := make([]VersionInfo, 0)
	seen := make(map[string]bool)
	// Try to parse as GitHub Releases JSON first (array of releases)
	var releases []struct {
		TagName     string `json:"tag_name"`
		PublishedAt string `json:"published_at"`
	}

	if err := json.Unmarshal([]byte(content), &releases); err == nil {
		for _, r := range releases {
			tag := strings.TrimSpace(r.TagName)
			// strip optional "mariadb-" prefix
			version := strings.TrimPrefix(tag, "mariadb-")
			version = strings.TrimSpace(version)
			if version == "" {
				continue
			}
			if !seen[version] && IsValidVersion(version) {
				addUniqueVersion(&versions, seen, VersionInfo{
					Version:     version,
					Type:        DetermineVersionType(version),
					ReleaseDate: NormalizeDate(r.PublishedAt),
				})
			}
		}
	} else {
		// fallback to regex-based extraction when content isn't JSON
		matches := githubPrimaryPattern.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) != 3 {
				continue
			}

			version := strings.TrimSpace(match[1])
			if !seen[version] && IsValidVersion(version) {
				addUniqueVersion(&versions, seen, VersionInfo{
					Version:     version,
					Type:        DetermineVersionType(version),
					ReleaseDate: NormalizeDate(match[2]),
				})
			}
		}
	}

	// Fallback: tag_name only
	if len(versions) == 0 {
		matchesFB := githubFallbackPattern.FindAllStringSubmatch(content, -1)

		for _, match := range matchesFB {
			if len(match) != 2 {
				continue
			}

			version := strings.TrimSpace(match[1])
			if !seen[version] && IsValidVersion(version) {
				addUniqueVersion(&versions, seen, VersionInfo{
					Version: version,
					Type:    DetermineVersionType(version),
				})
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
	lower := strings.ToLower(content)
	isHTML := strings.Contains(lower, "<html") || strings.Contains(lower, "<!doctype") || strings.Contains(lower, "<body")

	if isHTML {
		// Use goquery to extract candidate texts from common HTML nodes
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(content)))
		if err == nil {
			sel := doc.Find("a, h1, h2, h3, p, span, li")
			sel.Each(func(i int, s *goquery.Selection) {
				txt := strings.TrimSpace(s.Text())
				if txt == "" {
					return
				}
				for _, pattern := range downloadsPatterns {
					matches := pattern.FindAllStringSubmatch(txt, -1)
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
							addUniqueVersion(&versions, seen, VersionInfo{Version: version, Type: versionType})
						}
					}
				}
			})
		}
	}

	// fallback / also scan raw text to catch cases where HTML parsing didn't find versions
	for _, pattern := range downloadsPatterns {
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

				addUniqueVersion(&versions, seen, VersionInfo{
					Version: version,
					Type:    versionType,
				})
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
	for _, pattern := range repoPatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			version := strings.TrimSpace(match[1])
			if !seen[version] && IsValidVersion(version) {
				addUniqueVersion(&versions, seen, VersionInfo{
					Version: version,
					Type:    DetermineVersionType(version),
				})
			}
		}
	}

	return versions, nil
}

// addUniqueVersion appends v to versions if not already seen and marks seen
func addUniqueVersion(versions *[]VersionInfo, seen map[string]bool, v VersionInfo) {
	if v.Version == "" {
		return
	}
	if seen[v.Version] {
		return
	}
	*versions = append(*versions, v)
	seen[v.Version] = true
}

// NormalizeDate tries multiple layouts and returns yyyy-mm-dd or empty string
func NormalizeDate(dateStr string) string {
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
