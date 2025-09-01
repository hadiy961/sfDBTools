package delete

import (
	"fmt"
	"path/filepath"

	coredbconfig "sfDBTools/internal/core/dbconfig"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"
)

// Processor handles delete operations for database configurations
type Processor struct {
	*coredbconfig.BaseProcessor
	configHelper *coredbconfig.ConfigHelper
}

// NewProcessor creates a new delete processor
func NewProcessor() (*Processor, error) {
	base, err := coredbconfig.NewBaseProcessor()
	if err != nil {
		return nil, err
	}

	configHelper, err := coredbconfig.NewConfigHelper()
	if err != nil {
		return nil, err
	}

	return &Processor{
		BaseProcessor: base,
		configHelper:  configHelper,
	}, nil
}

// ProcessDelete handles the core delete operation logic
func ProcessDelete(cfg *dbconfig.Config, args []string) error {
	processor, err := NewProcessor()
	if err != nil {
		return err
	}

	processor.LogOperation("database configuration deletion", "")

	// Route to appropriate handler based on parameters
	switch {
	case cfg.DeleteAll:
		return processor.processDeleteAll(cfg)
	case cfg.FilePath != "":
		return processor.processDeleteSpecific(cfg.FilePath, cfg.ForceDelete)
	case len(args) > 0:
		return processor.processDeleteMultiple(args, cfg.ForceDelete)
	default:
		return processor.processDeleteWithSelection(cfg.ForceDelete)
	}
}

// processDeleteSpecific deletes a specific configuration file
func (p *Processor) processDeleteSpecific(filePath string, forceDelete bool) error {
	// Validate config file
	if err := p.configHelper.ValidateConfigExists(filePath); err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}

	// Get file info for display
	filename := filepath.Base(filePath)

	// Confirmation unless force flag is used
	if !forceDelete {
		if !dbconfig.ConfirmDeletion(dbconfig.DeletionSingle, []string{filename}) {
			terminal.PrintWarning("❌ Deletion cancelled.")
			return nil
		}
	}

	// Delete the file
	if err := p.configHelper.DeleteConfigFile(filePath); err != nil {
		return fmt.Errorf("failed to delete config file %s: %w", filename, err)
	}

	terminal.PrintSuccess(fmt.Sprintf("✅ Configuration '%s' deleted successfully", filename))
	return nil
}

// processDeleteMultiple deletes multiple configuration files
func (p *Processor) processDeleteMultiple(filePaths []string, forceDelete bool) error {
	fileManager := p.configHelper.GetFileManager()

	// Validate all files first
	validFiles := []string{}
	for _, filePath := range filePaths {
		if err := p.configHelper.ValidateConfigExists(filePath); err != nil {
			terminal.PrintWarning(fmt.Sprintf("⚠️ Skipping invalid file: %s", filepath.Base(filePath)))
			continue
		}
		validFiles = append(validFiles, filePath)
	}

	if len(validFiles) == 0 {
		return fmt.Errorf("no valid configuration files to delete")
	}

	// Show files to be deleted
	terminal.PrintSubHeader("📁 Files to be deleted:")
	for i, filePath := range validFiles {
		fmt.Printf("   %d. %s\n", i+1, filepath.Base(filePath))
	}

	// Confirmation unless force flag is used
	if !forceDelete {
		filenames := make([]string, len(validFiles))
		for i, filePath := range validFiles {
			filenames[i] = filepath.Base(filePath)
		}
		if !dbconfig.ConfirmDeletion(dbconfig.DeletionMultiple, filenames) {
			terminal.PrintWarning("❌ Deletion cancelled.")
			return nil
		}
	}

	// Delete files using FileManager
	result := fileManager.DeleteMultipleFiles(validFiles, true)

	// Display result
	dbconfig.DisplayDeleteSummary(result)

	if result.ErrorCount > 0 {
		return fmt.Errorf("some files could not be deleted")
	}

	return nil
}

// processDeleteAll deletes all configuration files
func (p *Processor) processDeleteAll(cfg *dbconfig.Config) error {
	fileManager := p.configHelper.GetFileManager()

	// Get all config files
	files, err := fileManager.ListConfigFiles()
	if err != nil {
		return fmt.Errorf("error listing config files: %v", err)
	}

	if len(files) == 0 {
		terminal.PrintWarning("No configuration files found to delete.")
		return nil
	}

	// Show files to be deleted
	terminal.PrintSubHeader(fmt.Sprintf("📁 All Configuration Files (%d found):", len(files)))
	for i, file := range files {
		fmt.Printf("   %d. %s\n", i+1, file.Name)
	}

	// Confirmation unless force flag is used
	if !cfg.ForceDelete {
		filenames := make([]string, len(files))
		for i, file := range files {
			filenames[i] = file.Name
		}
		if !dbconfig.ConfirmDeletion(dbconfig.DeletionAll, filenames) {
			terminal.PrintWarning("❌ Deletion cancelled.")
			return nil
		}
	}

	// Extract file paths
	filePaths := make([]string, len(files))
	for i, file := range files {
		filePaths[i] = file.Path
	}

	// Delete all files using FileManager
	result := fileManager.DeleteMultipleFiles(filePaths, true)

	// Display result
	dbconfig.DisplayDeleteSummary(result)

	if result.ErrorCount > 0 {
		return fmt.Errorf("some files could not be deleted")
	}

	return nil
}

// processDeleteWithSelection allows user to select files for deletion
func (p *Processor) processDeleteWithSelection(forceDelete bool) error {
	fileManager := p.configHelper.GetFileManager()

	// Get all config files
	files, err := fileManager.ListConfigFiles()
	if err != nil {
		return fmt.Errorf("error listing config files: %v", err)
	}

	if len(files) == 0 {
		terminal.PrintWarning("No configuration files found.")
		return nil
	}

	// Let user select files
	selectedPaths, err := dbconfig.SelectFilesForDeletion(files)
	if err != nil {
		return err
	}

	return p.processDeleteMultiple(selectedPaths, forceDelete)
}
