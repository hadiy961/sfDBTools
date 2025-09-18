package fs

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"sfDBTools/internal/logger"
)

// ChecksumOperations provides file checksum calculation and comparison operations
type ChecksumOperations interface {
	CalculateMD5(filePath string) (string, error)
	CalculateSHA256(filePath string) (string, error)
	CompareFiles(file1, file2 string) (bool, error)
	VerifyChecksum(filePath, expectedChecksum, algorithm string) (bool, error)
}

type checksumOperations struct {
	logger *logger.Logger
}

func newChecksumOperations(logger *logger.Logger) ChecksumOperations {
	return &checksumOperations{
		logger: logger,
	}
}

// CalculateMD5 calculates MD5 checksum of a file
func (c *checksumOperations) CalculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate MD5 for %s: %w", filePath, err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// CalculateSHA256 calculates SHA256 checksum of a file
func (c *checksumOperations) CalculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate SHA256 for %s: %w", filePath, err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// CompareFiles compares two files by their MD5 checksums
func (c *checksumOperations) CompareFiles(file1, file2 string) (bool, error) {
	checksum1, err := c.CalculateMD5(file1)
	if err != nil {
		return false, fmt.Errorf("failed to calculate checksum for %s: %w", file1, err)
	}

	checksum2, err := c.CalculateMD5(file2)
	if err != nil {
		return false, fmt.Errorf("failed to calculate checksum for %s: %w", file2, err)
	}

	return checksum1 == checksum2, nil
}

// VerifyChecksum verifies a file against an expected checksum using specified algorithm
func (c *checksumOperations) VerifyChecksum(filePath, expectedChecksum, algorithm string) (bool, error) {
	var actualChecksum string
	var err error

	switch algorithm {
	case "md5":
		actualChecksum, err = c.CalculateMD5(filePath)
	case "sha256":
		actualChecksum, err = c.CalculateSHA256(filePath)
	default:
		return false, fmt.Errorf("unsupported checksum algorithm: %s", algorithm)
	}

	if err != nil {
		return false, err
	}

	return actualChecksum == expectedChecksum, nil
}
