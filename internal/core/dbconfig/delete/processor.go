package delete

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"
)

// ProcessDelete handles the core delete operation logic
func ProcessDelete(cfg *dbconfig.Config, args []string) error {
	// If --all flag is used
	if cfg.DeleteAll {
		return processDeleteAll(cfg)
	}

	// If specific file is provided via flag
	if cfg.FilePath != "" {
		return processDeleteSpecific(cfg.FilePath, cfg.ForceDelete)
	}

	// If files are provided as arguments
	if len(args) > 0 {
		return processDeleteMultiple(args, cfg.ForceDelete)
	}

	// Interactive mode - list all encrypted config files and let user choose
	return processDeleteWithSelection(cfg.ForceDelete)
}

// processDeleteSpecific deletes a specific configuration file
func processDeleteSpecific(filePath string, forceDelete bool) error {
	lg, _ := logger.Get()

	// Validate config file
	if err := common.ValidateConfigFile(filePath); err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}

	// Get file info for display
	filename := filepath.Base(filePath)

	// Confirmation unless force flag is used
	if !forceDelete {
		if !dbconfig.ConfirmSingleDeletion(filename) {
			terminal.PrintWarning("❌ Deletion cancelled.")
			return nil
		}
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete config file %s: %w", filename, err)
	}

	terminal.PrintSuccess(fmt.Sprintf("✅ Successfully deleted config file: %s", filename))
	lg.Info("Config file deleted", logger.String("file", filePath))
	return nil
}

// processDeleteMultiple deletes multiple configuration files
func processDeleteMultiple(filePaths []string, forceDelete bool) error {
	lg, _ := logger.Get()
	result := &dbconfig.DeleteResult{}

	var validFiles []string

	// Validate all files first
	for _, filePath := range filePaths {
		if err := common.ValidateConfigFile(filePath); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Invalid config file: %s (%v)", filepath.Base(filePath), err))
			terminal.PrintError(fmt.Sprintf("❌ Invalid config file: %s (%v)", filepath.Base(filePath), err))
		} else {
			validFiles = append(validFiles, filePath)
		}
	}

	if len(validFiles) == 0 {
		return fmt.Errorf("no valid config files to delete")
	}

	if len(result.Errors) > 0 {
		terminal.PrintWarning(fmt.Sprintf("⚠️  %d file(s) will be skipped due to validation errors.", len(result.Errors)))
	}

	// Show files to be deleted
	terminal.PrintSubHeader("📁 Files to be deleted")
	for i, filePath := range validFiles {
		fmt.Printf("   %d. %s\n", i+1, filepath.Base(filePath))
	}

	// Confirmation unless force flag is used
	if !forceDelete {
		if !dbconfig.ConfirmMultipleDeletion(len(validFiles)) {
			terminal.PrintWarning("❌ Deletion cancelled.")
			return nil
		}
	}

	// Delete files
	for _, filePath := range validFiles {
		filename := filepath.Base(filePath)
		if err := os.Remove(filePath); err != nil {
			errorMsg := fmt.Sprintf("Failed to delete %s: %v", filename, err)
			result.Errors = append(result.Errors, errorMsg)
			terminal.PrintError(fmt.Sprintf("❌ %s", errorMsg))
			lg.Error("Failed to delete config file",
				logger.String("file", filePath),
				logger.Error(err))
			result.ErrorCount++
		} else {
			result.DeletedFiles = append(result.DeletedFiles, filePath)
			terminal.PrintSuccess(fmt.Sprintf("✅ Deleted: %s", filename))
			lg.Info("Config file deleted", logger.String("file", filePath))
			result.DeletedCount++
		}
	}

	dbconfig.DisplayDeleteSummary(result)

	if result.ErrorCount > 0 {
		return fmt.Errorf("failed to delete %d config file(s)", result.ErrorCount)
	}

	return nil
}

// processDeleteAll deletes all encrypted configuration files
func processDeleteAll(cfg *dbconfig.Config) error {
	// Get config directory
	configDir, err := config.GetDatabaseConfigDirectory()
	if err != nil {
		return fmt.Errorf("failed to get database config directory: %w", err)
	}

	// Find all encrypted config files
	encFiles, err := common.FindEncryptedConfigFiles(configDir)
	if err != nil {
		return fmt.Errorf("failed to find encrypted config files: %w", err)
	}

	if len(encFiles) == 0 {
		terminal.PrintInfo("ℹ️  No encrypted configuration files found to delete.")
		return nil
	}

	// Show files to be deleted
	terminal.PrintWarning(fmt.Sprintf("⚠️  About to delete ALL %d encrypted configuration files:", len(encFiles)))
	terminal.PrintSubHeader("Files to be deleted")
	for i, file := range encFiles {
		filename := filepath.Base(file)
		fmt.Printf("   %d. %s\n", i+1, filename)
	}

	// Confirmation unless force flag is used
	if !cfg.ForceDelete {
		if !dbconfig.ConfirmAllDeletion(len(encFiles)) {
			terminal.PrintWarning("❌ Deletion cancelled.")
			return nil
		}
	}

	// Delete all files using the multiple delete logic
	err = processDeleteMultiple(encFiles, cfg.ForceDelete)
	if err != nil {
		return err
	}

	terminal.PrintSuccess("🗑️  All encrypted configuration files have been deleted.")
	return nil
}

// processDeleteWithSelection lists all encrypted config files and lets user choose
func processDeleteWithSelection(forceDelete bool) error {
	// Get config directory
	configDir, err := config.GetDatabaseConfigDirectory()
	if err != nil {
		return fmt.Errorf("failed to get database config directory: %w", err)
	}

	encFiles, err := common.FindEncryptedConfigFiles(configDir)
	if err != nil {
		return fmt.Errorf("failed to find encrypted config files: %w", err)
	}

	if len(encFiles) == 0 {
		terminal.PrintError("❌ No encrypted configuration files found.")
		terminal.PrintInfo("   Use 'dbconfig generate' to create one.")
		return nil
	}

	// Display available files
	terminal.PrintSubHeader("📁 Available Encrypted Configuration Files")
	for i, file := range encFiles {
		filename := filepath.Base(file)
		fmt.Printf("   %d. %s\n", i+1, filename)
	}

	// Let user choose
	choice, err := dbconfig.PromptForSelection(len(encFiles))
	if err != nil {
		return err
	}

	// Handle 'all' selection
	if strings.ToLower(choice) == "all" {
		return processDeleteAllFromList(encFiles, forceDelete)
	}

	// Parse comma-separated selections
	selectedFiles, err := dbconfig.ParseFileSelections(choice, encFiles)
	if err != nil {
		return err
	}

	return processDeleteMultiple(selectedFiles, forceDelete)
}

// processDeleteAllFromList deletes all files from the provided list
func processDeleteAllFromList(encFiles []string, forceDelete bool) error {
	terminal.PrintWarning(fmt.Sprintf("⚠️  About to delete ALL %d encrypted configuration files.", len(encFiles)))

	if !forceDelete {
		if !dbconfig.ConfirmAllDeletion(len(encFiles)) {
			terminal.PrintWarning("❌ Deletion cancelled.")
			return nil
		}
	}

	return processDeleteMultiple(encFiles, forceDelete)
}
