package file

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sfDBTools/internal/logger"
	"syscall"
)

// JSONWriter handles JSON file operations
type JSONWriter struct{}

// NewJSONWriter creates a new JSONWriter instance
func NewJSONWriter() *JSONWriter {
	return &JSONWriter{}
}

// WriteToFile writes data as JSON to the specified file path
func (jw *JSONWriter) WriteToFile(filePath string, data interface{}) error {
	// Ensure parent directory exists using manager (afero)
	m := NewManager()
	if err := m.EnsureParentDir(filePath); err != nil {
		return fmt.Errorf("failed to ensure parent directory: %w", err)
	}

	file, err := m.fs.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
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

// EnsureParentDir ensures the parent directory of filePath exists
func (m *Manager) EnsureParentDir(filePath string) error {
	parent := filepath.Dir(filePath)
	if parent == "" || parent == "." {
		return nil
	}
	if err := m.fs.MkdirAll(parent, 0o755); err != nil {
		m.logger.Error("failed to ensure parent directory",
			logger.String("path", parent),
			logger.Error(err))
		return fmt.Errorf("failed to ensure directory '%s': %w", parent, err)
	}
	return nil
}

// CopyFile copies a regular file from src to dst using the manager filesystem,
// then attempts to set permissions and ownership when possible.
func (m *Manager) CopyFile(src, dst string, info os.FileInfo) error {
	if err := m.EnsureParentDir(dst); err != nil {
		return err
	}

	srcFile, err := m.fs.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := m.fs.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer func() {
		if cerr := dstFile.Close(); cerr != nil {
			m.logger.Warn("Failed to close destination file", logger.String("file", dst))
		}
	}()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Try to set permissions via the underlying FS if supported
	if err := m.fs.Chmod(dst, info.Mode().Perm()); err != nil {
		// Not fatal; log and continue
		m.logger.Warn("Failed to set permissions on destination file", logger.String("file", dst), logger.Error(err))
	}

	// Try to preserve ownership from stat info
	if statT, ok := info.Sys().(*syscall.Stat_t); ok {
		_ = os.Chown(dst, int(statT.Uid), int(statT.Gid))
	}

	return nil
}
