// interfaces.go
package file

import (
	"fmt"
	"sfDBTools/utils/fs"
)

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
	DirectoryManager fs.DirectoryManager
}

// NewFileManager creates a new FileManager with default implementations
func NewFileManager() *FileManager {
	fm := &FileManager{
		JSONWriter: NewJSONWriter(),
	}
	return fm
}

// WriteJSONToDirectory writes JSON data to a file in a validated directory
func (fm *FileManager) WriteJSONToDirectory(dirPath, fileName string, data interface{}) error {
	// Validate directory first (allow caller to set DirectoryManager)
	if fm.DirectoryManager == nil {
		return fmt.Errorf("DirectoryManager is not set on FileManager")
	}
	if err := fm.DirectoryManager.Validate(dirPath); err != nil {
		return err
	}

	// Write JSON file
	filePath := dirPath + "/" + fileName
	return fm.JSONWriter.WriteToFile(filePath, data)
}
