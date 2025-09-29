// Package disk provides disk space checking utilities.
// This package focuses on core disk space validation functionality.
// For detailed disk usage statistics, use the usage.go module.
// For disk monitoring, use the monitor.go module.
package disk

import (
	"fmt"
	"math"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/common/format"
	"sfDBTools/utils/fs"
)

// CheckDiskSpace checks if at least minFreeSpace (MB) is available
// on the filesystem containing the given path. If path is empty, root is used.
// minFreeSpace is assumed to be in megabytes.
func CheckDiskSpace(path string, minFreeSpace int64) error {
	lg, _ := logger.Get()

	// Use fs package for consistent path handling and disk usage
	fsManager := fs.NewManager()
	usage, err := fsManager.Dir().GetDiskUsage(path)
	if err != nil {
		lg.Error("Failed to get disk usage info",
			logger.String("path", path),
			logger.Error(err))
		return fmt.Errorf("failed to get disk usage info for '%s': %w", path, err)
	}

	required := mbToBytes(minFreeSpace)
	percentFree := 100.0 - usage.UsedPercent

	if usage.Free < required {
		lg.Error("Insufficient disk space",
			logger.String("available", common.FormatSizeWithPrecision(usage.Free, 2)),
			logger.String("required", common.FormatSizeWithPrecision(required, 2)),
			logger.String("checked_path", usage.Path),
			logger.String("percent_free", format.FormatPercent(percentFree, 1)))
		return fmt.Errorf("insufficient disk space: available %s, required %s (free %s)",
			common.FormatSizeWithPrecision(usage.Free, 2),
			common.FormatSizeWithPrecision(required, 2),
			format.FormatPercent(percentFree, 1))
	}

	lg.Debug("Disk space check passed",
		logger.String("available", common.FormatSizeWithPrecision(usage.Free, 2)),
		logger.String("checked_path", usage.Path),
		logger.String("percent_free", format.FormatPercent(percentFree, 1)))
	return nil
}

// CheckDiskSpaceBytes checks free bytes threshold instead of MB.
func CheckDiskSpaceBytes(path string, minFreeBytes int64) error {
	lg, _ := logger.Get()

	fsManager := fs.NewManager()
	usage, err := fsManager.Dir().GetDiskUsage(path)
	if err != nil {
		lg.Error("Failed to get disk usage info",
			logger.String("path", path),
			logger.Error(err))
		return fmt.Errorf("failed to get disk usage info for '%s': %w", path, err)
	}

	if usage.Free < minFreeBytes {
		lg.Error("Insufficient disk space",
			logger.String("available", common.FormatSizeWithPrecision(usage.Free, 2)),
			logger.String("required", common.FormatSizeWithPrecision(minFreeBytes, 2)),
			logger.String("path", usage.Path))
		return fmt.Errorf("insufficient disk space: available %s, required %s",
			common.FormatSizeWithPrecision(usage.Free, 2),
			common.FormatSizeWithPrecision(minFreeBytes, 2))
	}

	lg.Debug("Disk space check passed",
		logger.String("available", common.FormatSizeWithPrecision(usage.Free, 2)),
		logger.String("required", common.FormatSizeWithPrecision(minFreeBytes, 2)),
		logger.String("path", usage.Path))
	return nil
}

// mbToBytes converts megabytes to bytes safely, preventing overflow.
func mbToBytes(mb int64) int64 {
	const mbUnit = int64(1024 * 1024)
	if mb <= 0 {
		return 0
	}
	if mb > math.MaxInt64/mbUnit {
		return math.MaxInt64
	}
	return mb * mbUnit
}
