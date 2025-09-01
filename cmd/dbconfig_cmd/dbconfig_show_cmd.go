package dbconfig_cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/crypto"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

var ShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show database configuration from encrypted files",
	Long: `Show database configuration from encrypted files.
If no file is specified, it will list all available encrypted config files
and allow you to choose one. Database password will be displayed in plain text.
You will always be prompted for the encryption password (environment variables are ignored for security).`,
	Run: func(cmd *cobra.Command, args []string) {
		// Clear screen and show header
		terminal.ClearAndShowHeader("Show Database Configuration")

		if err := showConfigEnhanced(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("Failed to show config", logger.Error(err))
			terminal.PrintError(fmt.Sprintf("Failed to show configuration: %v", err))
			terminal.WaitForEnterWithMessage("Press Enter to continue...")
			os.Exit(1)
		}
	},
}

var configFile string

func init() {
	ShowCmd.Flags().StringVarP(&configFile, "file", "f", "", "Specific encrypted config file to show")
}

// showConfigEnhanced is the enhanced version with terminal utilities
func showConfigEnhanced(cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Showing database configuration")

	// Show loading spinner
	spinner := terminal.NewProgressSpinner("Loading configuration files...")
	spinner.Start()

	// If specific file is provided via flag
	if configFile != "" {
		spinner.Stop()
		terminal.PrintSubHeader("📂 Loading Specific Configuration")
		terminal.PrintInfo(fmt.Sprintf("File: %s", configFile))
		return showSpecificConfigEnhanced(configFile)
	}

	spinner.Stop()

	// List all encrypted config files and let user choose
	return showConfigWithSelectionEnhanced()
}

// showConfigWithSelectionEnhanced lists all encrypted config files with enhanced UI
func showConfigWithSelectionEnhanced() error {
	terminal.PrintSubHeader("📂 Select Configuration File")

	selectedFile, err := common.SelectConfigFileInteractive()
	if err != nil {
		return err
	}
	return showSpecificConfigEnhanced(selectedFile)
}

// showSpecificConfigEnhanced shows specific config with enhanced display
func showSpecificConfigEnhanced(filePath string) error {
	// Validate config file
	if err := common.ValidateConfigFile(filePath); err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}

	// Get encryption password
	terminal.PrintSubHeader("🔐 Authentication Required")
	terminal.PrintInfo("Enter your encryption password to decrypt the configuration.")

	encryptionPassword, err := crypto.GetEncryptionPassword("🔑 Encryption password: ")
	if err != nil {
		return fmt.Errorf("failed to get encryption password: %w", err)
	}

	// Load and decrypt configuration
	spinner := terminal.NewProgressSpinner("Decrypting configuration...")
	spinner.Start()

	dbConfig, err := common.LoadEncryptedConfigFromFile(filePath, encryptionPassword)
	spinner.Stop()

	if err != nil {
		return common.HandleDecryptionError(err, filePath)
	}

	// Display configuration with enhanced formatting
	displayConfigurationEnhanced(filePath, dbConfig)
	terminal.WaitForEnterWithMessage("\nPress Enter to continue...")
	return nil
}

// displayConfigurationEnhanced displays configuration with enhanced formatting
func displayConfigurationEnhanced(filePath string, dbConfig *config.EncryptedDatabaseConfig) {
	terminal.ClearAndShowHeader("🔧 Database Configuration Details")

	// Configuration table
	headers := []string{"Property", "Value", "Description"}
	rows := [][]string{
		{"📁 Source File", filepath.Base(filePath), "Configuration file name"},
		{"🌐 Host", dbConfig.Host, "Database server hostname/IP"},
		{"🔌 Port", fmt.Sprintf("%d", dbConfig.Port), "Database server port"},
		{"👤 Username", dbConfig.User, "Database username"},
		{"🔑 Password", maskPassword(dbConfig.Password), "Database password (masked)"},
	}

	terminal.FormatTable(headers, rows)

	// Show full file path
	fmt.Println()
	terminal.PrintSubHeader("📂 File Information")
	terminal.PrintInfo(fmt.Sprintf("Full path: %s", filePath))

	// Get file info
	if info, err := os.Stat(filePath); err == nil {
		terminal.PrintInfo(fmt.Sprintf("File size: %.2f KB", float64(info.Size())/1024))
		terminal.PrintInfo(fmt.Sprintf("Last modified: %s", info.ModTime().Format("2006-01-02 15:04:05")))
	}

	// Security warning
	fmt.Println()
	terminal.PrintWarning("⚠️ Sensitive data displayed - ensure your screen is not being observed")

	// Option to show actual password
	fmt.Println()
	terminal.PrintInfo("To view the actual password, type 'show' (otherwise press Enter):")
	var input string
	fmt.Scanln(&input)

	if strings.ToLower(strings.TrimSpace(input)) == "show" {
		terminal.PrintSubHeader("🔑 Actual Password")
		terminal.PrintColoredText("Password: ", terminal.ColorRed)
		terminal.PrintColoredLine(dbConfig.Password, terminal.ColorBold)
		terminal.PrintWarning("⚠️ Password is now visible on screen!")
	}
}

// maskPassword masks password for display
func maskPassword(password string) string {
	if len(password) <= 2 {
		return "••••"
	}
	return password[:2] + strings.Repeat("•", len(password)-2)
}

func init() {
	// This will be called from config_cmd.go
}
