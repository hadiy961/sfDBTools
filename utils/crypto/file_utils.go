package crypto

import (
	"fmt"
	"io"
	"os"
)

// DecryptFileStreaming decrypts a file using streaming AES-GCM decryption
func DecryptFileStreaming(sourceFilePath, destFilePath string, key []byte) error {
	// Open the encrypted source file
	sourceFile, err := os.Open(sourceFilePath)
	if err != nil {
		return fmt.Errorf("failed to open encrypted file: %w", err)
	}
	defer sourceFile.Close()

	// Create the destination file
	destFile, err := os.OpenFile(destFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Create decrypting reader
	decReader, err := NewGCMDecryptingReader(sourceFile, key)
	if err != nil {
		return fmt.Errorf("failed to create decrypting reader: %w", err)
	}

	// Copy decrypted data to destination
	if _, err := io.Copy(destFile, decReader); err != nil {
		return fmt.Errorf("failed to copy decrypted data: %w", err)
	}

	return nil
}

// ValidateEncryptedFileStreaming validates an encrypted file by attempting to read its structure
func ValidateEncryptedFileStreaming(filePath string, key []byte) error {
	// Open the encrypted file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open encrypted file for validation: %w", err)
	}
	defer file.Close()

	// Get file info for size
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats: %w", err)
	}

	// Check if file is empty
	if fileInfo.Size() == 0 {
		return fmt.Errorf("encrypted file is empty")
	}

	// Check minimum size (nonce + tag)
	minSize := int64(GCMNonceSize + GCMTagSize)
	if fileInfo.Size() < minSize {
		return fmt.Errorf("encrypted file too small for GCM format: %d bytes (minimum %d)", fileInfo.Size(), minSize)
	}

	// Try to create decrypting reader (validates nonce reading)
	decReader, err := NewGCMDecryptingReader(file, key)
	if err != nil {
		return fmt.Errorf("failed to create decrypting reader (invalid nonce?): %w", err)
	}

	// Try to read a small amount of data to validate the structure
	testBuf := make([]byte, 1024)
	_, err = decReader.Read(testBuf)
	if err != nil && err.Error() != "EOF" {
		return fmt.Errorf("validation failed - file cannot be decrypted: %w", err)
	}

	return nil
}
