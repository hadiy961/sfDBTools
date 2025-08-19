package backup_utils

import (
	"os"
	"path/filepath"
	"time"
)

// cleanupOldBackups removes dated subdirectories older than retentionDays from outputDir.
// It returns a slice of removed directory names. Directories must be named in YYYY_MM_DD format to be considered.
func CleanupOldBackups(outputDir string, retentionDays int) ([]string, error) {
	if retentionDays <= 0 {
		return nil, nil
	}

	threshold := time.Now().AddDate(0, 0, -retentionDays)

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, err
	}

	var removed []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		date, err := time.Parse("2006_01_02", entry.Name())
		if err != nil {
			// not a dated directory
			continue
		}

		if date.Before(threshold) {
			os.RemoveAll(filepath.Join(outputDir, entry.Name()))
			removed = append(removed, entry.Name())
		}
	}

	return removed, nil
}
