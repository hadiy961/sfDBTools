// interfaces.go
package file

import "sfDBTools/utils/dir"

// Writer defines the interface for writing data to files
type Writer interface {
	WriteToFile(filePath string, data interface{}) error
}

// JSONWriterInterface defines the interface for JSON writing operations
type JSONWriterInterface interface {
	Writer
}

// FileManager combines all file and directory operations
type FileManager struct {
	JSONWriter       JSONWriterInterface
	DirectoryManager dir.Manager
}

// NewFileManager creates a new FileManager with default implementations
func NewFileManager() *FileManager {
	return &FileManager{
		JSONWriter: NewJSONWriter(),
	}
}

// WriteJSONToDirectory writes JSON data to a file in a validated directory
func (fm *FileManager) WriteJSONToDirectory(dirPath, fileName string, data interface{}) error {
	// Validate directory first
	if err := fm.DirectoryManager.Validate(dirPath); err != nil {
		return err
	}

	// Write JSON file
	filePath := dirPath + "/" + fileName
	return fm.JSONWriter.WriteToFile(filePath, data)
}
