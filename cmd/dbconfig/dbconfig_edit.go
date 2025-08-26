package dbconfig

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
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

var EditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit encrypted database configuration",
	Long: `Edit encrypted database configuration file.
If no file is specified, it will list all available encrypted config files
and allow you to choose one. You can modify name, host, port, user, and password.
If the name changes, the file will be renamed accordingly.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Clear screen and show header
		terminal.ClearAndShowHeader("✏️ Edit Database Configuration")

		if err := editConfigCommandEnhanced(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("Failed to edit config", logger.Error(err))
			terminal.PrintError(fmt.Sprintf("Edit failed: %v", err))
			terminal.WaitForEnterWithMessage("Press Enter to continue...")
			os.Exit(1)
		}
	},
}

var editFilePath string

func init() {
	EditCmd.Flags().StringVarP(&editFilePath, "file", "f", "", "Specific encrypted config file to edit")
}

// editConfigCommandEnhanced is the enhanced version with terminal utilities
func editConfigCommandEnhanced(cmd *cobra.Command) error {
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

	// Brief delay for preparation
	time.Sleep(300 * time.Millisecond)

	// If specific file is provided via flag
	if editFilePath != "" {
		spinner.Stop()
		fmt.Println()
		terminal.PrintInfo(fmt.Sprintf("Editing file: %s", editFilePath))
		return editSpecificConfigEnhanced(editFilePath)
	}

	spinner.Stop()
	fmt.Println()

	// List all encrypted config files and let user choose
	terminal.PrintSubHeader("📂 Select Configuration File")
	selectedFile, err := common.SelectConfigFileInteractive()
	if err != nil {
		return err
	}
	return editSpecificConfigEnhanced(selectedFile)
}

// editSpecificConfigEnhanced edits specific config with enhanced display
func editSpecificConfigEnhanced(filePath string) error {
	// Show loading with spinner
	terminal.PrintSubHeader("🔧 Loading Configuration")
	spinner := terminal.NewProgressSpinner("Loading configuration for editing...")
	spinner.Start()

	// Brief delay to show loading
	time.Sleep(500 * time.Millisecond)

	// Stop spinner before interactive editing
	spinner.Stop()
	fmt.Println()

	err := editSpecificConfig(filePath)

	if err != nil {
		terminal.PrintError(fmt.Sprintf("Edit failed: %v", err))
		return err
	}

	terminal.PrintSuccess("✅ Configuration updated successfully!")
	terminal.WaitForEnterWithMessage("Press Enter to continue...")
	return nil
}

func editSpecificConfig(filePath string) error {
	// Validate config file
	if err := common.ValidateConfigFile(filePath); err != nil {
		return err
	}

	// Get encryption password (use environment variable if available)
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
	if newName != currentName {
		fmt.Printf("   Name: %s → %s\n", currentName, newName)
	}
	if newHost != dbConfig.Host {
		fmt.Printf("   Host: %s → %s\n", dbConfig.Host, newHost)
	}
	if newPort != dbConfig.Port {
		fmt.Printf("   Port: %d → %d\n", dbConfig.Port, newPort)
	}
	if newUser != dbConfig.User {
		fmt.Printf("   User: %s → %s\n", dbConfig.User, newUser)
	}
	if newPassword != dbConfig.Password {
		fmt.Printf("   Password: %s → %s\n", strings.Repeat("*", len(dbConfig.Password)), strings.Repeat("*", len(newPassword)))
	}

	// Check if any changes were made
	hasChanges := newName != currentName || newHost != dbConfig.Host || newPort != dbConfig.Port || newUser != dbConfig.User || newPassword != dbConfig.Password

	if !hasChanges {
		fmt.Println("   No changes made.")
		fmt.Println("✅ Configuration editing completed (no changes).")
		return nil
	}

	// Confirm changes
	fmt.Print("\nSave these changes? [Y/n]: ")
	confirmInput, _ := reader.ReadString('\n')
	confirm := strings.ToLower(strings.TrimSpace(confirmInput))
	if confirm != "" && confirm != "y" && confirm != "yes" {
		fmt.Println("❌ Changes cancelled.")
		return nil
	}

	// Save the updated configuration
	return saveUpdatedConfig(filePath, currentName, newName, updatedConfig, encryptionPassword)
}

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

		fmt.Printf("✅ Configuration saved to: %s\n", newFilePath)
		fmt.Printf("🗑️  Old file removed: %s\n", originalPath)
	} else {
		// Same name, just overwrite
		if err := os.WriteFile(originalPath, encryptedData, 0600); err != nil {
			return fmt.Errorf("failed to write encrypted config file: %w", err)
		}

		fmt.Printf("✅ Configuration updated: %s\n", originalPath)
	}

	fmt.Println("🔐 Configuration encrypted and saved successfully!")
	return nil
}
