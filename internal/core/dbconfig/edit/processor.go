package edit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/crypto"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"
)

// ProcessEdit handles the core edit operation logic
func ProcessEdit(cfg *dbconfig.Config) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting database configuration editing")

	// Show edit info
	terminal.PrintSubHeader("✏️ Configuration Editor")
	terminal.PrintInfo("This will allow you to modify database connection details.")
	terminal.PrintInfo("You can change host, port, username, password, and configuration name.")
	fmt.Println()

	// Show spinner while preparing
	spinner := terminal.NewProgressSpinner("Preparing configuration editor...")
	spinner.Start()
	time.Sleep(300 * time.Millisecond)
	spinner.Stop()
	fmt.Println()

	return editSpecificConfig(cfg.FilePath)
}

// editSpecificConfig edits a specific configuration file
func editSpecificConfig(filePath string) error {
	// Show loading with spinner
	terminal.PrintSubHeader("🔧 Loading Configuration")
	spinner := terminal.NewProgressSpinner("Loading configuration for editing...")
	spinner.Start()
	time.Sleep(500 * time.Millisecond)
	spinner.Stop()
	fmt.Println()

	// Validate config file
	if err := common.ValidateConfigFile(filePath); err != nil {
		return err
	}

	// Get encryption password
	encryptionPassword, err := crypto.GetEncryptionPassword("Enter encryption password to decrypt config: ")
	if err != nil {
		return fmt.Errorf("failed to get encryption password: %w", err)
	}

	// Load and decrypt the configuration
	dbConfig, err := common.LoadEncryptedConfigFromFile(filePath, encryptionPassword)
	if err != nil {
		return common.HandleDecryptionError(err, filePath)
	}

	// Display current configuration
	fmt.Println("🔧 Current Database Configuration:")
	fmt.Println("==================================")
	fmt.Printf("📁 Source: %s\n", filePath)
	fmt.Printf("   Host: %s\n", dbConfig.Host)
	fmt.Printf("   Port: %d\n", dbConfig.Port)
	fmt.Printf("   User: %s\n", dbConfig.User)
	fmt.Printf("   Password: %s\n", dbConfig.Password)

	// Get current name from file path
	currentName := strings.TrimSuffix(strings.TrimSuffix(filePath, ".cnf.enc"), "config/")
	if strings.Contains(currentName, "/") {
		parts := strings.Split(currentName, "/")
		currentName = parts[len(parts)-1]
	}

	// Prompt for new values
	updatedConfig, newName, hasChanges, err := promptForUpdates(dbConfig, currentName)
	if err != nil {
		return err
	}

	// Check if any changes were made
	if !hasChanges {
		fmt.Println("   No changes made.")
		fmt.Println("✅ Configuration editing completed (no changes).")
		return nil
	}

	// Confirm changes
	if !dbconfig.ConfirmSaveChanges() {
		terminal.PrintWarning("❌ Changes cancelled.")
		return nil
	}

	// Save the updated configuration
	return saveUpdatedConfig(filePath, currentName, newName, updatedConfig, encryptionPassword)
}

