package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// EncryptionMethod represents supported encryption methods
type EncryptionMethod string

// Supported encryption methods
const (
	AES_GCM EncryptionMethod = "AES_GCM" // AES in GCM mode (authenticated encryption)
)

// DefaultIterations is the default number of iterations for PBKDF2
const DefaultIterations = 10000

// DeriveKeyFromPassword derives an encryption key from a password using PBKDF2
func DeriveKeyFromPassword(password, salt []byte, keyLength int, iterations int) []byte {
	if iterations <= 0 {
		iterations = DefaultIterations
	}
	return pbkdf2.Key(password, salt, iterations, keyLength, sha512.New)
}

// DeriveKeyFromConfigValues derives an encryption key using configuration values
func DeriveKeyFromConfigValues(appName, clientCode, version, author string, salt []byte, keyLength int, iterations int) []byte {
	password := []byte(appName + clientCode + version + author)
	return DeriveKeyFromPassword(password, salt, keyLength, iterations)
}

// GenerateRandomBytes generates cryptographically secure random bytes
func GenerateRandomBytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, errors.New("invalid length for random bytes")
	}

	bytes := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

// GenerateRandomSalt generates a random salt for key derivation
func GenerateRandomSalt(length int) ([]byte, error) {
	return GenerateRandomBytes(length)
}

// EncryptData encrypts data using AES-GCM
func EncryptData(data, key []byte, method EncryptionMethod) ([]byte, error) {
	if method != AES_GCM {
		return nil, fmt.Errorf("unsupported encryption method: %s (only AES_GCM is supported)", method)
	}
	return encryptAES_GCM(data, key)
}

// DecryptData decrypts data using AES-GCM
func DecryptData(encryptedData, key []byte, method EncryptionMethod) ([]byte, error) {
	if method != AES_GCM {
		return nil, fmt.Errorf("unsupported encryption method: %s (only AES_GCM is supported)", method)
	}
	return decryptAES_GCM(encryptedData, key)
}

// encryptAES_GCM encrypts data using AES in GCM mode (authenticated encryption)
// The returned data format is: Nonce + Ciphertext + AuthTag
func encryptAES_GCM(plaintext, key []byte) ([]byte, error) {
	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce, err := GenerateRandomBytes(gcm.NonceSize())
	if err != nil {
		return nil, err
	}

	// Seal will append the ciphertext to the nonce
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// decryptAES_GCM decrypts data using AES in GCM mode (authenticated encryption)
// The expected data format is: Nonce + Ciphertext + AuthTag
func decryptAES_GCM(ciphertext, key []byte) ([]byte, error) {
	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Check minimum length
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	// Decrypt and verify
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// Hash computes a hash of the input data using the specified algorithm
func Hash(data []byte, algorithm string) (string, error) {
	switch algorithm {
	case "sha256":
		hash := sha256.Sum256(data)
		return hex.EncodeToString(hash[:]), nil
	case "sha512":
		hash := sha512.Sum512(data)
		return hex.EncodeToString(hash[:]), nil
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}
}

// DeriveKeyFromAppConfig generates an encryption key from application configuration values
func DeriveKeyFromAppConfig(appName, clientCode, version, author string) ([]byte, error) {
	// Generate salt for key derivation (use a fixed salt for consistency across runs)
	salt := []byte("sfdb_encryption_salt_v3")

	// Derive encryption key from provided values
	key := DeriveKeyFromConfigValues(
		appName,
		clientCode,
		version,
		author,
		salt,
		32, // AES-256 key length
		DefaultIterations,
	)

	return key, nil
}

// ValidateEncryptedFile checks if a file is properly encrypted and can be decrypted with the provided key
// For GCM mode, it performs structural validation only to avoid expensive full decryption
// For CBC mode, it validates structure and attempts partial decryption
func ValidateEncryptedFile(filePath string, key []byte, method EncryptionMethod) error {
	if method != AES_GCM {
		return fmt.Errorf("unsupported encryption method for validation: %s (only AES_GCM is supported)", method)
	}

	return ValidateEncryptedFileStreaming(filePath, key)
}
