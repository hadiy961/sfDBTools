package delete

import (
	"fmt"
	"path/filepath"

	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"
)

// processDeleteSpecific deletes a specific configuration file
func (p *Processor) processDeleteSpecific(filePath string, forceDelete bool) error {
	// Validate and delete using the shared deletion path to avoid duplication
	valid, err := p.validateFiles([]string{filePath})
	if err != nil {
		return fmt.Errorf("delete: validate file '%s': %w", filepath.Base(filePath), err)
	}

	if err := p.deleteValidatedPaths(valid, dbconfig.DeletionSingle, forceDelete); err != nil {
		return err
	}
	return nil
}

// processDeleteMultiple deletes multiple configuration files
func (p *Processor) processDeleteMultiple(filePaths []string, forceDelete bool) error {
	// Validate and delete using shared helper
	validFiles, err := p.validateFiles(filePaths)
	if err != nil {
		return fmt.Errorf("delete: validate multiple files: %w", err)
	}
	return p.deleteValidatedPaths(validFiles, dbconfig.DeletionMultiple, forceDelete)
}

// processDeleteAll deletes all configuration files
func (p *Processor) processDeleteAll(cfg *dbconfig.Config) error {
	fileManager := p.configHelper.GetFileManager()

	// Get all config files
	files, err := fileManager.ListConfigFiles()
	if err != nil {
		return fmt.Errorf("delete: error listing config files: %v", err)
	}

	if len(files) == 0 {
		terminal.PrintWarning("No configuration files found to delete.")
		return nil
	}

	// Prepare paths
	paths := make([]string, len(files))
	for i, file := range files {
		paths[i] = file.Path
	}

	// Use shared deletion helper
	return p.deleteValidatedPaths(paths, dbconfig.DeletionAll, cfg.ForceDelete)
}

// processDeleteWithSelection allows user to select files for deletion
func (p *Processor) processDeleteWithSelection(forceDelete bool) error {
	fileManager := p.configHelper.GetFileManager()

	// Get all config files
	files, err := fileManager.ListConfigFiles()
	if err != nil {
		return fmt.Errorf("delete: error listing config files: %v", err)
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
