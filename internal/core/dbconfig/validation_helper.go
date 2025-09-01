package dbconfig

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"sfDBTools/internal/config"
	"sfDBTools/utils/common"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// ValidationHelper provides common validation functionality
type ValidationHelper struct {
	*BaseProcessor
}

// NewValidationHelper creates a new validation helper
func NewValidationHelper() (*ValidationHelper, error) {
	base, err := NewBaseProcessor()
	if err != nil {
		return nil, err
	}

	return &ValidationHelper{
		BaseProcessor: base,
	}, nil
}

// ValidateConfigFile validates a config file with comprehensive checks
func (vh *ValidationHelper) ValidateConfigFile(filePath string) (*dbconfig.ValidationResult, error) {
	// Use the validation module to validate the file
	result, err := dbconfig.ValidateConfigFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("validation error: %v", err)
	}

	return result, nil
}

// ValidateWithDecryption performs validation by attempting decryption and connection
func (vh *ValidationHelper) ValidateWithDecryption(filePath string, result *dbconfig.ValidationResult) error {
	// Get encryption password
	encryptionPassword, err := vh.GetEncryptionPassword("validate the configuration")
	if err != nil {
		return err
	}

	// Load and decrypt configuration
	spinner := terminal.NewProgressSpinner("Decrypting configuration...")
	spinner.Start()

	dbConfig, err := common.LoadEncryptedConfigFromFile(filePath, encryptionPassword)
	spinner.Stop()

	if err != nil {
		result.Errors = append(result.Errors, "Decryption failed")
		terminal.PrintError("❌ Decryption failed")
		return common.HandleDecryptionError(err, filePath)
	}

	terminal.PrintSuccess("✅ Configuration decrypted successfully")

	// Display configuration info
	vh.displayConfigInfo(dbConfig)

	// Test database connection
	return vh.testDatabaseConnection(dbConfig, result)
}

// displayConfigInfo shows configuration details in a formatted way
func (vh *ValidationHelper) displayConfigInfo(dbConfig *config.EncryptedDatabaseConfig) {
	terminal.PrintSubHeader("📋 Configuration Details")

	headers := []string{"Property", "Value"}
	rows := [][]string{
		{"🏠 Host", dbConfig.Host},
		{"🔌 Port", fmt.Sprintf("%d", dbConfig.Port)},
		{"👤 Username", dbConfig.User},
	}
	terminal.FormatTable(headers, rows)
}

// testDatabaseConnection tests database connection with timeout
func (vh *ValidationHelper) testDatabaseConnection(dbConfig *config.EncryptedDatabaseConfig, result *dbconfig.ValidationResult) error {
	terminal.PrintSubHeader("🔗 Database Connection Test")

	// Build DSN for MySQL connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port)

	// Connection attempt with progress
	spinner := terminal.NewProgressSpinner("Connecting to database...")
	spinner.Start()

	// Use context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		spinner.Stop()
		result.Errors = append(result.Errors, fmt.Sprintf("Connection failed: %v", err))
		terminal.PrintError("❌ Failed to connect to database")
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	// Test connection
	err = db.PingContext(ctx)
	if err != nil {
		spinner.Stop()
		result.Errors = append(result.Errors, fmt.Sprintf("Connection test failed: %v", err))
		terminal.PrintError("❌ Database connection test failed")
		return fmt.Errorf("database connection test failed: %w", err)
	}

	// Get server version
	var serverVersion string
	err = db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&serverVersion)
	spinner.Stop()

	if err != nil {
		result.Warnings = append(result.Warnings, "Could not retrieve server version")
		terminal.PrintWarning("⚠️ Could not retrieve server version")
	} else {
		result.TestResults["connection_test"] = true
		terminal.PrintSuccess("✅ Database connection successful")
		terminal.PrintInfo(fmt.Sprintf("🗃️ Server Version: %s", serverVersion))
	}

	return nil
}

// ValidateFileBasic performs basic file validation without decryption
func (vh *ValidationHelper) ValidateFileBasic(filePath string) error {
	if err := common.ValidateConfigFile(filePath); err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}
	return nil
}
