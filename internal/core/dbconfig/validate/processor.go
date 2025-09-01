package validate

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/crypto"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// ProcessValidate handles the core validation operation logic
func ProcessValidate(cfg *dbconfig.Config) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting database configuration validation")

	// Show validation info
	terminal.PrintSubHeader("🔍 Configuration Validation")
	terminal.PrintInfo("This will decrypt and test your database configuration.")
	terminal.PrintInfo("A connection test will be performed to verify credentials.")
	fmt.Println()

	// Process validation
	return validateSpecificConfig(cfg.FilePath)
}

// validateSpecificConfig validates specific config with enhanced display
func validateSpecificConfig(filePath string) error {
	result := &dbconfig.ValidationResult{}

	// Step 1: File validation
	terminal.PrintSubHeader("📁 File Validation")
	spinner := terminal.NewProgressSpinner("Validating configuration file...")
	spinner.Start()

	if err := common.ValidateConfigFile(filePath); err != nil {
		spinner.Stop()
		result.FileValid = false
		terminal.PrintError(fmt.Sprintf("File validation failed: %v", err))
		return err
	}

	spinner.Stop()
	result.FileValid = true
	terminal.PrintSuccess("✅ Configuration file is valid")

	// Step 2: Decryption
	terminal.PrintSubHeader("🔐 Decryption")
	terminal.PrintInfo("Enter your encryption password to decrypt the configuration.")

	encryptionPassword, err := crypto.GetEncryptionPassword("🔑 Encryption password: ")
	if err != nil {
		return fmt.Errorf("failed to get encryption password: %w", err)
	}

	spinner = terminal.NewProgressSpinner("Decrypting configuration...")
	spinner.Start()

	dbConfig, err := common.LoadEncryptedConfigFromFile(filePath, encryptionPassword)
	spinner.Stop()

	if err != nil {
		result.DecryptionValid = false
		terminal.PrintError("❌ Decryption failed")
		return common.HandleDecryptionError(err, filePath)
	}

	result.DecryptionValid = true
	terminal.PrintSuccess("✅ Configuration decrypted successfully")

	// Step 3: Display configuration info
	terminal.PrintSubHeader("📋 Configuration Details")
	headers := []string{"Property", "Value"}
	rows := [][]string{
		{"📁 Source File", filePath},
		{"🌐 Host", dbConfig.Host},
		{"🔌 Port", fmt.Sprintf("%d", dbConfig.Port)},
		{"👤 Username", dbConfig.User},
	}
	terminal.FormatTable(headers, rows)

	// Step 4: Connection test
	terminal.PrintSubHeader("🔗 Database Connection Test")
	serverVersion, err := testDatabaseConnection(dbConfig)
	if err != nil {
		result.ConnectionValid = false
		return err
	}

	result.ConnectionValid = true
	result.ServerVersion = serverVersion

	// Display final results
	dbconfig.DisplayValidationResults(result, serverVersion)
	terminal.PrintSuccess("🎉 Configuration validation completed successfully!")
	terminal.WaitForEnterWithMessage("\nPress Enter to continue...")

	return nil
}

// testDatabaseConnection tests database connection with enhanced display
func testDatabaseConnection(dbConfig *config.EncryptedDatabaseConfig) (string, error) {
	// Build DSN for MySQL connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port)

	// Connection attempt with progress
	spinner := terminal.NewProgressSpinner("Connecting to database...")
	spinner.Start()

	// Open database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		spinner.Stop()
		terminal.PrintError(fmt.Sprintf("Failed to open connection: %v", err))
		return "", fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	// Set connection timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	spinner.UpdateMessage("Testing connection...")

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		spinner.Stop()
		terminal.PrintError(fmt.Sprintf("Connection test failed: %v", err))
		return "", fmt.Errorf("failed to ping database: %w", err)
	}

	spinner.UpdateMessage("Gathering server information...")

	// Get server version and info
	var version string
	err = db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)

	spinner.Stop()

	if err != nil {
		terminal.PrintWarning("Could not retrieve server version")
		version = "" // Reset version if failed
	}

	// Display success results
	terminal.PrintSuccess("✅ Database connection successful!")

	if version != "" {
		terminal.PrintSubHeader("📊 Server Information")
		terminal.PrintInfo(fmt.Sprintf("Version: %s", version))
	}

	return version, nil
}
