package disk

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
)

// CheckDiskSpace checks if there's enough disk space available
func CheckDiskSpace(path string, minFreeSpace int64) error {
	lg, _ := logger.Get()

	// Find the first existing parent directory to check disk space
	checkPath := path
	for {
		var stat syscall.Statfs_t
		if err := syscall.Statfs(checkPath, &stat); err != nil {
			// If path doesn't exist, try parent directory
			parentDir := filepath.Dir(checkPath)
			if parentDir == checkPath || parentDir == "." || parentDir == "/" {
				// We've reached the root and still can't find an existing directory
				lg.Error("Failed to get disk space info", logger.Error(err))
				return fmt.Errorf("failed to get disk space info: %w", err)
			}
			checkPath = parentDir
			continue
		}

		// Available space in bytes
		available := int64(stat.Bavail) * int64(stat.Bsize)

		if available < minFreeSpace*1024*1024 { // Convert MB to bytes
			lg.Error("Insufficient disk space",
				logger.String("available", common.FormatSize(available)),
				logger.String("required", common.FormatSize(minFreeSpace*1024*1024)),
				logger.String("checked_path", checkPath))
			return fmt.Errorf("insufficient disk space: available %s, required %s MB",
				common.FormatSize(available), common.FormatSize(minFreeSpace*1024*1024))
		}

		lg.Debug("Disk space check passed",
			logger.String("available", common.FormatSize(available)),
			logger.String("checked_path", checkPath))
		return nil
	}
}

// CreateOutputDirectory creates the output directory if it doesn't exist
func CreateOutputDirectory(outputDir string) error {
	lg, _ := logger.Get()

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		lg.Error("Failed to create output directory",
			logger.String("dir", outputDir),
			logger.Error(err))
		return fmt.Errorf("failed to create output directory '%s': %w", outputDir, err)
	}

	lg.Debug("Output directory created/verified", logger.String("dir", outputDir))
	return nil
}

// ValidateOutputDir memastikan outputDir ada dan bisa ditulis.
// Jika belum ada, akan mencoba membuatnya.
func ValidateOutputDir(outputDir string) error {
	lg, _ := logger.Get()

	if outputDir == "" {
		lg.Error("Output directory is required")
		return fmt.Errorf("output directory is required")
	}
	info, err := os.Stat(outputDir)
	if os.IsNotExist(err) {
		lg.Debug("Output directory does not exist, attempting to create", logger.String("dir", outputDir))
		// Coba buat direktori
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			lg.Error("Failed to create output directory",
				logger.String("dir", outputDir),
				logger.Error(err))
			return fmt.Errorf("failed to create output directory '%s': %w", outputDir, err)
		}
		lg.Info("Output directory created", logger.String("dir", outputDir))
		return nil
	}
	if err != nil {
		lg.Error("Failed to access output directory",
			logger.String("dir", outputDir),
			logger.Error(err))
		return fmt.Errorf("failed to access output directory '%s': %w", outputDir, err)
	}
	if !info.IsDir() {
		lg.Error("Output path is not a directory", logger.String("dir", outputDir))
		return fmt.Errorf("output path '%s' is not a directory", outputDir)
	}
	// Cek permission tulis
	testFile := outputDir + "/.sfbackup_test"
	lg.Debug("Checking write permission for output directory", logger.String("dir", outputDir))
	f, err := os.Create(testFile)
	if err != nil {
		lg.Error("Output directory is not writable",
			logger.String("dir", outputDir),
			logger.Error(err))
		return fmt.Errorf("output directory '%s' is not writable: %w", outputDir, err)
	}
	f.Close()
	os.Remove(testFile)
	lg.Debug("Output directory is writable", logger.String("dir", outputDir))
	return nil
}
