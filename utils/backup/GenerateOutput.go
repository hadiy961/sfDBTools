package backup_utils

import (
	"fmt"
	"path/filepath"
	"sfDBTools/utils/compression"
	"time"
)

// GenerateOutputPaths generates the output file path and metadata file path
func GenerateOutputPaths(options BackupOptions) (string, string) {
	timestamp := time.Now().Format("2006_01_02")

	// Create subdirectory for the database
	dbDir := filepath.Join(options.OutputDir, timestamp, options.DBName)

	// Generate base filename
	baseFilename := fmt.Sprintf("%s_%s", options.DBName, timestamp)

	// Add appropriate extension based on compression
	var extension string
	if options.Compress {
		// Validate compression type and get extension
		compressionType, err := compression.ValidateCompressionType(options.Compression)
		if err != nil {
			// Default to gzip if invalid
			compressionType = compression.CompressionGzip
		}
		extension = ".sql" + compression.GetFileExtension(compressionType)
	} else {
		extension = ".sql"
	}

	// Add .enc extension if encryption is enabled
	if options.Encrypt {
		extension = extension + ".enc"
	}

	outputFile := filepath.Join(dbDir, baseFilename+extension)
	metaFile := filepath.Join(dbDir, baseFilename+".json")

	return outputFile, metaFile
}