// promptForUpdates prompts user for configuration updates
func promptForUpdates(dbConfig *config.EncryptedDatabaseConfig, currentName string) (*config.EncryptedDatabaseConfig, string, bool, error) {
	fmt.Println("\n✏️  Edit Configuration:")
	fmt.Println("========================")
	fmt.Println("Press Enter to keep current value, or type new value:")

	reader := bufio.NewReader(os.Stdin)

	// Edit configuration name
	fmt.Printf("Configuration name [%s]: ", currentName)
	newNameInput, _ := reader.ReadString('\n')
	newName := strings.TrimSpace(newNameInput)
	if newName == "" {
		newName = currentName
	}

	// Edit host
	fmt.Printf("Host [%s]: ", dbConfig.Host)
	newHostInput, _ := reader.ReadString('\n')
	newHost := strings.TrimSpace(newHostInput)
	if newHost == "" {
		newHost = dbConfig.Host
	}

	// Edit port
	fmt.Printf("Port [%d]: ", dbConfig.Port)
	newPortInput, _ := reader.ReadString('\n')
	newPortStr := strings.TrimSpace(newPortInput)
	newPort := dbConfig.Port
	if newPortStr != "" {
		if port, err := strconv.Atoi(newPortStr); err == nil {
			if port >= 1 && port <= 65535 {
				newPort = port
			} else {
				fmt.Printf("❌ Invalid port number: %d. Using current value: %d\n", port, dbConfig.Port)
			}
		} else {
			fmt.Printf("❌ Invalid port format. Using current value: %d\n", dbConfig.Port)
		}
	}

	// Edit user
	fmt.Printf("User [%s]: ", dbConfig.User)
	newUserInput, _ := reader.ReadString('\n')
	newUser := strings.TrimSpace(newUserInput)
	if newUser == "" {
		newUser = dbConfig.User
	}

	// Edit password
	fmt.Print("Password [current password]: ")
	newPasswordInput, _ := reader.ReadString('\n')
	newPassword := strings.TrimSpace(newPasswordInput)
	if newPassword == "" {
		newPassword = dbConfig.Password
	}

	// Create updated configuration
	updatedConfig := &config.EncryptedDatabaseConfig{
		Host:     newHost,
		Port:     newPort,
		User:     newUser,
		Password: newPassword,
	}

	// Display changes summary
	fmt.Println("\n📋 Changes Summary:")
	fmt.Println("===================")
	hasChanges := false
	if newName != currentName {
		fmt.Printf("   Name: %s → %s\n", currentName, newName)
		hasChanges = true
	}
	if newHost != dbConfig.Host {
		fmt.Printf("   Host: %s → %s\n", dbConfig.Host, newHost)
		hasChanges = true
	}
	if newPort != dbConfig.Port {
		fmt.Printf("   Port: %d → %d\n", dbConfig.Port, newPort)
		hasChanges = true
	}
	if newUser != dbConfig.User {
		fmt.Printf("   User: %s → %s\n", dbConfig.User, newUser)
		hasChanges = true
	}
	if newPassword != dbConfig.Password {
		fmt.Printf("   Password: %s → %s\n", strings.Repeat("*", len(dbConfig.Password)), strings.Repeat("*", len(newPassword)))
		hasChanges = true
	}

	return updatedConfig, newName, hasChanges, nil
}

// saveUpdatedConfig saves the updated configuration to file
func saveUpdatedConfig(originalPath, currentName, newName string, dbConfig *config.EncryptedDatabaseConfig, encryptionPassword string) error {
	// Generate encryption key from user password only
	key, err := crypto.DeriveKeyWithPassword(encryptionPassword)
	if err != nil {
		return fmt.Errorf("failed to derive encryption key: %w", err)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal database config to JSON: %w", err)
	}

	// Encrypt the configuration
	encryptedData, err := crypto.EncryptData(jsonData, key, crypto.AES_GCM)
	if err != nil {
		return fmt.Errorf("failed to encrypt database configuration: %w", err)
	}

	// Determine new file path
	configDir, err := config.GetDatabaseConfigDirectory()
	if err != nil {
		return fmt.Errorf("failed to get database config directory: %w", err)
	}

	newFileName := newName + ".cnf.enc"
	newFilePath := filepath.Join(configDir, newFileName)

	// If name changed, we need to handle file rename
	if newName != currentName {
		// Check if new file already exists
		if _, err := os.Stat(newFilePath); err == nil {
			return fmt.Errorf("configuration file with name '%s' already exists", newName)
		}

		// Write to new file
		if err := os.WriteFile(newFilePath, encryptedData, 0600); err != nil {
			return fmt.Errorf("failed to write new encrypted config file: %w", err)
		}

		// Remove old file
		if err := os.Remove(originalPath); err != nil {
			// If we can't remove old file, warn but don't fail
			fmt.Printf("⚠️  Warning: Could not remove old file %s: %v\n", originalPath, err)
		}

		terminal.PrintSuccess(fmt.Sprintf("✅ Configuration saved to: %s", newFilePath))
		terminal.PrintInfo(fmt.Sprintf("🗑️  Old file removed: %s", originalPath))
	} else {
		// Same name, just overwrite
		if err := os.WriteFile(originalPath, encryptedData, 0600); err != nil {
			return fmt.Errorf("failed to write encrypted config file: %w", err)
		}

		terminal.PrintSuccess(fmt.Sprintf("✅ Configuration updated: %s", originalPath))
	}

	terminal.PrintSuccess("🔐 Configuration encrypted and saved successfully!")
	return nil
}
