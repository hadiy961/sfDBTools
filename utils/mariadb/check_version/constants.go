package check_version

import "time"

const (
	DefaultEndOfLifeAPI      = "https://endoflife.date/api/mariadb/%s.json"
	DefaultGitHubReleasesAPI = "https://api.github.com/repos/MariaDB/server/releases"
	DefaultHTTPTimeout       = 30 * time.Second
	DefaultEOLTimeout        = 3 * time.Second
	DefaultUserAgent         = "sfDBTools/1.0 MariaDB-Version-Checker"
	NoLTS                    = "No LTS"
	TBD                      = "TBD"
)
