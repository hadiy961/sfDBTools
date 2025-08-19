package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestGCMStreamingEncryption(t *testing.T) {
	// Generate a test key
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Test data
	testData := []byte("Hello, this is a test of streaming AES-GCM encryption!")

	// Encrypt using streaming writer
	var encryptedBuf bytes.Buffer
	encWriter, err := NewGCMEncryptingWriter(&encryptedBuf, key)
	if err != nil {
		t.Fatalf("Failed to create encrypting writer: %v", err)
	}

	// Write test data
	if _, err := encWriter.Write(testData); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	// Close to finalize encryption
	if err := encWriter.Close(); err != nil {
		t.Fatalf("Failed to close encrypting writer: %v", err)
	}

	// Verify encrypted data is different from original
	encryptedData := encryptedBuf.Bytes()
	if len(encryptedData) <= len(testData) {
		t.Fatalf("Encrypted data should be longer than original (includes nonce + tag)")
	}

	// Decrypt using streaming reader
	decReader, err := NewGCMDecryptingReader(&encryptedBuf, key)
	if err != nil {
		t.Fatalf("Failed to create decrypting reader: %v", err)
	}

	// Read decrypted data
	decryptedData := make([]byte, len(testData)+100) // Extra buffer
	n, err := decReader.Read(decryptedData)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("Failed to read decrypted data: %v", err)
	}

	decryptedData = decryptedData[:n]

	// Verify decrypted data matches original
	if !bytes.Equal(testData, decryptedData) {
		t.Fatalf("Decrypted data doesn't match original.\nOriginal: %s\nDecrypted: %s",
			string(testData), string(decryptedData))
	}

	t.Logf("Successfully encrypted and decrypted %d bytes", len(testData))
}

func TestGCMStreamingEncryptionLargeData(t *testing.T) {
	// Generate a test key
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create larger test data (100KB)
	testData := make([]byte, 100*1024)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	// Encrypt using streaming writer
	var encryptedBuf bytes.Buffer
	encWriter, err := NewGCMEncryptingWriter(&encryptedBuf, key)
	if err != nil {
		t.Fatalf("Failed to create encrypting writer: %v", err)
	}

	// Write test data in chunks
	chunkSize := 8192
	for i := 0; i < len(testData); i += chunkSize {
		end := i + chunkSize
		if end > len(testData) {
			end = len(testData)
		}

		if _, err := encWriter.Write(testData[i:end]); err != nil {
			t.Fatalf("Failed to write chunk %d: %v", i/chunkSize, err)
		}
	}

	// Close to finalize encryption
	if err := encWriter.Close(); err != nil {
		t.Fatalf("Failed to close encrypting writer: %v", err)
	}

	// Decrypt using streaming reader
	decReader, err := NewGCMDecryptingReader(&encryptedBuf, key)
	if err != nil {
		t.Fatalf("Failed to create decrypting reader: %v", err)
	}

	// Read all decrypted data
	var decryptedBuf bytes.Buffer
	buf := make([]byte, 4096)
	for {
		n, err := decReader.Read(buf)
		if n > 0 {
			decryptedBuf.Write(buf[:n])
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Fatalf("Failed to read decrypted data: %v", err)
		}
	}

	decryptedData := decryptedBuf.Bytes()

	// Verify decrypted data matches original
	if !bytes.Equal(testData, decryptedData) {
		t.Fatalf("Decrypted data doesn't match original. Expected %d bytes, got %d bytes",
			len(testData), len(decryptedData))
	}

	t.Logf("Successfully encrypted and decrypted %d bytes in chunks", len(testData))
}
