package dbconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"sfDBTools/internal/config"
	"sfDBTools/utils/terminal"
)

// FileManager handles dbconfig file operations
type FileManager struct {
	configDir string
}

// NewFileManager creates a new FileManager instance
func NewFileManager() *FileManager {
	configDir, err := config.GetDatabaseConfigDirectory()
	if err != nil {
		// Fallback to a default path
		configDir = "/opt/sfDBTools/config"
	}

	return &FileManager{
		configDir: configDir,
	}
}

// ListConfigFiles returns all database configuration files
func (fm *FileManager) ListConfigFiles() ([]*FileInfo, error) {
	configPath := fm.configDir
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return []*FileInfo{}, nil
	}

	entries, err := os.ReadDir(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config directory: %v", err)
	}

	var files []*FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".cnf.enc") {
			continue
		}

		fullPath := filepath.Join(configPath, filename)
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileInfo := &FileInfo{
			Name:    strings.TrimSuffix(filename, ".cnf.enc"),
			Path:    fullPath,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsValid: fm.isValidConfigFile(fullPath),
		}

		files = append(files, fileInfo)
	}

	// Sort by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	return files, nil
}

// FindConfigFile finds a config file by name
func (fm *FileManager) FindConfigFile(name string) (*FileInfo, error) {
	files, err := fm.ListConfigFiles()
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.Name == name {
			return file, nil
		}
	}

	return nil, fmt.Errorf("configuration '%s' not found", name)
}

// DeleteConfigFile deletes a configuration file
func (fm *FileManager) DeleteConfigFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	return os.Remove(filePath)
}

// DeleteMultipleFiles deletes multiple configuration files
func (fm *FileManager) DeleteMultipleFiles(filePaths []string, showProgress bool) *DeleteResult {
	result := &DeleteResult{
		DeletedFiles: []string{},
		Errors:       []string{},
	}

	if showProgress {
		terminal.PrintInfo(fmt.Sprintf("Deleting %d configuration files...", len(filePaths)))
	}

	for i, filePath := range filePaths {
		if showProgress && len(filePaths) > 1 {
			progress := float64(i+1) / float64(len(filePaths)) * 100
			terminal.PrintInfo(fmt.Sprintf("Progress: %.0f%% (%d/%d)", progress, i+1, len(filePaths)))
		}

		if err := fm.DeleteConfigFile(filePath); err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete %s: %v", filepath.Base(filePath), err))
		} else {
			result.DeletedCount++
			result.DeletedFiles = append(result.DeletedFiles, filepath.Base(filePath))
		}
	}

	return result
}

// GetConfigFilePath returns the full path for a config file
func (fm *FileManager) GetConfigFilePath(name string) string {
	if !strings.HasSuffix(name, ".cnf.enc") {
		name += ".cnf.enc"
	}
	return filepath.Join(fm.configDir, name)
}

// BackupConfigFile creates a backup of a configuration file
func (fm *FileManager) BackupConfigFile(filePath string) (string, error) {
	backupPath := filePath + ".backup." + time.Now().Format("20060102-150405")

	sourceFile, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %v", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %v", err)
	}
	defer destFile.Close()

	if _, err := sourceFile.WriteTo(destFile); err != nil {
		os.Remove(backupPath) // Clean up on error
		return "", fmt.Errorf("failed to copy file: %v", err)
	}

	return backupPath, nil
}

// RestoreBackup restores a backup file
func (fm *FileManager) RestoreBackup(backupPath, originalPath string) error {
	return os.Rename(backupPath, originalPath)
}

// CleanupBackups removes backup files older than specified days
func (fm *FileManager) CleanupBackups(days int) (int, error) {
	configPath := fm.configDir
	entries, err := os.ReadDir(configPath)
	if err != nil {
		return 0, fmt.Errorf("error reading config directory: %v", err)
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	cleanedCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.Contains(filename, ".backup.") {
			continue
		}

		fullPath := filepath.Join(configPath, filename)
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			if err := os.Remove(fullPath); err == nil {
				cleanedCount++
			}
		}
	}

	return cleanedCount, nil
}

// EnsureConfigDir ensures the config directory exists
func (fm *FileManager) EnsureConfigDir() error {
	return os.MkdirAll(fm.configDir, 0700)
}

// GetConfigDir returns the configuration directory path
func (fm *FileManager) GetConfigDir() string {
	return fm.configDir
}

// isValidConfigFile performs basic validation of config file
func (fm *FileManager) isValidConfigFile(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	// Check if file is not empty and has reasonable size
	if info.Size() == 0 || info.Size() > 10*1024*1024 {
		return false
	}

	// Check file extension
	if !strings.HasSuffix(filePath, ".cnf.enc") {
		return false
	}

	return true
}

// DisplayFileListSummary shows a summary of configuration files
func (fm *FileManager) DisplayFileListSummary(files []*FileInfo) {
	if len(files) == 0 {
		terminal.PrintWarning("No configuration files found.")
		return
	}

	terminal.PrintSubHeader(fmt.Sprintf("Found %d configuration files:", len(files)))

	data := [][]string{}
	for i, file := range files {
		status := "✅ Valid"
		if !file.IsValid {
			status = "❌ Invalid"
		}

		data = append(data, []string{
			fmt.Sprintf("%d", i+1),
			file.Name,
			file.GetFormattedSize(),
			file.ModTime.Format("2006-01-02 15:04"),
			status,
		})
	}

	headers := []string{"#", "Name", "Size", "Modified", "Status"}
	terminal.FormatTable(headers, data)
}
