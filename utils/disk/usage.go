// Package disk provides disk usage statistics and reporting functionality.
// This file contains functions for getting detailed disk usage information,
// partition listings, and finding optimal storage locations.
package disk

import (
	"fmt"
	"math"

	gopsutildisk "github.com/shirou/gopsutil/v3/disk"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/common/format"
	"sfDBTools/utils/fs"
)

// UsageStatistics provides detailed disk usage information.
type UsageStatistics struct {
	Path        string  `json:"path"`
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	Total       int64   `json:"total"`
	Free        int64   `json:"free"`
	Used        int64   `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

// GetUsageStatistics returns detailed disk usage statistics for the given path.
// This function provides more detailed information than the basic fs.DiskUsage.
func GetUsageStatistics(path string) (*UsageStatistics, error) {
	lg, _ := logger.Get()

	// Use fs package for path resolution and basic usage
	fsManager := fs.NewManager()
	fsUsage, err := fsManager.Dir().GetDiskUsage(path)
	if err != nil {
		lg.Error("Failed to get disk usage statistics",
			logger.String("path", path),
			logger.Error(err))
		return nil, fmt.Errorf("failed to get disk usage statistics for '%s': %w", path, err)
	}

	// Get additional system information using gopsutil directly
	usage, err := gopsutildisk.Usage(fsUsage.Path)
	if err != nil {
		lg.Error("Failed to get detailed disk statistics",
			logger.String("path", path),
			logger.Error(err))
		return nil, fmt.Errorf("failed to get detailed disk statistics for '%s': %w", path, err)
	}

	// Convert to our statistics format with safe int64 conversion
	stats := &UsageStatistics{
		Path:        path,
		Mountpoint:  usage.Path,
		Fstype:      usage.Fstype,
		Total:       safeUint64ToInt64(usage.Total),
		Free:        safeUint64ToInt64(usage.Free),
		Used:        safeUint64ToInt64(usage.Used),
		UsedPercent: usage.UsedPercent,
	}

	lg.Debug("Disk usage statistics retrieved",
		logger.String("path", path),
		logger.String("mountpoint", stats.Mountpoint),
		logger.String("fstype", stats.Fstype),
		logger.String("total", common.FormatSizeWithPrecision(stats.Total, 2)),
		logger.String("free", common.FormatSizeWithPrecision(stats.Free, 2)),
		logger.Float64("used_percent", stats.UsedPercent))

	return stats, nil
}

// GetFreeBytes returns free bytes for the specified path.
func GetFreeBytes(path string) (int64, error) {
	stats, err := GetUsageStatistics(path)
	if err != nil {
		return 0, err
	}
	return stats.Free, nil
}

// GetTotalBytes returns total bytes for the specified path.
func GetTotalBytes(path string) (int64, error) {
	stats, err := GetUsageStatistics(path)
	if err != nil {
		return 0, err
	}
	return stats.Total, nil
}

// GetUsedPercent returns used percentage for the specified path.
func GetUsedPercent(path string) (float64, error) {
	stats, err := GetUsageStatistics(path)
	if err != nil {
		return 0, err
	}
	return stats.UsedPercent, nil
}

// GetAllPartitions returns usage statistics for all mounted partitions.
// This is similar to the `df -h` command output.
func GetAllPartitions() ([]*UsageStatistics, error) {
	lg, _ := logger.Get()

	partitions, err := gopsutildisk.Partitions(false)
	if err != nil {
		lg.Error("Failed to list partitions", logger.Error(err))
		return nil, fmt.Errorf("failed to list partitions: %w", err)
	}

	var results []*UsageStatistics
	for _, partition := range partitions {
		// Skip partitions that may not be accessible
		usage, err := gopsutildisk.Usage(partition.Mountpoint)
		if err != nil {
			lg.Debug("Skipping partition due to access error",
				logger.String("mountpoint", partition.Mountpoint),
				logger.Error(err))
			continue
		}

		stats := &UsageStatistics{
			Path:        partition.Mountpoint,
			Mountpoint:  partition.Mountpoint,
			Fstype:      usage.Fstype,
			Total:       safeUint64ToInt64(usage.Total),
			Free:        safeUint64ToInt64(usage.Free),
			Used:        safeUint64ToInt64(usage.Used),
			UsedPercent: usage.UsedPercent,
		}

		results = append(results, stats)
	}

	lg.Debug("Retrieved partition statistics",
		logger.Int("partition_count", len(results)))

	return results, nil
}

// FindBestStorageLocation finds the path with the most free space from the given candidates.
// This is useful for selecting optimal locations for backups or temporary files.
func FindBestStorageLocation(candidates []string) (*UsageStatistics, error) {
	lg, _ := logger.Get()

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no candidate paths provided")
	}

	var bestStats *UsageStatistics
	var bestFree int64 = -1

	for _, candidate := range candidates {
		stats, err := GetUsageStatistics(candidate)
		if err != nil {
			lg.Debug("Skipping candidate due to error",
				logger.String("path", candidate),
				logger.Error(err))
			continue
		}

		if stats.Free > bestFree {
			bestStats = stats
			bestFree = stats.Free
		}
	}

	if bestStats == nil {
		return nil, fmt.Errorf("no valid candidate paths found")
	}

	lg.Info("Best storage location selected",
		logger.String("path", bestStats.Path),
		logger.String("mountpoint", bestStats.Mountpoint),
		logger.String("free_space", common.FormatSizeWithPrecision(bestStats.Free, 2)),
		logger.Float64("used_percent", bestStats.UsedPercent))

	return bestStats, nil
}

// FormatUsageReport returns a human-readable string representation of disk usage statistics.
func (s *UsageStatistics) FormatUsageReport() string {
	return fmt.Sprintf(
		"Path: %s\nMountpoint: %s\nFilesystem: %s\nTotal: %s\nUsed: %s (%s)\nFree: %s",
		s.Path,
		s.Mountpoint,
		s.Fstype,
		common.FormatSizeWithPrecision(s.Total, 2),
		common.FormatSizeWithPrecision(s.Used, 2),
		format.FormatPercent(s.UsedPercent, 1),
		common.FormatSizeWithPrecision(s.Free, 2),
	)
}

// safeUint64ToInt64 converts uint64 to int64 safely, preventing overflow.
func safeUint64ToInt64(val uint64) int64 {
	if val > uint64(math.MaxInt64) {
		return math.MaxInt64
	}
	return int64(val)
}
