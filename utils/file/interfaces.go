// interfaces.go
package file

// Writer defines the interface for writing data to files
type Writer interface {
	WriteToFile(filePath string, data interface{}) error
}

// JSONWriterInterface defines the interface for JSON writing operations
type JSONWriterInterface interface {
	Writer
}

// DirectoryManagerInterface defines the interface for directory operations
type DirectoryManagerInterface interface {
	Create(dir string) error
	Validate(dir string) error
}

// FileManager combines all file and directory operations
type FileManager struct {
	JSONWriter       JSONWriterInterface
	DirectoryManager DirectoryManagerInterface
}

// NewFileManager creates a new FileManager with default implementations
func NewFileManager() *FileManager {
	return &FileManager{
		JSONWriter:       NewJSONWriter(),
		DirectoryManager: NewDirectoryManager(),
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
