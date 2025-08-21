package restore_utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// BackupFileInfo represents information about a backup file
type BackupFileInfo struct {
	Path         string
	Name         string
	Size         int64
	ModTime      time.Time
	DatabaseName string
}

// FindBackupFiles finds all backup files in the specified directory and subdirectories
func FindBackupFiles(dir string) ([]BackupFileInfo, error) {
	var files []BackupFileInfo

	// Common backup file extensions - including combined extensions
	backupExtensions := []string{
		".sql.gz.enc",  // Compressed and encrypted SQL
		".sql.zst.enc", // Zstandard compressed and encrypted SQL
		".sql.enc",     // Encrypted SQL
		".sql.gz",      // Compressed SQL
		".sql.zst",     // Zstandard compressed SQL
		".dump.gz",     // Compressed dump
		".dump.enc",    // Encrypted dump
		".backup.gz",   // Compressed backup
		".backup.enc",  // Encrypted backup
		".sql",         // Plain SQL
		".dump",        // Plain dump
		".backup",      // Plain backup
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking even if we can't access a file
		}

		if info.IsDir() {
			return nil
		}

		// Check if file has backup extension
		isBackupFile := false
		fileName := strings.ToLower(info.Name())

		for _, ext := range backupExtensions {
			if strings.HasSuffix(fileName, ext) {
				isBackupFile = true
				break
			}
		}

		if !isBackupFile {
			return nil
		}

		// Extract database name from filename (common patterns)
		dbName := extractDatabaseNameFromFilename(info.Name())

		files = append(files, BackupFileInfo{
			Path:         path,
			Name:         info.Name(),
			Size:         info.Size(),
			ModTime:      info.ModTime(),
			DatabaseName: dbName,
		})

		return nil
	})

	return files, err
}

// extractDatabaseNameFromFilename tries to extract database name from filename
func extractDatabaseNameFromFilename(filename string) string {
	// Remove common backup extensions - handle combined extensions first
	name := filename
	extensions := []string{
		".sql.gz.enc",    // Compressed and encrypted SQL
		".sql.zst.enc",   // Zstandard compressed and encrypted SQL
		".dump.gz.enc",   // Compressed and encrypted dump
		".backup.gz.enc", // Compressed and encrypted backup
		".sql.gz",        // Compressed SQL
		".sql.zst",       // Zstandard compressed SQL
		".sql.enc",       // Encrypted SQL
		".dump.gz",       // Compressed dump
		".dump.enc",      // Encrypted dump
		".backup.gz",     // Compressed backup
		".backup.enc",    // Encrypted backup
		".sql",           // Plain SQL
		".dump",          // Plain dump
		".backup",        // Plain backup
	}

	// Remove extensions in order (longest first to handle combined extensions)
	for _, ext := range extensions {
		if strings.HasSuffix(strings.ToLower(name), ext) {
			name = name[:len(name)-len(ext)]
			break
		}
	}

	// Try to extract database name from common patterns:
	// pattern1: gmit_pelita_2025_08_04.sql.gz.enc -> gmit_pelita
	// pattern2: backup_dbname_20250803.sql -> dbname
	// pattern3: dbname.sql -> dbname
	parts := strings.Split(name, "_")
	if len(parts) >= 2 {
		var dbNameParts []string

		// Collect all parts that are not date-like
		for _, part := range parts {
			if len(part) > 0 && !isDateLike(part) {
				dbNameParts = append(dbNameParts, part)
			}
		}

		// If we found non-date parts, join them with underscore
		if len(dbNameParts) > 0 {
			return strings.Join(dbNameParts, "_")
		}
	}

	// If no pattern matched, return the filename without extension
	return name
}

// isDateLike checks if a string looks like a date (contains only digits)
func isDateLike(s string) bool {
	// Check if string contains only digits
	if len(s) == 0 {
		return false
	}

	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}

	// Common date patterns:
	// - 4 digits: year (2025)
	// - 2 digits: month/day (08, 04)
	// - 8 digits: date format like 20250804
	return len(s) == 2 || len(s) == 4 || len(s) == 8
}

// SelectBackupFileInteractive shows available backup files and lets user choose one
func SelectBackupFileInteractive(baseDir string) (string, error) {
	backupDirs := []string{baseDir, "./backup", "./backups", "./data/backup"}
	var allFiles []BackupFileInfo
	seenFiles := make(map[string]bool) // Track files by their absolute path to avoid duplicates

	// Search in multiple possible backup directories
	for _, dir := range backupDirs {
		if _, err := os.Stat(dir); err == nil {
			files, err := FindBackupFiles(dir)
			if err == nil {
				for _, file := range files {
					// Get absolute path to check for duplicates
					absPath, err := filepath.Abs(file.Path)
					if err != nil {
						absPath = file.Path // fallback to relative path
					}

					// Only add if we haven't seen this file before
					if !seenFiles[absPath] {
						seenFiles[absPath] = true
						allFiles = append(allFiles, file)
					}
				}
			}
		}
	}

	if len(allFiles) == 0 {
		fmt.Println("❌ No backup files found.")
		fmt.Printf("   Searched in directories: %s\n", strings.Join(backupDirs, ", "))
		fmt.Println("   Use --file flag to specify backup file path manually.")
		return "", fmt.Errorf("no backup files found")
	}

	// Sort files by modification time (newest first)
	for i := 0; i < len(allFiles)-1; i++ {
		for j := i + 1; j < len(allFiles); j++ {
			if allFiles[i].ModTime.Before(allFiles[j].ModTime) {
				allFiles[i], allFiles[j] = allFiles[j], allFiles[i]
			}
		}
	}

	// Display available files
	fmt.Println("📁 Available Backup Files:")
	fmt.Println("===========================")
	for i, file := range allFiles {
		relPath, err := filepath.Rel(".", file.Path)
		if err != nil {
			relPath = file.Path // fallback to absolute path if relative fails
		}
		sizeStr := formatFileSize(file.Size)
		timeStr := file.ModTime.Format("2006-01-02 15:04")

		fmt.Printf("   %d. %s\n", i+1, file.Name)
		fmt.Printf("      Path: %s\n", relPath)
		fmt.Printf("      Database: %s | Size: %s | Date: %s\n",
			file.DatabaseName, sizeStr, timeStr)
		fmt.Println()
	}

	// Let user choose
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nSelect backup file (1-%d): ", len(allFiles))
	choice, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read selection: %w", err)
	}

	choice = strings.TrimSpace(choice)
	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(allFiles) {
		return "", fmt.Errorf("invalid selection: %s", choice)
	}

	return allFiles[index-1].Path, nil
}

