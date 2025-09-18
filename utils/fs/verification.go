package fs

import (
	"fmt"
	"os"
	"path/filepath"

	"sfDBTools/internal/logger"
)

// VerificationResult represents the result of a verification operation
type VerificationResult struct {
	File    string
	Status  string // PASSED, FAILED, SKIPPED
	Error   error
	Details string
}

// FileVerificationOperations provides file verification and comparison operations
type FileVerificationOperations interface {
	// Existence checks
	FileExists(path string) bool
	FilesExist(paths []string) map[string]bool

	// Size operations
	CompareSizes(file1, file2 string) (bool, error)
	GetFileSize(path string) (int64, error)

	// Batch verification
	VerifyFiles(sourceDir, destDir string, filenames []string) []VerificationResult
	VerifyFileSizes(sourceDir, destDir string, filenames []string) error

	// Content verification
	VerifyFileIntegrity(sourcePath, destPath string, maxSizeForChecksum int64) (*VerificationResult, error)
}

type fileVerificationOperations struct {
	logger   *logger.Logger
	checksum ChecksumOperations
}

func newFileVerificationOperations(logger *logger.Logger) FileVerificationOperations {
	return &fileVerificationOperations{
		logger:   logger,
		checksum: newChecksumOperations(logger),
	}
}

// FileExists checks if a file (not directory) exists
func (f *fileVerificationOperations) FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// FilesExist checks multiple files for existence
func (f *fileVerificationOperations) FilesExist(paths []string) map[string]bool {
	results := make(map[string]bool)
	for _, path := range paths {
		results[path] = f.FileExists(path)
	}
	return results
}

// CompareSizes compares the size of two files
func (f *fileVerificationOperations) CompareSizes(file1, file2 string) (bool, error) {
	info1, err := os.Stat(file1)
	if err != nil {
		return false, fmt.Errorf("failed to stat %s: %w", file1, err)
	}

	info2, err := os.Stat(file2)
	if err != nil {
		return false, fmt.Errorf("failed to stat %s: %w", file2, err)
	}

	// Both must be files (not directories)
	if info1.IsDir() || info2.IsDir() {
		return false, fmt.Errorf("cannot compare sizes: one or both paths are directories")
	}

	return info1.Size() == info2.Size(), nil
}

// GetFileSize returns the size of a file in bytes
func (f *fileVerificationOperations) GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("failed to stat %s: %w", path, err)
	}

	if info.IsDir() {
		return 0, fmt.Errorf("path is a directory, not a file: %s", path)
	}

	return info.Size(), nil
}

// VerifyFiles verifies multiple files between source and destination directories
func (f *fileVerificationOperations) VerifyFiles(sourceDir, destDir string, filenames []string) []VerificationResult {
	var results []VerificationResult

	for _, filename := range filenames {
		sourcePath := filepath.Join(sourceDir, filename)
		destPath := filepath.Join(destDir, filename)

		// Check if file exists in source
		if !f.FileExists(sourcePath) && !f.dirExists(sourcePath) {
			results = append(results, VerificationResult{
				File:    filename,
				Status:  "SKIPPED",
				Details: "File does not exist in source",
			})
			continue
		}

		// Check if file exists in destination
		if !f.FileExists(destPath) && !f.dirExists(destPath) {
			results = append(results, VerificationResult{
				File:   filename,
				Status: "FAILED",
				Error:  fmt.Errorf("file missing in destination: %s", filename),
			})
			continue
		}

		results = append(results, VerificationResult{
			File:   filename,
			Status: "PASSED",
		})
	}

	return results
}

// VerifyFileSizes compares file sizes between source and destination directories
func (f *fileVerificationOperations) VerifyFileSizes(sourceDir, destDir string, filenames []string) error {
	for _, filename := range filenames {
		sourcePath := filepath.Join(sourceDir, filename)
		destPath := filepath.Join(destDir, filename)

		// Skip if source file doesn't exist
		if !f.FileExists(sourcePath) {
			continue
		}

		// Skip if destination file doesn't exist
		if !f.FileExists(destPath) {
			continue
		}

		sourceInfo, err := os.Stat(sourcePath)
		if err != nil {
			continue
		}

		destInfo, err := os.Stat(destPath)
		if err != nil {
			continue
		}

		// For files, check exact size match
		if !sourceInfo.IsDir() && !destInfo.IsDir() {
			if sourceInfo.Size() != destInfo.Size() {
				return fmt.Errorf("file size mismatch for %s: source=%d, dest=%d",
					filename, sourceInfo.Size(), destInfo.Size())
			}
		}
	}

	return nil
}

// VerifyFileIntegrity performs comprehensive verification of a file including checksum
func (f *fileVerificationOperations) VerifyFileIntegrity(sourcePath, destPath string, maxSizeForChecksum int64) (*VerificationResult, error) {
	result := &VerificationResult{
		File: filepath.Base(sourcePath),
	}

	// Check if both files exist
	if !f.FileExists(sourcePath) {
		result.Status = "SKIPPED"
		result.Details = "Source file does not exist"
		return result, nil
	}

	if !f.FileExists(destPath) {
		result.Status = "FAILED"
		result.Error = fmt.Errorf("destination file does not exist")
		return result, result.Error
	}

	// Compare sizes
	sizeMatch, err := f.CompareSizes(sourcePath, destPath)
	if err != nil {
		result.Status = "FAILED"
		result.Error = fmt.Errorf("failed to compare sizes: %w", err)
		return result, result.Error
	}

	if !sizeMatch {
		result.Status = "FAILED"
		result.Error = fmt.Errorf("file sizes do not match")
		return result, result.Error
	}

	// If file is small enough, verify checksum
	sourceInfo, err := os.Stat(sourcePath)
	if err == nil && !sourceInfo.IsDir() && sourceInfo.Size() <= maxSizeForChecksum {
		match, err := f.checksum.CompareFiles(sourcePath, destPath)
		if err != nil {
			f.logger.Warn("Failed to compare checksums",
				logger.String("source", sourcePath),
				logger.String("dest", destPath),
				logger.Error(err))
			// Don't fail verification on checksum errors, just log warning
		} else if !match {
			result.Status = "FAILED"
			result.Error = fmt.Errorf("checksum mismatch")
			return result, result.Error
		}
	}

	result.Status = "PASSED"
	return result, nil
}

// dirExists checks if a directory exists
func (f *fileVerificationOperations) dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
