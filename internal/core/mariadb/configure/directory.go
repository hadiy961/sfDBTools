package configure

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// DirectoryManager handles directory creation and permissions
type DirectoryManager struct {
	settings *MariaDBSettings
}

// NewDirectoryManager creates a new directory manager
func NewDirectoryManager(settings *MariaDBSettings) *DirectoryManager {
	return &DirectoryManager{
		settings: settings,
	}
}

// SetupDirectories creates and configures all required directories
func (d *DirectoryManager) SetupDirectories() error {
	lg, _ := logger.Get()

	directories := []string{
		d.settings.DataDir,
		d.settings.BinlogDir,
		d.settings.LogDir,
	}

	// Create directories
	for _, dir := range directories {
		if err := d.createDirectory(dir); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Set ownership to mysql:mysql
	if err := d.setOwnership(directories); err != nil {
		return fmt.Errorf("failed to set directory ownership: %w", err)
	}

	lg.Info("All MariaDB directories created and configured successfully")
	terminal.PrintSuccess("All required directories created and configured")
	return nil
}

// createDirectory creates a directory if it doesn't exist
func (d *DirectoryManager) createDirectory(path string) error {
	lg, _ := logger.Get()

	// Check if directory already exists
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		lg.Info("Directory already exists", logger.String("path", path))
		return nil
	}

	// Create directory with proper permissions
	if err := os.MkdirAll(path, 0755); err != nil {
		lg.Error("Failed to create directory",
			logger.String("path", path),
			logger.Error(err))
		return err
	}

	lg.Info("Directory created successfully", logger.String("path", path))
	terminal.PrintInfo(fmt.Sprintf("Created directory: %s", path))
	return nil
}

// setOwnership sets mysql:mysql ownership on directories
func (d *DirectoryManager) setOwnership(directories []string) error {
	lg, _ := logger.Get()

	for _, dir := range directories {
		cmd := exec.Command("chown", "-R", "mysql:mysql", dir)
		if err := cmd.Run(); err != nil {
			lg.Error("Failed to set ownership",
				logger.String("directory", dir),
				logger.Error(err))
			return fmt.Errorf("failed to chown %s: %w", dir, err)
		}

		lg.Info("Ownership set successfully",
			logger.String("directory", dir),
			logger.String("owner", "mysql:mysql"))
	}

	terminal.PrintInfo("Directory ownership set to mysql:mysql")
	return nil
}

// DataMigrator handles data migration from default location
type DataMigrator struct {
	sourceDir    string
	targetDir    string
	removeSource bool
}

// NewDataMigrator creates a new data migrator
func NewDataMigrator(sourceDir, targetDir string) *DataMigrator {
	return &DataMigrator{
		sourceDir:    sourceDir,
		targetDir:    targetDir,
		removeSource: false, // Default: don't remove source
	}
}

// NewDataMigratorWithCleanup creates a new data migrator that removes source after successful migration
func NewDataMigratorWithCleanup(sourceDir, targetDir string) *DataMigrator {
	return &DataMigrator{
		sourceDir:    sourceDir,
		targetDir:    targetDir,
		removeSource: true,
	}
}

// SetRemoveSource configures whether to remove source directory after migration
func (m *DataMigrator) SetRemoveSource(remove bool) {
	m.removeSource = remove
}

// MigrateData migrates data from source to target directory
func (m *DataMigrator) MigrateData() error {
	lg, _ := logger.Get()

	// Check if source directory exists and has data
	if !m.hasDataToMigrate() {
		lg.Info("No data to migrate", logger.String("source", m.sourceDir))
		terminal.PrintInfo("No existing data to migrate")
		return nil
	}

	// Skip migration if source and target are the same
	if m.sourceDir == m.targetDir {
		lg.Info("Source and target directories are the same, skipping migration",
			logger.String("directory", m.sourceDir))
		terminal.PrintInfo("Source and target directories are the same, no migration needed")
		return nil
	}

	// Check if target directory is empty
	if !m.isTargetEmpty() {
		lg.Info("Target directory is not empty, skipping migration",
			logger.String("target", m.targetDir))
		terminal.PrintWarning("Target directory is not empty, skipping data migration")
		return nil
	}

	terminal.PrintInfo("Migrating data from default location...")

	// Use rsync for safe data migration
	cmd := exec.Command("rsync", "-av", m.sourceDir+"/", m.targetDir+"/")
	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to migrate data",
			logger.String("source", m.sourceDir),
			logger.String("target", m.targetDir),
			logger.Error(err),
			logger.String("output", string(output)))
		return fmt.Errorf("failed to migrate data: %w", err)
	}

	lg.Info("Data migration completed successfully",
		logger.String("source", m.sourceDir),
		logger.String("target", m.targetDir))

	terminal.PrintSuccess("Data migration completed successfully")

	// Ensure ownership is correct after migration
	if err := m.setTargetOwnership(); err != nil {
		lg.Warn("Failed to set ownership after migration",
			logger.String("target", m.targetDir),
			logger.Error(err))
		terminal.PrintWarning(fmt.Sprintf("Warning: Failed to set ownership on %s", m.targetDir))
	}

	// Remove source directory if configured to do so
	if m.removeSource {
		if err := m.removeSourceDirectory(); err != nil {
			// Log warning but don't fail the entire migration
			lg.Warn("Failed to remove source directory after migration",
				logger.String("source", m.sourceDir),
				logger.Error(err))
			terminal.PrintWarning(fmt.Sprintf("Warning: Failed to remove source directory %s", m.sourceDir))
		}
	}

	return nil
}

