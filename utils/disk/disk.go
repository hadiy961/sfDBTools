// Package disk menyediakan utilitas pemeriksaan disk (multi-platform).
// Sebelumnya file ini hanya untuk Linux; build tag dihapus supaya bisa
// dikompilasi di platform lain. Implementasi menggunakan gopsutil yang
// bersifat cross-platform.
package disk

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	gopsutildisk "github.com/shirou/gopsutil/v3/disk"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
)

// CheckDiskSpace mengecek apakah tersedia minimal `minFreeSpace` (MB)
// pada filesystem yang menampung `path`. Jika `path` kosong, root akan
// dipakai. Nilai `minFreeSpace` diasumsikan dalam megabyte.
func CheckDiskSpace(path string, minFreeSpace int64) error {
	lg, _ := logger.Get()

	// normalize and find an existing directory to check
	checkPath := path
	if checkPath == "" {
		checkPath = string(os.PathSeparator)
	}

	// Walk up until we find an existing path (or reach root)
	for {
		if _, err := os.Stat(checkPath); err != nil {
			if os.IsNotExist(err) {
				parent := filepath.Dir(checkPath)
				if parent == checkPath || parent == "." || parent == string(os.PathSeparator) {
					// as a last resort use root
					checkPath = string(os.PathSeparator)
					break
				}
				checkPath = parent
				continue
			}
			lg.Error("Failed to stat path for disk check",
				logger.String("path", checkPath),
				logger.Error(err))
			return fmt.Errorf("failed to access path '%s': %w", checkPath, err)
		}
		break
	}

	// Gunakan gopsutil untuk mengambil informasi penggunaan disk
	usage, err := gopsutildisk.Usage(checkPath)
	if err != nil {
		lg.Error("Failed to get disk usage info",
			logger.String("path", checkPath),
			logger.Error(err))
		return fmt.Errorf("failed to get disk usage info for '%s': %w", checkPath, err)
	}
	// gopsutil mengembalikan Free sebagai uint64, jaga agar tidak overflow
	var available int64
	if usage.Free > uint64(math.MaxInt64) {
		available = math.MaxInt64
	} else {
		available = int64(usage.Free)
	}

	required := mbToBytes(minFreeSpace)

	percentFree := 100.0 - usage.UsedPercent // aproksimasi

	if available < required {
		lg.Error("Ruang disk tidak mencukupi",
			logger.String("available", common.FormatSizeWithPrecision(available, 2)),
			logger.String("required", common.FormatSizeWithPrecision(required, 2)),
			logger.String("checked_path", checkPath),
			logger.String("percent_free", common.FormatPercent(percentFree, 1)))
		return fmt.Errorf("insufficient disk space: available %s, required %s (free %s)",
			common.FormatSizeWithPrecision(available, 2), common.FormatSizeWithPrecision(required, 2), common.FormatPercent(percentFree, 1))
	}

	lg.Debug("Disk space check passed",
		logger.String("available", common.FormatSizeWithPrecision(available, 2)),
		logger.String("checked_path", checkPath),
		logger.String("percent_free", common.FormatPercent(percentFree, 1)))
	return nil
}

// DiskUsage represents usage metrics for a filesystem/mountpoint.
type DiskUsage struct {
	Path        string
	Mountpoint  string
	Fstype      string
	Total       int64
	Free        int64
	Used        int64
	UsedPercent float64
}

// DiskProvider abstracts gopsutil so it can be mocked in tests.
type DiskProvider interface {
	Usage(path string) (*gopsutildisk.UsageStat, error)
}

type realDiskProvider struct{}

func (r realDiskProvider) Usage(path string) (*gopsutildisk.UsageStat, error) {
	return gopsutildisk.Usage(path)
}

var defaultDiskProvider DiskProvider = realDiskProvider{}

// GetUsage returns DiskUsage for the given path using the DiskProvider.
func GetUsage(path string) (DiskUsage, error) {
	lg, _ := logger.Get()
	checkPath := path
	if checkPath == "" {
		checkPath = string(os.PathSeparator)
	}

	// Walk up to find existing path as in CheckDiskSpace
	for {
		if _, err := os.Stat(checkPath); err != nil {
			if os.IsNotExist(err) {
				parent := filepath.Dir(checkPath)
				if parent == checkPath || parent == "." || parent == string(os.PathSeparator) {
					checkPath = string(os.PathSeparator)
					break
				}
				checkPath = parent
				continue
			}
			lg.Error("Failed to stat path for GetUsage", logger.String("path", checkPath), logger.Error(err))
			return DiskUsage{}, fmt.Errorf("failed to access path '%s': %w", checkPath, err)
		}
		break
	}

	usage, err := defaultDiskProvider.Usage(checkPath)
	if err != nil {
		lg.Error("Failed to get disk usage info", logger.String("path", checkPath), logger.Error(err))
		return DiskUsage{}, fmt.Errorf("failed to get disk usage info for '%s': %w", checkPath, err)
	}

	var free int64
	if usage.Free > uint64(math.MaxInt64) {
		free = math.MaxInt64
	} else {
		free = int64(usage.Free)
	}

	var total int64
	if usage.Total > uint64(math.MaxInt64) {
		total = math.MaxInt64
	} else {
		total = int64(usage.Total)
	}

	used := total - free

	du := DiskUsage{
		Path:        path,
		Mountpoint:  usage.Path,
		Fstype:      usage.Fstype,
		Total:       total,
		Free:        free,
		Used:        used,
		UsedPercent: usage.UsedPercent,
	}
	return du, nil
}

