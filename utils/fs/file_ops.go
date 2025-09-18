package fs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"sfDBTools/internal/logger"

	"github.com/spf13/afero"
)

// fileOperations mengimplementasikan FileOperations interface
type fileOperations struct {
	fs     afero.Fs
	logger *logger.Logger
}

// newFileOperations membuat instance file operations baru
func newFileOperations(fs afero.Fs, logger *logger.Logger) FileOperations {
	return &fileOperations{
		fs:     fs,
		logger: logger,
	}
}

// Copy menyalin file dari src ke dst dengan permission default
func (f *fileOperations) Copy(src, dst string) error {
	info, err := f.fs.Stat(src)
	if err != nil {
		return fmt.Errorf("gagal stat file src %s: %w", src, err)
	}
	return f.CopyWithInfo(src, dst, info)
}

// CopyWithInfo menyalin file dengan informasi FileInfo yang diberikan
func (f *fileOperations) CopyWithInfo(src, dst string, info os.FileInfo) error {
	if err := f.EnsureDir(filepath.Dir(dst)); err != nil {
		return fmt.Errorf("gagal pastikan parent dir: %w", err)
	}

	srcFile, err := f.fs.Open(src)
	if err != nil {
		return fmt.Errorf("gagal buka file src: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := f.fs.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return fmt.Errorf("gagal buat file dst: %w", err)
	}
	defer func() {
		if cerr := dstFile.Close(); cerr != nil {
			f.logger.Warn("Gagal tutup file dst", logger.String("file", dst), logger.Error(cerr))
		}
	}()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("gagal copy konten file: %w", err)
	}

	// Set permission via filesystem jika didukung
	if err := f.fs.Chmod(dst, info.Mode().Perm()); err != nil {
		f.logger.Warn("Gagal set permission file", logger.String("file", dst), logger.Error(err))
	}

	// Coba preserve ownership menggunakan syscall
	if statT, ok := info.Sys().(*syscall.Stat_t); ok {
		_ = os.Chown(dst, int(statT.Uid), int(statT.Gid))
	}

	return nil
}

// Move memindahkan file dari src ke dst
func (f *fileOperations) Move(src, dst string) error {
	if err := f.Copy(src, dst); err != nil {
		return err
	}
	return f.fs.Remove(src)
}

// EnsureDir memastikan direktori ada, buat jika tidak ada
func (f *fileOperations) EnsureDir(path string) error {
	if path == "" || path == "." {
		return nil
	}

	normalizedPath := filepath.Clean(path)
	return f.fs.MkdirAll(normalizedPath, 0755)
}

// WriteJSON menulis data sebagai JSON ke file
func (f *fileOperations) WriteJSON(path string, data interface{}) error {
	if err := f.EnsureDir(filepath.Dir(path)); err != nil {
		return fmt.Errorf("gagal pastikan parent dir: %w", err)
	}

	file, err := f.fs.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("gagal buat file JSON: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("gagal encode JSON: %w", err)
	}

	return nil
}
