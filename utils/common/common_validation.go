package common

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// isRemoteConnection checks if the connection is remote (not localhost)
func IsRemoteConnection(host string) bool {
	return host != "localhost" && host != "127.0.0.1" && host != "::1" && host != ""
}

// calculateChecksum calculates SHA256 checksum of the backup file
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