// SelectGrantsFileInteractive shows available grants files and lets user choose one
func SelectGrantsFileInteractive(baseDir string) (string, error) {
	fmt.Println("🔐 Searching for grants backup files...")

	// Find all grants files specifically
	allFiles, err := FindGrantsFiles(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to find grants files: %w", err)
	}

	if len(allFiles) == 0 {
		fmt.Println("❌ No grants backup files found.")
		fmt.Printf("   Searched in: %s (looking for 'grants' and 'user_grants' subdirectories)\n", baseDir)
		fmt.Println("   Use --file flag to specify grants file path manually.")
		return "", fmt.Errorf("no grants backup files found in %s", baseDir)
	}

	fmt.Printf("Found %d grants backup files\n", len(allFiles))

	// Display available files
	fmt.Println("📁 Available Grants Backup Files:")
	fmt.Println("==================================")
	for i, file := range allFiles {
		relPath, err := filepath.Rel(".", file.Path)
		if err != nil {
			relPath = file.Path // fallback to absolute path if relative fails
		}
		sizeStr := formatFileSize(file.Size)
		timeStr := file.ModTime.Format("2006-01-02 15:04")

		fmt.Printf("   %d. %s\n", i+1, file.Name)
		fmt.Printf("      Path: %s\n", relPath)
		fmt.Printf("      Type: %s | Size: %s | Date: %s\n",
			file.DatabaseName, sizeStr, timeStr)
		fmt.Println()
	}

	// Let user choose
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nSelect grants backup file (1-%d): ", len(allFiles))
	choice, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read selection: %w", err)
	}

	choice = strings.TrimSpace(choice)
	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(allFiles) {
		return "", fmt.Errorf("invalid selection: %s", choice)
	}

	return allFiles[index-1].Path, nil
}

// FindGrantsFiles finds all grants backup files in the specified directory and subdirectories
func FindGrantsFiles(dir string) ([]BackupFileInfo, error) {
	var files []BackupFileInfo

	// Common backup file extensions - including combined extensions
	grantsExtensions := []string{
		".sql.gz.enc",  // Compressed and encrypted SQL grants
		".sql.zst.enc", // Zstandard compressed and encrypted SQL grants
		".sql.enc",     // Encrypted SQL grants
		".sql.gz",      // Compressed SQL grants
		".sql.zst",     // Zstandard compressed SQL grants
		".sql",         // Plain SQL grants
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking even if we can't access a file
		}

		if info.IsDir() {
			return nil
		}

		// Only look in grants directories and check if file has grants-related name
		dirName := filepath.Base(filepath.Dir(path))
		if dirName != "grants" && dirName != "user_grants" {
			return nil
		}

		// Check if file has backup extension
		isGrantsFile := false
		fileName := strings.ToLower(info.Name())

		for _, ext := range grantsExtensions {
			if strings.HasSuffix(fileName, ext) {
				isGrantsFile = true
				break
			}
		}

		if !isGrantsFile {
			return nil
		}

		// Determine grants type from filename
		grantsType := extractGrantsTypeFromFilename(info.Name())

		files = append(files, BackupFileInfo{
			Path:         path,
			Name:         info.Name(),
			Size:         info.Size(),
			ModTime:      info.ModTime(),
			DatabaseName: grantsType,
		})

		return nil
	})

	return files, err
}

// extractGrantsTypeFromFilename tries to extract grants type from filename
func extractGrantsTypeFromFilename(filename string) string {
	// Extract type from grants filename patterns:
	// system_users_2025_08_11_140853.sql.gz.enc -> System Users Grants
	// dbname_grants_2025_08_11_140853.sql.gz.enc -> Database Grants (dbname)

	if strings.Contains(filename, "system_users") {
		return "System Users Grants"
	}

	if strings.Contains(filename, "_grants_") {
		// Extract database name from pattern: dbname_grants_timestamp
		parts := strings.Split(filename, "_grants_")
		if len(parts) > 0 {
			dbName := parts[0]
			return fmt.Sprintf("Database Grants (%s)", dbName)
		}
	}

	return "User Grants"
}

// formatFileSize formats file size in human readable format
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// ValidateBackupFile checks if the backup file exists and is readable
func ValidateBackupFile(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("backup file path cannot be empty")
	}

	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", filePath)
	}
	if err != nil {
		return fmt.Errorf("cannot access backup file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	if info.Size() == 0 {
		return fmt.Errorf("backup file is empty: %s", filePath)
	}

	// Check if file is readable
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot read backup file: %w", err)
	}
	file.Close()

	return nil
}
