package delete

import (
	"fmt"
	"path/filepath"

	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"
)

// validateFiles validates paths and returns only the valid ones or an error if none valid
func (p *Processor) validateFiles(filePaths []string) ([]string, error) {
	valid := make([]string, 0, len(filePaths))
	for _, fp := range filePaths {
		if err := p.configHelper.ValidateConfigExists(fp); err != nil {
			terminal.PrintWarning(fmt.Sprintf("⚠️ Skipping invalid file: %s", filepath.Base(fp)))
			continue
		}
		valid = append(valid, fp)
	}
	if len(valid) == 0 {
		return nil, fmt.Errorf("delete: no valid configuration files to delete")
	}
	return valid, nil
}

// namesFromPaths returns base filenames for a list of paths
func (p *Processor) namesFromPaths(paths []string) []string {
	names := make([]string, len(paths))
	for i, pth := range paths {
		names[i] = filepath.Base(pth)
	}
	return names
}

// printFileList displays a subheader and enumerated file names
func (p *Processor) printFileList(title string, names []string) {
	terminal.PrintSubHeader(title)
	for i, n := range names {
		terminal.PrintInfo(fmt.Sprintf("   %d. %s", i+1, n))
	}
}

// confirmList prints names and runs confirmation unless forced
func (p *Processor) confirmList(title string, names []string, kind dbconfig.DeletionType, force bool) bool {
	p.printFileList(title, names)
	if force {
		return true
	}
	if !dbconfig.ConfirmDeletion(kind, names) {
		terminal.PrintWarning("❌ Deletion cancelled.")
		return false
	}
	return true
}

// deleteValidatedPaths deletes already-validated paths using the FileManager and displays a summary
func (p *Processor) deleteValidatedPaths(paths []string, kind dbconfig.DeletionType, force bool) error {
	if len(paths) == 0 {
		return fmt.Errorf("delete: no files provided for deletion")
	}

	fileManager := p.configHelper.GetFileManager()

	names := p.namesFromPaths(paths)
	title := " Files to be deleted:"
	if kind == dbconfig.DeletionAll {
		title = fmt.Sprintf("All Configuration Files (%d found):", len(paths))
	}

	if !p.confirmList(title, names, kind, force) {
		return nil
	}

	result := fileManager.DeleteMultipleFiles(paths, true)
	dbconfig.DisplayDeleteSummary(result)
	if result.ErrorCount > 0 {
		return fmt.Errorf("delete: some files could not be deleted")
	}
	return nil
}
