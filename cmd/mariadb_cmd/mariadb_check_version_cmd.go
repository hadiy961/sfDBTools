package mariadb_cmd

import (
	"os"

	"sfDBTools/internal/core/mariadb/check_version"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/mariadb"

	"github.com/spf13/cobra"
)

// CheckVersionCmd command for checking available MariaDB versions
var CheckVersionCmd = &cobra.Command{
	Use:   "check_version",
	Short: "Check available MariaDB versions",
	Long: `Check available MariaDB versions from official repositories.
This command fetches the list of supported MariaDB versions that can be installed
using the official MariaDB repository setup script.

Output Control:
- Use --output json for JSON format
- Use 2>/dev/null to suppress operational logs (e.g., for clean JSON output)
- Use --quiet to reduce some operational logs

Examples:
  # Standard table output
  sfdbtools mariadb check_version

  # Clean JSON output (suppress logs)
  sfdbtools mariadb check_version --output json 2>/dev/null

  # With additional details
  sfdbtools mariadb check_version --details --latest-minor`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeVersionCheck(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("Version check failed", logger.Error(err))
			os.Exit(1)
		}
	},
	Annotations: map[string]string{
		"command":  "check_version",
		"category": "mariadb",
	},
}

func executeVersionCheck(cmd *cobra.Command) error {
	// 1. Resolve configuration first
	config, err := mariadb.ResolveVersionConfig(cmd)
	if err != nil {
		return err
	}

	// 2. Get logger
	lg, err := logger.Get()
	if err != nil {
		return err
	}

	// 3. Only log if not in quiet mode
	if !config.Quiet {
		lg.Info("Starting MariaDB version check operation")
	}

	// 4. Create version checker with default config
	checkerConfig := check_version.DefaultConfig()

	// 5. Create version checker
	checker, err := check_version.NewChecker(checkerConfig)
	if err != nil {
		return err
	}

	// 6. Get available versions
	result, err := checker.CheckAvailableVersions()
	if err != nil {
		return err
	}

	// 7. Convert to generic format and display results
	genericResult := &mariadb.GenericVersionResult{
		Versions:  convertToVersionInfoMap(result.AvailableVersions),
		Meta:      convertToGenericMeta(result),
		FetchedAt: result.CheckTime,
	}

	return mariadb.DisplayVersionsFromGenericResult(genericResult, config)
}

// convertToVersionInfoMap converts check_version results to utils format
func convertToVersionInfoMap(versions []check_version.VersionInfo) map[string]mariadb.VersionInfo {
	result := make(map[string]mariadb.VersionInfo)
	for _, info := range versions {
		result[info.Version] = mariadb.VersionInfo{
			Version:     info.Version,
			Type:        info.Type,
			ReleaseDate: info.ReleaseDate,
		}
	}
	return result
}

// convertToGenericMeta converts check_version result to utils format
func convertToGenericMeta(result *check_version.VersionCheckResult) mariadb.GenericMetaInfo {
	osInfo := mariadb.OSInfo{
		OS:   "unknown",
		Arch: "unknown",
	}

	if result.OSInfo != nil {
		osInfo.OS = result.OSInfo.ID
		osInfo.Arch = result.OSInfo.Architecture
	}

	return mariadb.GenericMetaInfo{
		OSInfo: osInfo,
		Sources: []string{
			"MariaDB GitHub API",
			"MariaDB Downloads Page",
			"MariaDB Repository Scripts",
		},
		Count:     len(result.AvailableVersions),
		FetchedAt: result.CheckTime,
	}
}

func init() {
	mariadb.AddCommonVersionFlags(CheckVersionCmd)
}
