package backup_utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/file"
	"time"
)

// CalculateChecksum calculates SHA256 checksum of the backup file
func CalculateChecksum(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// initializeBackupResult creates and initializes a backup result structure
func InitializeBackupResult(options BackupOptions) *BackupResult {
	return &BackupResult{
		CompressionUsed: options.Compression,
		Encrypted:       options.Encrypt,
		IncludedData:    options.IncludeData,
	}
}

// setupBackupPaths generates output file paths and creates directories
func SetupBackupPaths(options BackupOptions) (string, string, error) {
	// Create output directory
	if err := file.CreateOutputDirectory(options.OutputDir); err != nil {
		return "", "", err
	}

	// Generate output file path
	outputFile, metaFile := GenerateOutputPaths(options)
	return outputFile, metaFile, nil
}

// finalizeBackupResult calculates final metrics for backup result
func FinalizeBackupResult(result *BackupResult, outputFile string, startTime time.Time, options BackupOptions) error {
	lg, _ := logger.Get()

	// Get output file size
	if stat, err := os.Stat(outputFile); err == nil {
		result.OutputSize = stat.Size()
	}

	// Calculate duration and speed
	result.Duration = time.Since(startTime)
	if result.Duration.Seconds() > 0 {
		result.AverageSpeed = float64(result.OutputSize) / result.Duration.Seconds()
	}

	// Calculate checksum if requested
	if options.CalculateChecksum {
		if checksum, err := common.CalculateChecksum(outputFile); err == nil {
			result.Checksum = checksum
		} else {
			lg.Warn("Failed to calculate checksum", logger.Error(err))
		}
	}

	// Update encryption status
	result.Encrypted = options.Encrypt

	result.Success = true
	return nil
}
