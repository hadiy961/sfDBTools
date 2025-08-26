package dbconfig_cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete encrypted database configuration files",
	Long: `Delete encrypted database configuration files.
⚠️  WARNING: Deleted files cannot be recovered. Use with caution.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := deleteConfigCommand(cmd, args); err != nil {
			lg, _ := logger.Get()
			lg.Error("Failed to delete config", logger.Error(err))
			os.Exit(1)
		}
	},
}

var (
	deleteConfigFile string
	forceDelete      bool
	deleteAll        bool
)

func init() {
	DeleteCmd.Flags().StringVarP(&deleteConfigFile, "file", "f", "", "Specific encrypted config file to delete")
	DeleteCmd.Flags().BoolVar(&forceDelete, "force", false, "Skip confirmation prompts")
	DeleteCmd.Flags().BoolVar(&deleteAll, "all", false, "Delete all encrypted config files")
}

func deleteConfigCommand(cmd *cobra.Command, args []string) error {

	// If --all flag is used
	if deleteAll {
		return deleteAllConfigs()
	}

	// If specific file is provided via flag
	if deleteConfigFile != "" {
		return deleteSpecificConfig(deleteConfigFile)
	}

	// If files are provided as arguments
	if len(args) > 0 {
		return deleteMultipleConfigs(args)
	}

	// Interactive mode - list all encrypted config files and let user choose
	return deleteConfigWithSelection()
}

// deleteSpecificConfig deletes a specific configuration file
func deleteSpecificConfig(filePath string) error {
	lg, _ := logger.Get()

	// Validate config file
	if err := common.ValidateConfigFile(filePath); err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}

	// Get file info for display
	filename := filepath.Base(filePath)

	// Confirmation unless force flag is used
	if !forceDelete {
		if !confirmDeletion(filename) {
			fmt.Println("❌ Deletion cancelled.")
			return nil
		}
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete config file %s: %w", filename, err)
	}

	fmt.Printf("✅ Successfully deleted config file: %s\n", filename)
	lg.Info("Config file deleted", logger.String("file", filePath))
	return nil
}

// deleteMultipleConfigs deletes multiple configuration files
func deleteMultipleConfigs(filePaths []string) error {
	lg, _ := logger.Get()

	var validFiles []string
	var invalidFiles []string

	// Validate all files first
	for _, filePath := range filePaths {
		if err := common.ValidateConfigFile(filePath); err != nil {
			invalidFiles = append(invalidFiles, filePath)
			fmt.Printf("❌ Invalid config file: %s (%v)\n", filepath.Base(filePath), err)
		} else {
			validFiles = append(validFiles, filePath)
		}
	}

	if len(validFiles) == 0 {
		return fmt.Errorf("no valid config files to delete")
	}

	if len(invalidFiles) > 0 {
		fmt.Printf("\n⚠️  %d file(s) will be skipped due to validation errors.\n", len(invalidFiles))
	}

	// Show files to be deleted
	fmt.Println("\n📁 Files to be deleted:")
	for i, filePath := range validFiles {
		fmt.Printf("   %d. %s\n", i+1, filepath.Base(filePath))
	}

	// Confirmation unless force flag is used
	if !forceDelete {
		if !confirmMultipleDeletion(len(validFiles)) {
			fmt.Println("❌ Deletion cancelled.")
			return nil
		}
	}

	// Delete files
	var deletedCount int
	var errorCount int

	for _, filePath := range validFiles {
		filename := filepath.Base(filePath)
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("❌ Failed to delete %s: %v\n", filename, err)
			lg.Error("Failed to delete config file",
				logger.String("file", filePath),
				logger.Error(err))
			errorCount++
		} else {
			fmt.Printf("✅ Deleted: %s\n", filename)
			lg.Info("Config file deleted", logger.String("file", filePath))
			deletedCount++
		}
	}

	fmt.Printf("\n📊 Summary: %d deleted, %d errors\n", deletedCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("failed to delete %d config file(s)", errorCount)
	}

	return nil
}

// deleteAllConfigs deletes all encrypted configuration files
func deleteAllConfigs() error {
	lg, _ := logger.Get()

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
		fmt.Println("ℹ️  No encrypted configuration files found to delete.")
		return nil
	}

	// Show files to be deleted
	fmt.Printf("⚠️  About to delete ALL %d encrypted configuration files:\n", len(encFiles))
	fmt.Println("==========================================")
	for i, file := range encFiles {
		filename := filepath.Base(file)
		fmt.Printf("   %d. %s\n", i+1, filename)
	}

	// Confirmation unless force flag is used
	if !forceDelete {
		if !confirmAllDeletion(len(encFiles)) {
			fmt.Println("❌ Deletion cancelled.")
			return nil
		}
	}

	// Delete all files
	var deletedCount int
	var errorCount int

	for _, filePath := range encFiles {
		filename := filepath.Base(filePath)
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("❌ Failed to delete %s: %v\n", filename, err)
			lg.Error("Failed to delete config file",
				logger.String("file", filePath),
				logger.Error(err))
			errorCount++
		} else {
			fmt.Printf("✅ Deleted: %s\n", filename)
			lg.Info("Config file deleted", logger.String("file", filePath))
			deletedCount++
		}
	}

	fmt.Printf("\n📊 Summary: %d deleted, %d errors\n", deletedCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("failed to delete %d config file(s)", errorCount)
	}

	fmt.Println("🗑️  All encrypted configuration files have been deleted.")
	return nil
}

// deleteConfigWithSelection lists all encrypted config files and lets user choose
func deleteConfigWithSelection() error {

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
		fmt.Println("❌ No encrypted configuration files found.")
		fmt.Println("   Use 'config generate' to create one.")
		return nil
	}

	// Display available files
	fmt.Println("📁 Available Encrypted Configuration Files:")
	fmt.Println("==========================================")
	for i, file := range encFiles {
		filename := filepath.Base(file)
		fmt.Printf("   %d. %s\n", i+1, filename)
	}

	// Let user choose
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nSelect configuration file(s) to delete (1-%d, comma-separated, or 'all'): ", len(encFiles))
	choice, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read selection: %w", err)
	}

	choice = strings.TrimSpace(choice)

	// Handle 'all' selection
	if strings.ToLower(choice) == "all" {
		return deleteAllConfigsFromList(encFiles)
	}

	// Parse comma-separated selections
	selectedFiles, err := parseFileSelections(choice, encFiles)
	if err != nil {
		return err
	}

	return deleteMultipleConfigs(selectedFiles)
}

// deleteAllConfigsFromList deletes all files from the provided list
func deleteAllConfigsFromList(encFiles []string) error {
	fmt.Printf("⚠️  About to delete ALL %d encrypted configuration files.\n", len(encFiles))

	if !forceDelete {
		if !confirmAllDeletion(len(encFiles)) {
			fmt.Println("❌ Deletion cancelled.")
			return nil
		}
	}

	return deleteMultipleConfigs(encFiles)
}

// parseFileSelections parses comma-separated file selections
func parseFileSelections(choice string, encFiles []string) ([]string, error) {
	if choice == "" {
		return nil, fmt.Errorf("no selection made")
	}

	selections := strings.Split(choice, ",")
	var selectedFiles []string

	for _, sel := range selections {
		sel = strings.TrimSpace(sel)
		if sel == "" {
			continue
		}

		// Try to parse as number
		if index := parseIndex(sel, len(encFiles)); index >= 0 {
			selectedFiles = append(selectedFiles, encFiles[index])
		} else {
			return nil, fmt.Errorf("invalid selection: %s", sel)
		}
	}

	if len(selectedFiles) == 0 {
		return nil, fmt.Errorf("no valid selections made")
	}

	return selectedFiles, nil
}

// parseIndex parses a string index and returns the corresponding array index (-1 if invalid)
func parseIndex(s string, maxLen int) int {
	if idx := parseInt(s); idx >= 1 && idx <= maxLen {
		return idx - 1 // Convert to 0-based index
	}
	return -1
}

// parseInt safely parses an integer
func parseInt(s string) int {
	if i := 0; len(s) > 0 {
		for _, char := range s {
			if char < '0' || char > '9' {
				return -1
			}
			i = i*10 + int(char-'0')
		}
		return i
	}
	return -1
}

// Confirmation functions
func confirmDeletion(filename string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("⚠️  Are you sure you want to delete '%s'? This action cannot be undone. [y/N]: ", filename)
	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(strings.ToLower(confirmation))
	return confirmation == "y" || confirmation == "yes"
}

func confirmMultipleDeletion(count int) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("⚠️  Are you sure you want to delete %d config file(s)? This action cannot be undone. [y/N]: ", count)
	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(strings.ToLower(confirmation))
	return confirmation == "y" || confirmation == "yes"
}

func confirmAllDeletion(count int) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("⚠️  Are you sure you want to delete ALL %d encrypted configuration files? This action cannot be undone. [y/N]: ", count)
	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(strings.ToLower(confirmation))
	return confirmation == "y" || confirmation == "yes"
}