// GetFreeBytes returns free bytes for path.
func GetFreeBytes(path string) (int64, error) {
	u, err := GetUsage(path)
	if err != nil {
		return 0, err
	}
	return u.Free, nil
}

// GetTotalBytes returns total bytes for path.
func GetTotalBytes(path string) (int64, error) {
	u, err := GetUsage(path)
	if err != nil {
		return 0, err
	}
	return u.Total, nil
}

// GetUsedPercent returns used percent for path.
func GetUsedPercent(path string) (float64, error) {
	u, err := GetUsage(path)
	if err != nil {
		return 0, err
	}
	return u.UsedPercent, nil
}

// CheckDiskSpaceBytes checks free bytes threshold instead of MB.
func CheckDiskSpaceBytes(path string, minFreeBytes int64) error {
	available, err := GetFreeBytes(path)
	if err != nil {
		return err
	}
	if available < minFreeBytes {
		return fmt.Errorf("insufficient disk space: available %s, required %s",
			common.FormatSizeWithPrecision(available, 2), common.FormatSizeWithPrecision(minFreeBytes, 2))
	}
	return nil
}

// MonitorDisk starts a background monitor that calls cb when used percent
// exceeds thresholdPercent. It returns a stop function to stop monitoring.
func MonitorDisk(path string, interval time.Duration, thresholdPercent float64, cb func(DiskUsage)) func() {
	stopCh := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				u, err := GetUsage(path)
				if err != nil {
					// don't spam logs, just continue
					lg, _ := logger.Get()
					lg.Error("MonitorDisk: failed to get usage", logger.Error(err))
					continue
				}
				if u.UsedPercent >= thresholdPercent {
					cb(u)
				}
			}
		}
	}()
	return func() { close(stopCh) }
}

// FindBestOutputMount picks the path with the most free bytes from candidates.
func FindBestOutputMount(candidates []string) (DiskUsage, error) {
	var best DiskUsage
	var found bool
	for _, p := range candidates {
		u, err := GetUsage(p)
		if err != nil {
			continue
		}
		if !found || u.Free > best.Free {
			best = u
			found = true
		}
	}
	if !found {
		return DiskUsage{}, fmt.Errorf("no valid candidate paths")
	}
	return best, nil
}

// GetAllUsages returns DiskUsage for all mounted partitions (similar to `df -h`).
func GetAllUsages() ([]DiskUsage, error) {
	lg, _ := logger.Get()
	parts, err := gopsutildisk.Partitions(false)
	if err != nil {
		lg.Error("Failed to list partitions", logger.Error(err))
		return nil, fmt.Errorf("failed to list partitions: %w", err)
	}

	var out []DiskUsage
	for _, p := range parts {
		// p.Mountpoint may be empty or not accessible; skip on error
		usage, err := defaultDiskProvider.Usage(p.Mountpoint)
		if err != nil {
			// skip this partition but log debug
			lg.Debug("Skipping partition, cannot get usage", logger.String("mount", p.Mountpoint), logger.Error(err))
			continue
		}

		var free int64
		if usage.Free > uint64(math.MaxInt64) {
			free = math.MaxInt64
		} else {
			free = int64(usage.Free)
		}
		var total int64
		if usage.Total > uint64(math.MaxInt64) {
			total = math.MaxInt64
		} else {
			total = int64(usage.Total)
		}
		used := total - free

		du := DiskUsage{
			Path:        p.Mountpoint,
			Mountpoint:  p.Mountpoint,
			Fstype:      usage.Fstype,
			Total:       total,
			Free:        free,
			Used:        used,
			UsedPercent: usage.UsedPercent,
		}
		out = append(out, du)
	}
	return out, nil
}

// mbToBytes converts megabytes to bytes safely.
// mbToBytes mengonversi MB ke byte dengan pengecekan overflow.
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
