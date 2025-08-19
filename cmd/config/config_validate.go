package config

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
		if err := validateConfigCommand(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("Failed to validate config", logger.Error(err))
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var validateFilePath string

func init() {
	ValidateCmd.Flags().StringVarP(&validateFilePath, "file", "f", "", "Specific encrypted config file to validate")
}

func validateConfigCommand(cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting database configuration validation")

	// If specific file is provided via flag
	if validateFilePath != "" {
		return validateSpecificConfig(validateFilePath)
	}

	// List all encrypted config files and let user choose
	selectedFile, err := common.SelectConfigFileInteractive()
	if err != nil {
		return err
	}
	return validateSpecificConfig(selectedFile)
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

	err = testDatabaseConnection(dbConfig)
	if err != nil {
		fmt.Printf("❌ Connection Failed: %v\n", err)
		return fmt.Errorf("database connection test failed")
	}

	fmt.Println("✅ Database connection successful!")
	fmt.Println("✅ Configuration validation completed successfully!")

	return nil
}

func testDatabaseConnection(dbConfig *config.EncryptedDatabaseConfig) error {
	// Build DSN for MySQL connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port)

	// Open database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	// Set connection timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}
