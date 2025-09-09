package file

import (
	"encoding/json"
	"fmt"
	"os"
	"sfDBTools/internal/logger"
)

// WriteJSON writes a map or struct as a JSON file
func WriteJSON(filePath string, data interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	return nil
}

// CreateOutputDirectory creates the output directory if it doesn't exist
func CreateOutputDirectory(outputDir string) error {
	lg, _ := logger.Get()

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		lg.Error("Failed to create output directory",
			logger.String("dir", outputDir),
			logger.Error(err))
		return fmt.Errorf("failed to create output directory '%s': %w", outputDir, err)
	}

	lg.Debug("Output directory created/verified", logger.String("dir", outputDir))
	return nil
}

// ValidateOutputDir memastikan outputDir ada dan bisa ditulis.
// Jika belum ada, akan mencoba membuatnya.
func ValidateOutputDir(outputDir string) error {
	lg, _ := logger.Get()

	if outputDir == "" {
		lg.Error("Output directory is required")
		return fmt.Errorf("output directory is required")
	}
	info, err := os.Stat(outputDir)
	if os.IsNotExist(err) {
		lg.Debug("Output directory does not exist, attempting to create", logger.String("dir", outputDir))
		// Coba buat direktori
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			lg.Error("Failed to create output directory",
				logger.String("dir", outputDir),
				logger.Error(err))
			return fmt.Errorf("failed to create output directory '%s': %w", outputDir, err)
		}
		lg.Info("Output directory created", logger.String("dir", outputDir))
		return nil
	}
	if err != nil {
		lg.Error("Failed to access output directory",
			logger.String("dir", outputDir),
			logger.Error(err))
		return fmt.Errorf("failed to access output directory '%s': %w", outputDir, err)
	}
	if !info.IsDir() {
		lg.Error("Output path is not a directory", logger.String("dir", outputDir))
		return fmt.Errorf("output path '%s' is not a directory", outputDir)
	}
	// Cek permission tulis
	testFile := outputDir + "/.sfbackup_test"
	lg.Debug("Checking write permission for output directory", logger.String("dir", outputDir))
	f, err := os.Create(testFile)
	if err != nil {
		lg.Error("Output directory is not writable",
			logger.String("dir", outputDir),
			logger.Error(err))
		return fmt.Errorf("output directory '%s' is not writable: %w", outputDir, err)
	}
	f.Close()
	os.Remove(testFile)
	lg.Debug("Output directory is writable", logger.String("dir", outputDir))
	return nil
}
