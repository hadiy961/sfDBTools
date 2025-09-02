package check_version

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
)

// GetVersionsForOS returns MariaDB versions available for the specified OS
func GetVersionsForOS(osInfo *common.OSInfo, fetchers []VersionFetcher) ([]VersionInfo, error) {
	lg, _ := logger.Get()

	lg.Info("Fetching MariaDB versions for OS",
		logger.String("os_id", osInfo.ID),
		logger.String("os_version", osInfo.Version),
		logger.String("package_type", osInfo.PackageType))

	var allVersions []VersionInfo
	seenVersions := make(map[string]bool)

	for _, fetcher := range fetchers {
		lg.Debug("Trying fetcher", logger.String("fetcher", fetcher.GetName()))

		versions, err := fetcher.FetchVersions()
		if err != nil {
			lg.Warn("Fetcher failed",
				logger.String("fetcher", fetcher.GetName()),
				logger.Error(err))
			continue
		}

		// Filter versions based on OS compatibility
		for _, version := range versions {
			if !seenVersions[version.Version] && isVersionCompatibleWithOS(version, osInfo) {
				allVersions = append(allVersions, version)
				seenVersions[version.Version] = true
			}
		}

		if len(versions) > 0 {
			lg.Info("Successfully fetched versions",
				logger.String("fetcher", fetcher.GetName()),
				logger.Int("version_count", len(versions)))
		}
	}

	if len(allVersions) == 0 {
		return nil, fmt.Errorf("no compatible versions found for OS %s %s", osInfo.ID, osInfo.Version)
	}

	return allVersions, nil
}

// isVersionCompatibleWithOS checks if a version is compatible with the given OS
func isVersionCompatibleWithOS(version VersionInfo, osInfo *common.OSInfo) bool {
	// All stable versions are generally compatible
	// This can be enhanced with more specific OS/version compatibility rules

	// Basic compatibility rules
	switch osInfo.ID {
	case "ubuntu", "debian":
		return osInfo.PackageType == "deb"
	case "centos", "rhel", "rocky", "almalinux":
		return osInfo.PackageType == "rpm"
	case "sles":
		return osInfo.PackageType == "rpm"
	default:
		// For unknown OS, assume compatibility
		return true
	}
}
