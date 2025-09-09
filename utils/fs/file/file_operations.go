package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// JSONWriter handles JSON file operations
type JSONWriter struct{}

// NewJSONWriter creates a new JSONWriter instance
func NewJSONWriter() *JSONWriter {
	return &JSONWriter{}
}

// WriteToFile writes data as JSON to the specified file path
func (jw *JSONWriter) WriteToFile(filePath string, data interface{}) error {
	// Ensure parent directory exists
	parent := filepath.Dir(filePath)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("failed to ensure directory '%s': %w", parent, err)
	}

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

// WriteJSON writes a map or struct as a JSON file (backward compatibility)
func WriteJSON(filePath string, data interface{}) error {
	writer := NewJSONWriter()
	return writer.WriteToFile(filePath, data)
}
