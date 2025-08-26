package dbconfig_cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/crypto"
	"sfDBTools/utils/terminal"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/spf13/cobra"
)

var ValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate encrypted database configuration and test connection",
	Long: `Validate that the encrypted database configuration can be properly decrypted
and test the actual database connection. If no file is specified, it will list all 
available encrypted config files and allow you to choose one.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Clear screen and show header
		terminal.ClearAndShowHeader("✅ Validate Database Configuration")

		if err := validateConfigCommandEnhanced(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("Failed to validate config", logger.Error(err))
			terminal.PrintError(fmt.Sprintf("Validation failed: %v", err))
			terminal.WaitForEnterWithMessage("Press Enter to continue...")
			os.Exit(1)
		}
	},
}

var validateFilePath string

func init() {
	ValidateCmd.Flags().StringVarP(&validateFilePath, "file", "f", "", "Specific encrypted config file to validate")
}

// validateConfigCommandEnhanced is the enhanced version with terminal utilities
func validateConfigCommandEnhanced(cmd *cobra.Command) error {
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

	// If specific file is provided via flag
	if validateFilePath != "" {
		terminal.PrintInfo(fmt.Sprintf("Validating file: %s", validateFilePath))
		return validateSpecificConfigEnhanced(validateFilePath)
	}

	// List all encrypted config files and let user choose
	terminal.PrintSubHeader("📂 Select Configuration File")
	selectedFile, err := common.SelectConfigFileInteractive()
	if err != nil {
		return err
	}
	return validateSpecificConfigEnhanced(selectedFile)
}

// validateSpecificConfigEnhanced validates specific config with enhanced display
func validateSpecificConfigEnhanced(filePath string) error {
	// Step 1: File validation
	terminal.PrintSubHeader("📁 File Validation")
	spinner := terminal.NewProgressSpinner("Validating configuration file...")
	spinner.Start()

	if err := common.ValidateConfigFile(filePath); err != nil {
		spinner.Stop()
		terminal.PrintError(fmt.Sprintf("File validation failed: %v", err))
		return err
	}

	spinner.Stop()
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
		terminal.PrintError("❌ Decryption failed")
		return common.HandleDecryptionError(err, filePath)
	}

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
	return testDatabaseConnectionEnhanced(dbConfig)
}

func validateSpecificConfig(filePath string) error {
	// Validate config file
	if err := common.ValidateConfigFile(filePath); err != nil {
		return err
	}

	// Get encryption password (use environment variable if available)
	encryptionPassword, err := crypto.GetEncryptionPassword("Enter encryption password: ")
	if err != nil {
		return fmt.Errorf("failed to get encryption password: %w", err)
	}

	// Load and decrypt the configuration
	dbConfig, err := common.LoadEncryptedConfigFromFile(filePath, encryptionPassword)
	if err != nil {
		return common.HandleDecryptionError(err, filePath)
	}

	// Display configuration info
	fmt.Println("🔧 Database Configuration:")
	fmt.Println("==========================")
	fmt.Printf("📁 Source: %s\n", filePath)
	fmt.Printf("   Host: %s\n", dbConfig.Host)
	fmt.Printf("   Port: %d\n", dbConfig.Port)
	fmt.Printf("   User: %s\n", dbConfig.User)

	// Test database connection
	fmt.Println("\n🔗 Testing Database Connection:")
	fmt.Println("===============================")

	err = testDatabaseConnectionEnhanced(dbConfig)
	if err != nil {
		fmt.Printf("❌ Connection Failed: %v\n", err)
		return fmt.Errorf("database connection test failed")
	}

	fmt.Println("✅ Database connection successful!")
	fmt.Println("✅ Configuration validation completed successfully!")

	return nil
}

// testDatabaseConnectionEnhanced tests database connection with enhanced display
func testDatabaseConnectionEnhanced(dbConfig *config.EncryptedDatabaseConfig) error {
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
		return fmt.Errorf("failed to open database connection: %w", err)
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
		return fmt.Errorf("failed to ping database: %w", err)
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

	// Connection summary
	terminal.PrintSubHeader("📋 Connection Summary")
	headers := []string{"Test", "Result", "Details"}
	rows := [][]string{
		{"Connection", terminal.ColorText("✅ Success", terminal.ColorGreen), "Database server is reachable"},
		{"Authentication", terminal.ColorText("✅ Success", terminal.ColorGreen), "Credentials are valid"},
		{"Response Time", terminal.ColorText("< 10s", terminal.ColorGreen), "Connection within timeout"},
	}

	if version != "" {
		rows = append(rows, []string{"Server Version", terminal.ColorText("✅ Retrieved", terminal.ColorGreen), version})
	}

	terminal.FormatTable(headers, rows)

	terminal.PrintSuccess("🎉 Configuration validation completed successfully!")
	terminal.WaitForEnterWithMessage("\nPress Enter to continue...")

	return nil
}
