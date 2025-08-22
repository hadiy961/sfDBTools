package config

import (
	"fmt"
	"os"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/crypto"

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
		if err := showConfig(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("Failed to show config", logger.Error(err))
			os.Exit(1)
		}
	},
}

var configFile string

func init() {
	ShowCmd.Flags().StringVarP(&configFile, "file", "f", "", "Specific encrypted config file to show")
}

func showConfig(cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Showing database configuration")

	// If specific file is provided via flag
	if configFile != "" {
		return showSpecificConfig(configFile)
	}

	// List all encrypted config files and let user choose
	return showConfigWithSelection()
}

// showSpecificConfig shows configuration from a specific file
func showSpecificConfig(filePath string) error {
	// Validate config file
	if err := common.ValidateConfigFile(filePath); err != nil {
		return err
	}

	// Get encryption password (always prompt, don't use environment variable)
	encryptionPassword, err := crypto.PromptEncryptionPassword("Enter encryption password to decrypt config: ")
	if err != nil {
		return fmt.Errorf("failed to get encryption password: %w", err)
	}

	// Load and decrypt the configuration
	dbConfig, err := common.LoadEncryptedConfigFromFile(filePath, encryptionPassword)
	if err != nil {
		return common.HandleDecryptionError(err, filePath)
	}

	// Display configuration
	displayConfiguration(filePath, dbConfig)
	return nil
}

// showConfigWithSelection lists all encrypted config files and lets user choose
func showConfigWithSelection() error {
	selectedFile, err := common.SelectConfigFileInteractive()
	if err != nil {
		return err
	}
	return showSpecificConfig(selectedFile)
}

// displayConfiguration displays the decrypted configuration
func displayConfiguration(filePath string, dbConfig *config.EncryptedDatabaseConfig) {
	fmt.Println("🔧 Database Configuration:")
	fmt.Println("==========================")
	fmt.Printf("📁 Source: %s\n", filePath)
	fmt.Printf("   Host: %s\n", dbConfig.Host)
	fmt.Printf("   Port: %d\n", dbConfig.Port)
	fmt.Printf("   User: %s\n", dbConfig.User)
	fmt.Printf("   Password: %s\n", dbConfig.Password)
}

func init() {
	// This will be called from config_cmd.go
}