// hasDataToMigrate checks if source directory has data to migrate
func (m *DataMigrator) hasDataToMigrate() bool {
	// Check if source directory exists
	info, err := os.Stat(m.sourceDir)
	if os.IsNotExist(err) || !info.IsDir() {
		return false
	}

	// Check if directory has content
	entries, err := os.ReadDir(m.sourceDir)
	if err != nil {
		return false
	}

	return len(entries) > 0
}

// isTargetEmpty checks if target directory is empty
func (m *DataMigrator) isTargetEmpty() bool {
	entries, err := os.ReadDir(m.targetDir)
	if os.IsNotExist(err) {
		return true
	}
	if err != nil {
		return false
	}

	return len(entries) == 0
}

// removeSourceDirectory safely removes the source directory after verification
func (m *DataMigrator) removeSourceDirectory() error {
	lg, _ := logger.Get()

	// Safety checks before removal
	if m.sourceDir == "" || m.sourceDir == "/" || m.sourceDir == "/var" || m.sourceDir == "/usr" {
		return fmt.Errorf("refusing to remove system directory: %s", m.sourceDir)
	}

	// Additional check: ensure target directory has the migrated data
	if !m.hasDataToMigrate() {
		lg.Info("Source directory is already empty or doesn't exist", logger.String("source", m.sourceDir))
		return nil
	}

	// Verify target has data before removing source
	if m.isTargetEmpty() {
		return fmt.Errorf("target directory is empty, refusing to remove source directory")
	}

	terminal.PrintInfo(fmt.Sprintf("Removing old data directory: %s", m.sourceDir))

	// Remove the source directory
	if err := os.RemoveAll(m.sourceDir); err != nil {
		return fmt.Errorf("failed to remove source directory %s: %w", m.sourceDir, err)
	}

	lg.Info("Source directory removed successfully after migration",
		logger.String("removed_directory", m.sourceDir),
		logger.String("target_directory", m.targetDir))

	terminal.PrintSuccess(fmt.Sprintf("Old data directory removed: %s", m.sourceDir))
	return nil
}

// setTargetOwnership sets mysql:mysql ownership on target directory
func (m *DataMigrator) setTargetOwnership() error {
	lg, _ := logger.Get()

	cmd := exec.Command("chown", "-R", "mysql:mysql", m.targetDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set ownership on %s: %w", m.targetDir, err)
	}

	lg.Info("Ownership set after migration",
		logger.String("directory", m.targetDir),
		logger.String("owner", "mysql:mysql"))

	return nil
}

// MigrateBinlogData migrates binlog data and updates index file with correct paths
func (m *DataMigrator) MigrateBinlogData() error {
	lg, _ := logger.Get()

	// First do regular migration
	if err := m.MigrateData(); err != nil {
		return err
	}

	// Then fix the binlog index file
	if err := m.fixBinlogIndexFile(); err != nil {
		lg.Warn("Failed to fix binlog index file",
			logger.String("target", m.targetDir),
			logger.Error(err))
		terminal.PrintWarning("Warning: Failed to fix binlog index file")
	}

	return nil
}

// fixBinlogIndexFile updates mysql-bin.index file to use correct paths
func (m *DataMigrator) fixBinlogIndexFile() error {
	lg, _ := logger.Get()

	indexFile := filepath.Join(m.targetDir, "mysql-bin.index")

	// Check if index file exists
	if _, err := os.Stat(indexFile); err != nil {
		lg.Info("No binlog index file found, skipping fix",
			logger.String("index_file", indexFile))
		return nil
	}

	// Read current index file
	content, err := os.ReadFile(indexFile)
	if err != nil {
		return fmt.Errorf("failed to read binlog index file: %w", err)
	}

	// Update paths in index file
	lines := strings.Split(string(content), "\n")
	updatedLines := make([]string, 0, len(lines))
	pathsUpdated := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Extract just the filename from the path
		filename := filepath.Base(line)

		// Create new path with target directory
		newPath := filepath.Join(m.targetDir, filename)
		updatedLines = append(updatedLines, newPath)

		if line != newPath {
			pathsUpdated++
		}
	}

	// Only rewrite if we made changes
	if pathsUpdated > 0 {
		// Write updated content
		newContent := strings.Join(updatedLines, "\n")
		if len(updatedLines) > 0 {
			newContent += "\n" // Ensure trailing newline
		}

		if err := os.WriteFile(indexFile, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to write updated binlog index file: %w", err)
		}

		// Set ownership
		cmd := exec.Command("chown", "mysql:mysql", indexFile)
		if err := cmd.Run(); err != nil {
			lg.Warn("Failed to set ownership on binlog index file", logger.Error(err))
		}

		lg.Info("Binlog index file updated",
			logger.String("index_file", indexFile),
			logger.Int("paths_updated", pathsUpdated))
		terminal.PrintSuccess(fmt.Sprintf("Updated %d binlog paths in index file", pathsUpdated))
	} else {
		lg.Info("Binlog index file already has correct paths")
	}

	return nil
}
