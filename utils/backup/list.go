package backup_utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
	"sfDBTools/utils/database/info"

	"github.com/spf13/cobra"
)

// DatabaseListResult represents the result of processing a database list
type DatabaseListResult struct {
	ValidDatabases   []string
	InvalidDatabases []string
	TotalFound       int
}

// MultiBackupResult represents the result of backing up multiple databases
type MultiBackupResult struct {
	TotalProcessed   int
	SuccessCount     int
	FailedDatabases  []string
	InvalidDatabases []string
}

// ResolveDBListFile resolves the database list file, either from flag or interactive selection
func ResolveDBListFile(cmd *cobra.Command) (string, error) {
	// Check if db_list flag is provided
	dbListFile, err := cmd.Flags().GetString("db_list")
	if err != nil {
		return "", fmt.Errorf("failed to get db_list flag: %w", err)
	}

	// If flag is provided, validate and use it
	if dbListFile != "" {
		// Validate file extension
		if !strings.HasSuffix(strings.ToLower(dbListFile), ".txt") {
			return "", fmt.Errorf("db_list file must have .txt extension")
		}

		// Convert to absolute path if needed
		if !filepath.IsAbs(dbListFile) {
			wd, _ := os.Getwd()
			dbListFile = filepath.Join(wd, dbListFile)
		}

		// Check if file exists
		if _, err := os.Stat(dbListFile); os.IsNotExist(err) {
			return "", fmt.Errorf("db_list file does not exist: %s", dbListFile)
		}

		lg, _ := logger.Get()
		lg.Info("Using provided db_list file", logger.String("file", dbListFile))
		return dbListFile, nil
	}

	// If no flag provided, show interactive selection
	fmt.Println("📁 Select Database List File:")
	fmt.Println("=============================")

	// Look for .txt files in config/db_list directory
	dbListDir := "./config/db_list"
	files, err := os.ReadDir(dbListDir)
	if err != nil {
		return "", fmt.Errorf("failed to read db_list directory: %w", err)
	}

	var txtFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".txt") {
			txtFiles = append(txtFiles, file.Name())
		}
	}

	if len(txtFiles) == 0 {
		return "", fmt.Errorf("no .txt files found in %s directory", dbListDir)
	}

	// Display available files
	fmt.Printf("📁 Available Database List Files:\n")
	fmt.Printf("==================================\n")
	for i, file := range txtFiles {
		fmt.Printf("   %d. %s\n", i+1, file)
	}

	// Get user selection
	var choice int
	for {
		fmt.Printf("\nSelect database list file (1-%d): ", len(txtFiles))
		if _, err := fmt.Scanf("%d", &choice); err != nil || choice < 1 || choice > len(txtFiles) {
			fmt.Printf("Invalid selection. Please enter a number between 1 and %d.\n", len(txtFiles))
			continue
		}
		break
	}

	selectedFile := filepath.Join(dbListDir, txtFiles[choice-1])
	lg, _ := logger.Get()
	lg.Info("Selected db_list file", logger.String("file", selectedFile))

	return selectedFile, nil
}

// ValidateDatabaseList validates databases from list against available databases on server
func ValidateDatabaseList(dbConfig database.Config, databases []string) (*DatabaseListResult, error) {
	lg, _ := logger.Get()
	lg.Info("Validating databases from list")

	availableDatabases, err := info.ListDatabases(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get available databases: %w", err)
	}

	result := &DatabaseListResult{
		ValidDatabases:   []string{},
		InvalidDatabases: []string{},
		TotalFound:       len(databases),
	}

	for _, db := range databases {
		found := false
		for _, availableDB := range availableDatabases {
			if db == availableDB {
				found = true
				break
			}
		}
		if found {
			result.ValidDatabases = append(result.ValidDatabases, db)
		} else {
			result.InvalidDatabases = append(result.InvalidDatabases, db)
		}
	}

	return result, nil
}

// DisplayDatabaseListValidation displays the validation results
func DisplayDatabaseListValidation(result *DatabaseListResult) error {
	lg, _ := logger.Get()

	if len(result.InvalidDatabases) > 0 {
		lg.Warn("Some databases were not found on the server",
			logger.Int("invalid_count", len(result.InvalidDatabases)),
			logger.Strings("invalid_databases", result.InvalidDatabases))
	}

	if len(result.ValidDatabases) == 0 {
		return fmt.Errorf("no valid databases found to backup")
	}

	lg.Info("Valid databases found for backup",
		logger.Int("valid_count", len(result.ValidDatabases)),
		logger.Strings("valid_databases", result.ValidDatabases))

	return nil
}

// DisplayMultiBackupSummary displays the final summary of multiple database backup
func DisplayMultiBackupSummary(result *MultiBackupResult) {
	lg, _ := logger.Get()

	lg.Info("Multi-database backup summary",
		logger.Int("total_processed", result.TotalProcessed),
		logger.Int("successful", result.SuccessCount),
		logger.Int("failed", len(result.FailedDatabases)))

	if len(result.FailedDatabases) > 0 {
		lg.Error("Some databases failed to backup",
			logger.Strings("failed_databases", result.FailedDatabases))
	}

	if len(result.InvalidDatabases) > 0 {
		lg.Warn("Some databases were skipped (not found on server)",
			logger.Strings("skipped_databases", result.InvalidDatabases))
	}
}

// LogMultiBackupCompletion logs the completion of multi-database backup
func LogMultiBackupCompletion(result *MultiBackupResult, operationType string) {
	lg, _ := logger.Get()
	lg.Info(fmt.Sprintf("%s backup process completed", operationType),
		logger.Int("total", result.TotalProcessed),
		logger.Int("success", result.SuccessCount),
		logger.Int("failed", len(result.FailedDatabases)))
}

// TestDatabaseConnection tests the database connection and displays result
func TestDatabaseConnection(dbConfig database.Config) error {
	lg, _ := logger.Get()
	lg.Info("Testing database connection", logger.String("host", dbConfig.Host), logger.Int("port", dbConfig.Port))

	db, err := database.GetWithoutDB(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	lg.Info("Successfully connected to database", logger.String("host", dbConfig.Host), logger.Int("port", dbConfig.Port))
	return nil
}
