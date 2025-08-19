package common

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sfDBTools/internal/config"
	"sfDBTools/utils/crypto"
)

// findEncryptedConfigFiles finds all .cnf.enc files in the specified directory
func FindEncryptedConfigFiles(dir string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return files, nil // Return empty slice if directory doesn't exist
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".cnf.enc") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files, nil
}

// SelectConfigFileInteractive shows a list of encrypted config files and lets user choose one
func SelectConfigFileInteractive() (string, error) {
	configDir := config.GetDatabaseConfigDirectory()
	encFiles, err := FindEncryptedConfigFiles(configDir)
	if err != nil {
		return "", fmt.Errorf("failed to find encrypted config files: %w", err)
	}

	if len(encFiles) == 0 {
		fmt.Println("❌ No encrypted configuration files found.")
		fmt.Println("   Use 'config generate' to create one.")
		return "", fmt.Errorf("no encrypted configuration files found")
	}

	// Display available files
	fmt.Println("📁 Available Encrypted Configuration Files:")
	fmt.Println("==========================================")
	for i, file := range encFiles {
		// Extract just the filename for display
		filename := filepath.Base(file)
		fmt.Printf("   %d. %s\n", i+1, filename)
	}

	// Let user choose
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nSelect configuration file (1-%d): ", len(encFiles))
	choice, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read selection: %w", err)
	}

	choice = strings.TrimSpace(choice)
	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(encFiles) {
		return "", fmt.Errorf("invalid selection: %s", choice)
	}

	return encFiles[index-1], nil
}

// LoadEncryptedConfigFromFile loads and decrypts config from a specific file
func LoadEncryptedConfigFromFile(filePath, encryptionPassword string) (*config.EncryptedDatabaseConfig, error) {
	// Load main config to get app settings
	cfg, err := config.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to load main configuration: %w", err)
	}

	// Use the function from internal/config package
	return config.LoadEncryptedDatabaseConfigFromFile(filePath, cfg, encryptionPassword)
}

// GetDatabaseConfigFromEncrypted gets database configuration from encrypted file
func GetDatabaseConfigFromEncrypted(configFilePath string) (host string, port int, user, password string, err error) {
	// Get encryption password
	encryptionPassword, err := crypto.GetEncryptionPassword("Enter encryption password: ")
	if err != nil {
		return "", 0, "", "", fmt.Errorf("failed to get encryption password: %w", err)
	}

	// Load and decrypt the configuration
	dbConfig, err := LoadEncryptedConfigFromFile(configFilePath, encryptionPassword)
	if err != nil {
		return "", 0, "", "", fmt.Errorf("failed to load encrypted config: %w", err)
	}

	return dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, nil
}

// ValidateConfigFile checks if the file path is valid for encrypted config
func ValidateConfigFile(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file does not exist: %s", filePath)
	}

	// Check if it's an encrypted config file
	if !strings.HasSuffix(filePath, ".cnf.enc") {
		return fmt.Errorf("file must be an encrypted config file (.cnf.enc)")
	}

	return nil
}

// HandleDecryptionError provides informative error messages for decryption failures
func HandleDecryptionError(err error, filePath string) error {
	if strings.Contains(err.Error(), "message authentication failed") {
		fmt.Println("❌ Decryption Failed: Incorrect Encryption Password")
		fmt.Println("   The password you entered does not match the one used to encrypt this configuration.")
		fmt.Println("   Please verify your encryption password and try again.")
		return fmt.Errorf("incorrect encryption password")
	}
	if strings.Contains(err.Error(), "failed to read encrypted config file") {
		fmt.Printf("❌ File Error: Cannot read the configuration file\n")
		fmt.Printf("   File: %s\n", filePath)
		fmt.Printf("   Please check if the file exists and you have read permissions.\n")
		return fmt.Errorf("cannot read configuration file")
	}
	if strings.Contains(err.Error(), "failed to parse decrypted database configuration") {
		fmt.Printf("❌ File Corruption: The configuration file appears to be corrupted\n")
		fmt.Printf("   File: %s\n", filePath)
		fmt.Printf("   The file may have been modified or corrupted after encryption.\n")
		return fmt.Errorf("corrupted configuration file")
	}
	fmt.Printf("❌ Decryption Error: %v\n", err)
	fmt.Printf("   File: %s\n", filePath)
	fmt.Printf("   Please check your encryption password and ensure the file is not corrupted.\n")
	return fmt.Errorf("decryption failed")
}
