package crypto

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestFileEncryptionDecryption(t *testing.T) {
	// Generate a test key
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create temp directory
	tempDir := t.TempDir()

	// Create test file with data
	originalFile := filepath.Join(tempDir, "original.txt")
	testData := []byte("This is a test file for encryption/decryption testing.\nIt contains multiple lines.\nAnd some special characters: !@#$%^&*()")

	if err := os.WriteFile(originalFile, testData, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Encrypt the file using streaming encryption
	encryptedFile := filepath.Join(tempDir, "encrypted.txt.enc")

	// Simulate the streaming encryption process
	inFile, err := os.Open(originalFile)
	if err != nil {
		t.Fatalf("Failed to open original file: %v", err)
	}
	defer inFile.Close()

	outFile, err := os.Create(encryptedFile)
	if err != nil {
		t.Fatalf("Failed to create encrypted file: %v", err)
	}

	encWriter, err := NewGCMEncryptingWriter(outFile, key)
	if err != nil {
		t.Fatalf("Failed to create encrypting writer: %v", err)
	}

	// Copy data through the encrypting writer
	buf := make([]byte, 1024)
	for {
		n, err := inFile.Read(buf)
		if n > 0 {
			if _, writeErr := encWriter.Write(buf[:n]); writeErr != nil {
				t.Fatalf("Failed to write encrypted data: %v", writeErr)
			}
		}
		if err != nil {
			break
		}
	}

	if err := encWriter.Close(); err != nil {
		t.Fatalf("Failed to close encrypting writer: %v", err)
	}
	outFile.Close()

	// Validate the encrypted file
	if err := ValidateEncryptedFileStreaming(encryptedFile, key); err != nil {
		t.Fatalf("Encrypted file validation failed: %v", err)
	}

	// Decrypt the file
	decryptedFile := filepath.Join(tempDir, "decrypted.txt")
	if err := DecryptFileStreaming(encryptedFile, decryptedFile, key); err != nil {
		t.Fatalf("Failed to decrypt file: %v", err)
	}

	// Read decrypted data
	decryptedData, err := os.ReadFile(decryptedFile)
	if err != nil {
		t.Fatalf("Failed to read decrypted file: %v", err)
	}

	// Compare with original
	if string(testData) != string(decryptedData) {
		t.Fatalf("Decrypted data doesn't match original.\nOriginal: %s\nDecrypted: %s",
			string(testData), string(decryptedData))
	}

	t.Logf("Successfully encrypted and decrypted file with %d bytes", len(testData))
}

func TestValidateEncryptedFileStreaming(t *testing.T) {
	// Generate a test key
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	tempDir := t.TempDir()

	// Test with empty file
	emptyFile := filepath.Join(tempDir, "empty.enc")
	if err := os.WriteFile(emptyFile, []byte{}, 0600); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	if err := ValidateEncryptedFileStreaming(emptyFile, key); err == nil {
		t.Fatal("Expected validation to fail for empty file")
	}

	// Test with too small file
	smallFile := filepath.Join(tempDir, "small.enc")
	if err := os.WriteFile(smallFile, []byte{1, 2, 3}, 0600); err != nil {
		t.Fatalf("Failed to create small file: %v", err)
	}

	if err := ValidateEncryptedFileStreaming(smallFile, key); err == nil {
		t.Fatal("Expected validation to fail for too small file")
	}

	// Test with valid encrypted file
	validFile := filepath.Join(tempDir, "valid.enc")
	testData := []byte("test data for validation")

	outFile, err := os.Create(validFile)
	if err != nil {
		t.Fatalf("Failed to create valid file: %v", err)
	}

	encWriter, err := NewGCMEncryptingWriter(outFile, key)
	if err != nil {
		t.Fatalf("Failed to create encrypting writer: %v", err)
	}

	if _, err := encWriter.Write(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	if err := encWriter.Close(); err != nil {
		t.Fatalf("Failed to close encrypting writer: %v", err)
	}
	outFile.Close()

	// This should pass validation
	if err := ValidateEncryptedFileStreaming(validFile, key); err != nil {
		t.Fatalf("Expected validation to pass for valid file: %v", err)
	}

	// Test with wrong key
	wrongKey := make([]byte, 32)
	if _, err := rand.Read(wrongKey); err != nil {
		t.Fatalf("Failed to generate wrong key: %v", err)
	}

	if err := ValidateEncryptedFileStreaming(validFile, wrongKey); err == nil {
		t.Fatal("Expected validation to fail with wrong key")
	}

	t.Log("File validation tests passed")
}
