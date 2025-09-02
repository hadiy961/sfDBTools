package check_version

import "time"

// VersionInfo represents MariaDB version information
type VersionInfo struct {
	Version     string `json:"version"`
	Type        string `json:"type"` // stable, rc, rolling
	ReleaseDate string `json:"release_date,omitempty"`
	EOLDate     string `json:"eol_date,omitempty"`
}

// VersionFetcher provides interface for fetching MariaDB versions
type VersionFetcher interface {
	FetchVersions() ([]VersionInfo, error)
	GetName() string
}

// VersionParser provides interface for parsing version information
type VersionParser interface {
	ParseVersions(content string) ([]VersionInfo, error)
}

// HTTPVersionFetcher fetches versions via HTTP
type HTTPVersionFetcher struct {
	URL       string
	Timeout   time.Duration
	UserAgent string
	Parser    VersionParser
}
