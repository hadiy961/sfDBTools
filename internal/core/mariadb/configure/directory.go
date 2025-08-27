package configure

import (
	"fmt"
	"os"
	"os/exec"

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
	sourceDir string
	targetDir string
}

// NewDataMigrator creates a new data migrator
func NewDataMigrator(sourceDir, targetDir string) *DataMigrator {
	return &DataMigrator{
		sourceDir: sourceDir,
		targetDir: targetDir,
	}
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
