package show

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/crypto"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"
)

// ProcessShow handles the core show operation logic
func ProcessShow(cfg *dbconfig.Config) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Showing database configuration")

	// Show loading spinner
	spinner := terminal.NewProgressSpinner("Loading configuration files...")
	spinner.Start()
	spinner.Stop()

	// Show specific file or use interactive selection
	return showSpecificConfig(cfg.FilePath)
}

// showSpecificConfig shows specific config with enhanced display
func showSpecificConfig(filePath string) error {
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
	dbconfig.DisplayConfigDetails(filePath, dbConfig)

	// Option to show password
	dbconfig.DisplayPasswordOption(dbConfig.Password)

	terminal.WaitForEnterWithMessage("\nPress Enter to continue...")
	return nil
}
