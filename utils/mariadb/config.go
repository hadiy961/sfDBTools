package mariadb

import (
	"sfDBTools/utils/common"
	"time"

	"github.com/spf13/cobra"
)

const (
	DefaultEndOfLifeAPI      = "https://endoflife.date/api/mariadb/%s.json"
	DefaultGitHubReleasesAPI = "https://api.github.com/repos/MariaDB/server/releases"
	DefaultHTTPTimeout       = 30 * time.Second
	DefaultEOLTimeout        = 3 * time.Second
	DefaultUserAgent         = "sfDBTools/1.0 MariaDB-Version-Checker"

	NoLTS = "No LTS"
	TBD   = "TBD"
)

// VersionConfig holds configuration for version checking operations
type VersionConfig struct {
	ShowDetails     bool   `json:"show_details"`
	OutputFormat    string `json:"output_format"` // table, json, simple
	Format          string `json:"format"`        // alias for OutputFormat
	ShowLatestMinor bool   `json:"show_latest_minor"`
	Quiet           bool   `json:"quiet"` // suppress operational logs
}

// AddCommonVersionFlags adds common flags for version checking commands
func AddCommonVersionFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("details", false, "Show detailed version information")
	cmd.Flags().String("output", "table", "Output format (table, json, simple)")
	cmd.Flags().Bool("latest-minor", false, "Show latest minor version for each major version")
	cmd.Flags().Bool("quiet", false, "Suppress operational logs, show only results")
}

// ResolveVersionConfig resolves version checking configuration from command flags
func ResolveVersionConfig(cmd *cobra.Command) (*VersionConfig, error) {
	config := &VersionConfig{}

	// Get flags using shared helpers
	config.ShowDetails = common.GetBoolFlagOrEnv(cmd, "details", "SFDBTOOLS_SHOW_DETAILS", false)
	config.OutputFormat = common.GetStringFlagOrEnv(cmd, "output", "SFDBTOOLS_OUTPUT_FORMAT", "table")
	config.Format = config.OutputFormat // Set alias
	config.ShowLatestMinor = common.GetBoolFlagOrEnv(cmd, "latest-minor", "SFDBTOOLS_SHOW_LATEST_MINOR", false)
	config.Quiet = common.GetBoolFlagOrEnv(cmd, "quiet", "SFDBTOOLS_QUIET", false)

	return config, nil
}
