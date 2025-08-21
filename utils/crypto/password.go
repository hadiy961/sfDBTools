package crypto

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

const (
	// Environment variable for encryption password
	ENV_ENCRYPTION_PASSWORD = "SFDB_ENCRYPTION_PASSWORD"

	// Environment variables for database credentials
	ENV_DB_HOST     = "SFDB_HOST"
	ENV_DB_PORT     = "SFDB_PORT"
	ENV_DB_USER     = "SFDB_USER"
	ENV_DB_PASSWORD = "SFDB_PASSWORD"
)

// GetEncryptionPassword gets encryption password from environment variable or user input
func GetEncryptionPassword(promptMessage string) (string, error) {
	// First, try to get password from environment variable
	if password := os.Getenv(ENV_ENCRYPTION_PASSWORD); password != "" {
		return password, nil
	}

	// If not found in environment, prompt user for password
	fmt.Print(promptMessage)
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // New line after password input

	password := strings.TrimSpace(string(passwordBytes))
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	return password, nil
}

// ConfirmEncryptionPassword prompts user to confirm password by entering it twice
func ConfirmEncryptionPassword(promptMessage string) (string, error) {
	// First, try to get password from environment variable
	if password := os.Getenv(ENV_ENCRYPTION_PASSWORD); password != "" {
		return password, nil
	}

	// Prompt for password
	password1, err := GetEncryptionPassword(promptMessage)
	if err != nil {
		return "", err
	}

	// Prompt for confirmation
	password2, err := GetEncryptionPassword("Confirm encryption password: ")
	if err != nil {
		return "", err
	}

	// Check if passwords match
	if password1 != password2 {
		return "", fmt.Errorf("passwords do not match")
	}

	return password1, nil
}

// DeriveKeyWithPassword derives an encryption key using only user password
func DeriveKeyWithPassword(userPassword string) ([]byte, error) {
	// Use only the user password for key derivation
	password := []byte(userPassword)

	// Generate a standard salt for consistency across runs
	salt := []byte("sfdb_encryption_salt_v3")

	// Derive encryption key
	key := DeriveKeyFromPassword(
		password,
		salt,
		32, // AES-256 key length
		DefaultIterations,
	)

	return key, nil
}

// ValidatePassword checks if a password can successfully decrypt a test piece of data
func ValidatePassword(encryptedData []byte, userPassword string) error {
	key, err := DeriveKeyWithPassword(userPassword)
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	_, err = DecryptData(encryptedData, key, AES_GCM)
	if err != nil {
		return fmt.Errorf("invalid password or corrupted data: %w", err)
	}

	return nil
}

// GetDatabasePassword gets database password from environment variable or user input
func GetDatabasePassword(promptMessage string) (string, error) {
	// First, try to get password from environment variable
	if password := os.Getenv(ENV_DB_PASSWORD); password != "" {
		return password, nil
	}

	// If not found in environment, prompt user for input
	fmt.Print(promptMessage)
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // New line after password input

	password := strings.TrimSpace(string(passwordBytes))
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	return password, nil
}

// PromptEncryptionPassword prompts user for encryption password without checking environment variable
// Used specifically for show command where we always want user to enter password
func PromptEncryptionPassword(promptMessage string) (string, error) {
	fmt.Print(promptMessage)
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // New line after password input

	password := strings.TrimSpace(string(passwordBytes))
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	return password, nil
}
