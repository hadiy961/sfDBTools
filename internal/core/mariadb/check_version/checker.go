package check_version

import (
	"context"
	"fmt"
	"sync"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/mariadb"
	"sfDBTools/utils/terminal"

	"golang.org/x/sync/errgroup"
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

// CheckAvailableVersions is the backwards-compatible entry point
func (c *Checker) CheckAvailableVersions() (*VersionCheckResult, error) {
	return c.CheckAvailableVersionsWithCtx(context.Background())
}

// CheckAvailableVersionsWithCtx fetches available MariaDB versions and respects ctx cancellation
func (c *Checker) CheckAvailableVersionsWithCtx(ctx context.Context) (*VersionCheckResult, error) {
	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting MariaDB version check")
	// show spinner while fetching versions
	spinner := terminal.NewProgressSpinner("Checking MariaDB versions...")
	spinner.Start()
	defer spinner.Stop()

	// Create fetchers based on configuration
	fetchers := c.createFetchers()

	// Get versions
	var versions []mariadb.VersionInfo
	if c.config.OSSpecific && c.osInfo != nil {
		// delegate to utils wrapper (keeps existing behavior)
		versions, err = mariadb.GetVersionsForOS(c.osInfo, fetchers)
	} else {
		versions, err = c.fetchFromAllSourcesWithCtx(ctx, fetchers)
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

// fetchFromAllSourcesWithCtx fetches versions from all available sources using context for cancellation
// This implementation runs fetchers in parallel with a bounded concurrency limit to reduce total latency.
func (c *Checker) fetchFromAllSourcesWithCtx(ctx context.Context, fetchers []mariadb.VersionFetcher) ([]mariadb.VersionInfo, error) {
	lg, _ := logger.Get()

	if len(fetchers) == 0 {
		return nil, fmt.Errorf("no fetchers configured")
	}

	// bounded concurrency
	maxConcurrency := 4
	if len(fetchers) < maxConcurrency {
		maxConcurrency = len(fetchers)
	}

	sem := make(chan struct{}, maxConcurrency)
	g, gctx := errgroup.WithContext(ctx)

	var mu sync.Mutex
	allVersions := make([]mariadb.VersionInfo, 0)
	seenVersions := make(map[string]bool)

	for i, fetcher := range fetchers {
		i := i
		fetcher := fetcher

		if gctx.Err() != nil {
			break
		}

		sem <- struct{}{}
		g.Go(func() error {
			defer func() { <-sem }()

			lg.Debug("Trying to fetch versions from source", logger.Int("source_index", i))

			var versions []mariadb.VersionInfo
			var err error

			if ctxFetcher, ok := fetcher.(interface {
				FetchVersionsWithCtx(context.Context) ([]mariadb.VersionInfo, error)
			}); ok {
				versions, err = ctxFetcher.FetchVersionsWithCtx(gctx)
			} else {
				versions, err = fetcher.FetchVersions()
			}

			if err != nil {
				lg.Debug("Source failed, trying next", logger.Int("source_index", i), logger.Error(err))
				return nil // don't fail whole group for single-source failure
			}

			mu.Lock()
			for _, version := range versions {
				if !seenVersions[version.Version] {
					allVersions = append(allVersions, version)
					seenVersions[version.Version] = true
				}
			}
			mu.Unlock()

			if len(versions) > 0 {
				lg.Info("Successfully fetched versions from source",
					logger.Int("source_index", i),
					logger.Int("version_count", len(versions)),
					logger.String("fetcher", fetcher.GetName()))
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
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
			// EOLDate will be computed lazily when rendering to avoid unnecessary external calls
			EOLDate: "",
		}
	}
	return result
}
