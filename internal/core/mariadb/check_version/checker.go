package check_version

import (
	"fmt"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/mariadb"
)

// Checker handles MariaDB version checking operations
type Checker struct {
	config   *Config
	osInfo   *common.OSInfo
	analyzer *mariadb.VersionAnalyzer
}

// NewChecker creates a new version checker instance
func NewChecker(config *Config) (*Checker, error) {
	// Detect OS if OS-specific checking is enabled
	var osInfo *common.OSInfo
	var err error

	if config.OSSpecific {
		detector := common.NewOSDetector()
		osInfo, err = detector.DetectOS()
		if err != nil {
			return nil, fmt.Errorf("failed to detect OS: %w", err)
		}
	}

	return &Checker{
		config:   config,
		osInfo:   osInfo,
		analyzer: mariadb.NewVersionAnalyzer(),
	}, nil
}

// CheckAvailableVersions fetches available MariaDB versions from official sources
func (c *Checker) CheckAvailableVersions() (*VersionCheckResult, error) {
	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting MariaDB version check")

	// Create fetchers based on configuration
	fetchers := c.createFetchers()

	// Get versions
	var versions []mariadb.VersionInfo
	if c.config.OSSpecific && c.osInfo != nil {
		versions, err = mariadb.GetVersionsForOS(c.osInfo, fetchers)
	} else {
		versions, err = c.fetchFromAllSources(fetchers)
	}

	if err != nil {
		lg.Error("Failed to fetch supported versions", logger.Error(err))
		return nil, fmt.Errorf("failed to fetch supported versions: %w", err)
	}

	// Sort versions
	sortedVersions := c.analyzer.SortVersions(versions)

	// Process and categorize versions
	result := &VersionCheckResult{
		AvailableVersions: convertToLocalVersionInfo(sortedVersions),
		CheckTime:         time.Now(),
	}

	// Add OS info if available
	if c.osInfo != nil {
		result.OSInfo = c.osInfo
	}

	// Find current stable and latest versions
	result.CurrentStable = c.analyzer.FindCurrentStable(sortedVersions)
	result.LatestVersion = c.analyzer.FindLatestVersion(sortedVersions)
	result.LatestMinor = c.analyzer.FindLatestMinor(sortedVersions)

	lg.Info("MariaDB version check completed",
		logger.Int("versions_found", len(sortedVersions)),
		logger.String("current_stable", result.CurrentStable),
		logger.String("latest_version", result.LatestVersion),
		logger.String("latest_minor", result.LatestMinor))

	return result, nil
}

// createFetchers creates appropriate fetchers based on configuration
func (c *Checker) createFetchers() []mariadb.VersionFetcher {
	fetchers := []mariadb.VersionFetcher{
		mariadb.NewHTTPVersionFetcher(
			"https://mariadb.org/download/",
			mariadb.NewDownloadsPageParser(),
		),
		mariadb.NewHTTPVersionFetcher(
			"https://api.github.com/repos/MariaDB/server/releases",
			mariadb.NewGitHubAPIParser(),
		),
	}

	if c.config.EnableFallback {
		fetchers = append(fetchers, mariadb.NewHTTPVersionFetcher(
			"https://r.mariadb.com/downloads/mariadb_repo_setup",
			mariadb.NewRepositoryScriptParser(),
		))
	}

	return fetchers
}

// fetchFromAllSources fetches versions from all available sources
func (c *Checker) fetchFromAllSources(fetchers []mariadb.VersionFetcher) ([]mariadb.VersionInfo, error) {
	lg, _ := logger.Get()

	var allVersions []mariadb.VersionInfo
	seenVersions := make(map[string]bool)

	for i, fetcher := range fetchers {
		lg.Debug("Trying to fetch versions from source", logger.Int("source_index", i))
		versions, err := fetcher.FetchVersions()
		if err != nil {
			lg.Debug("Source failed, trying next", logger.Int("source_index", i), logger.Error(err))
			continue
		}

		// Add unique versions
		for _, version := range versions {
			if !seenVersions[version.Version] {
				allVersions = append(allVersions, version)
				seenVersions[version.Version] = true
			}
		}

		if len(versions) > 0 {
			lg.Info("Successfully fetched versions from source",
				logger.Int("source_index", i),
				logger.Int("version_count", len(versions)),
				logger.String("fetcher", fetcher.GetName()))
		}
	}

	if len(allVersions) == 0 {
		return nil, fmt.Errorf("no versions found from any source")
	}

	return allVersions, nil
}

// convertToLocalVersionInfo converts mariadb.VersionInfo to local VersionInfo
func convertToLocalVersionInfo(versions []mariadb.VersionInfo) []VersionInfo {
	result := make([]VersionInfo, len(versions))
	for i, v := range versions {
		result[i] = VersionInfo{
			Version:     v.Version,
			Type:        v.Type,
			ReleaseDate: v.ReleaseDate,
		}
	}
	return result
}
